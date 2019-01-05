package machine

import (
	"encoding/json"
	"fmt"

	log "github.com/platform9/ssh-provider/pkg/logrus"

	"github.com/ghodss/yaml"

	"github.com/platform9/ssh-provider/pkg/nodeadm"

	spv1 "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	"github.com/platform9/ssh-provider/pkg/controller"
	"github.com/platform9/ssh-provider/pkg/machine"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func (a *Actuator) createMaster(cluster *clusterv1.Cluster, machine *clusterv1.Machine, pm *spv1.ProvisionedMachine, machineClient machine.Client) error {
	machineSpec, err := controller.GetMachineSpec(*machine)
	if err != nil {
		return fmt.Errorf("unable to decode spec of machine %q: %v", machine.Name, err)
	}
	// Install correct version of nodeadm
	if err := installNodeadm(machineSpec.ComponentVersions.NodeadmVersion, machineClient); err != nil {
		return fmt.Errorf("unable to install the correct version of nodeadm: %v", err)
	}
	// Install correct version of etcdadm
	if err := installEtcdadm(machineSpec.ComponentVersions.EtcdadmVersion, machineClient); err != nil {
		return fmt.Errorf("unable to install the correct version of etcdadm: %v", err)
	}
	// Write secrets
	if err := a.writeMasterSecretsToMachine(cluster, machineClient); err != nil {
		return fmt.Errorf("unable to write secrets to machine: %s", err)
	}
	// Deploy etcd with etcdadm
	if err := deployEtcd(cluster, machine, machineClient); err != nil {
		return fmt.Errorf("unable to deploy etcd: %v", err)
	}
	// Deploy Kubernetes with nodeadm
	if err := deployKubernetesMaster(cluster, machine, pm, machineClient); err != nil {
		return fmt.Errorf("unable to deploy kubernetes: %v", err)
	}
	return nil
}

func (a *Actuator) deleteMaster(machine *clusterv1.Machine, machineClient machine.Client) error {
	log.Println("resetting kubernetes on node")
	machineSpec, err := controller.GetMachineSpec(*machine)
	if err != nil {
		return fmt.Errorf("unable to decode spec of machine %q: %v", machine.Name, err)
	}
	// Install correct version of nodeadm
	if err := installNodeadm(machineSpec.ComponentVersions.NodeadmVersion, machineClient); err != nil {
		return fmt.Errorf("unable to install the correct version of nodeadm: %v", err)
	}
	// Install correct version of etcdadm
	if err := installEtcdadm(machineSpec.ComponentVersions.EtcdadmVersion, machineClient); err != nil {
		return fmt.Errorf("unable to install the correct version of etcdadm: %v", err)
	}
	if err := resetKubernetes(machineClient); err != nil {
		return fmt.Errorf("unable to reset kubernetes: %v", err)
	}
	log.Println("resetting etcd on node")
	if err := resetEtcd(machineClient); err != nil {
		return fmt.Errorf("unable to reset etcd: %v", err)
	}
	return nil
}

func etcdMemberFromMachine(machine *clusterv1.Machine, machineClient machine.Client) (spv1.EtcdMember, error) {
	var etcdMember spv1.EtcdMember
	cmd := fmt.Sprintf("%s info", EtcdadmPath)
	stdOut, stdErr, err := machineClient.RunCommand(cmd)
	if err != nil {
		log.Println(string(stdOut))
		log.Println(string(stdErr))
		return etcdMember, fmt.Errorf("error running %q: %v", cmd, err)
	}
	log.Println(string(stdOut))
	err = json.Unmarshal(stdOut, &etcdMember)
	if err != nil {
		// TODO(dlipovetsky)
		return etcdMember, fmt.Errorf("error unmarshalling etcdadm info output: %v", err)
	}
	return etcdMember, nil
}

func deployEtcd(cluster *clusterv1.Cluster, machine *clusterv1.Machine, machineClient machine.Client) error {
	machineStatus, err := controller.GetMachineStatus(*machine)
	if err != nil {
		return fmt.Errorf("unable to decode machine status: %v", err)
	}
	clusterStatus, err := controller.GetClusterStatus(*cluster)
	if err != nil {
		return fmt.Errorf("unable to decode cluster status: %v", err)
	}

	// This condition is racy; the cluster status must be updated with existing
	// masters' etcd information. Two masters created concurrently may form two
	// individual etcd clusters. For now, assume masters are created
	// sequentially, and that this information is updated between any two
	// actuator.Create calls.
	formNewEtcdCluster := len(clusterStatus.EtcdMembers) == 0

	var cmd string
	if formNewEtcdCluster {
		cmd = fmt.Sprintf("%s init", EtcdadmPath)
	} else {
		// We assume that any unhealthy members have already been removed from
		// the cluster. Therefore, the first etcd member should work as well as
		// any.
		etcdMember := clusterStatus.EtcdMembers[0]
		if len(etcdMember.ClientURLs) == 0 {
			return fmt.Errorf("etcd member %q has no ClientURLs", etcdMember.Name)
		}
		etcdEndpoint := etcdMember.ClientURLs[0]
		cmd = fmt.Sprintf("%s join %s", EtcdadmPath, etcdEndpoint)
	}
	log.Printf("running %q command. This might take some time..", cmd)
	stdOut, stdErr, err := machineClient.RunCommand(cmd)
	if err != nil {
		log.Println(string(stdOut))
		log.Println(string(stdErr))
		return fmt.Errorf("error running %q: %v", cmd, err)
	}
	log.Println(string(stdOut))

	// Update cluster status with new etcd member
	newEtcdMember, err := etcdMemberFromMachine(machine, machineClient)
	if err != nil {
		return fmt.Errorf("error reading etcd member data from machine: %v", err)
	}
	machineStatus.EtcdMember = &newEtcdMember
	if err := controller.PutMachineStatus(*machineStatus, machine); err != nil {
		return fmt.Errorf("error updating status of machine: %v", err)
	}
	return nil
}

func resetEtcd(machineClient machine.Client) error {
	cmd := fmt.Sprintf("%s reset", EtcdadmPath)
	stdOut, stdErr, err := machineClient.RunCommand(cmd)
	if err != nil {
		log.Println(string(stdOut))
		log.Println(string(stdErr))
		return fmt.Errorf("error running %q: %v", cmd, err)
	}
	log.Println(string(stdOut))
	return nil
}

func deployKubernetesMaster(cluster *clusterv1.Cluster, machine *clusterv1.Machine, pm *spv1.ProvisionedMachine, machineClient machine.Client) error {
	initConfig, err := nodeadm.InitConfigurationForMachine(*cluster, *machine, *pm)
	if err != nil {
		return fmt.Errorf("error creating nodeadm init configuration: %v", err)
	}
	initConfigBytes, err := yaml.Marshal(initConfig)
	if err != nil {
		return fmt.Errorf("error marshalling nodeadm init configuration to YAML: %v", err)
	}
	log.Println("writing nodeadm configuration file")
	tmpNodeadmConfigPath := "/tmp/nodeadm.yaml"
	if err := machineClient.WriteFile(tmpNodeadmConfigPath, 0600, initConfigBytes); err != nil {
		return fmt.Errorf("error writing nodeadm init configuration to %q: %v", NodeadmConfigPath, err)
	}
	if err := machineClient.MoveFile(tmpNodeadmConfigPath, NodeadmConfigPath); err != nil {
		return fmt.Errorf("error moving file from %q to %q:%v", tmpNodeadmConfigPath, NodeadmConfigPath, err)
	}
	log.Println("deploying kubernetes. this might take a few minutes..")
	cmd := fmt.Sprintf("%s init --cfg %s", NodeadmPath, NodeadmConfigPath)
	stdOut, stdErr, err := machineClient.RunCommand(cmd)
	if err != nil {
		log.Println(string(stdOut))
		log.Println(string(stdErr))
		return fmt.Errorf("error running %q: %v", cmd, err)
	}
	log.Println(string(stdOut))
	return nil
}

func resetKubernetes(machineClient machine.Client) error {
	cmd := fmt.Sprintf("%s reset", NodeadmPath)
	stdOut, stdErr, err := machineClient.RunCommand(cmd)
	if err != nil {
		log.Println(string(stdOut))
		log.Println(string(stdErr))
		return fmt.Errorf("error running %q: %v", cmd, err)
	}
	log.Println(string(stdOut))
	return nil
}
