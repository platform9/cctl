package statefileutil

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/ghodss/yaml"
	"github.com/platform9/pf9-clusteradm/common"
	"github.com/platform9/ssh-provider/provisionedmachine"
	sshproviderv1 "github.com/platform9/ssh-provider/sshproviderconfig/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/util"
)

const (
	STATE_FILE_PATH = "/tmp/cluster-state.yaml"
)

func checkFileExists() (bool, error) {
	_, err := os.Stat(STATE_FILE_PATH)
	if err == nil {
		return true, nil
	}
	return false, err
}

func ReadStateFile() (common.ClusterState, error) {
	cs := new(common.ClusterState)
	ret, _ := checkFileExists()
	if ret == false {
		return *cs, nil
	}
	d, err := ioutil.ReadFile(STATE_FILE_PATH)
	if err != nil {
		return *cs, err
	}
	yaml.Unmarshal(d, cs)
	return *cs, nil
}

func WriteStateFile(cs *common.ClusterState) error {
	if cs != nil {
		bytes, _ := yaml.Marshal(cs)
		ioutil.WriteFile(STATE_FILE_PATH, bytes, 0600)
	}
	return nil
}

func GetMaster(cs *common.ClusterState) *provisionedmachine.ProvisionedMachine {
	for _, machine := range cs.Machines {
		if util.IsMaster(&machine) {
			return GetProvisionedMachine(cs, machine.Name)
		}
	}
	return nil
}

func GetMachine(cs *common.ClusterState, ip string) *clusterv1.Machine {
	for _, machine := range cs.Machines {
		if machine.Name == ip {
			return &machine
		}
	}
	return nil
}

func DeleteMachine(cs *common.ClusterState, ip string) {
	for i, machine := range cs.Machines {
		if machine.Name == ip {
			// Delete element without leaking memory.
			// See https://github.com/golang/go/wiki/SliceTricks
			copy(cs.Machines[i:], cs.Machines[i+1:])
			cs.Machines[len(cs.Machines)-1] = clusterv1.Machine{}
			cs.Machines = cs.Machines[:len(cs.Machines)-1]
			return
		}
	}
}

func DeleteEtcdMemberFromClusterState(cs *common.ClusterState, machine *clusterv1.Machine) {
	spv1Codec, err := sshproviderv1.NewCodec()
	if err != nil {
		log.Fatalf("Could not initialize codec for internal types: %v", err)
	}
	clusterProviderStatus := sshproviderv1.SSHClusterProviderStatus{}
	err = spv1Codec.DecodeFromProviderStatus(cs.Cluster.Status.ProviderStatus, &clusterProviderStatus)
	if err != nil {
		log.Fatalf("Could not decode cluster status from cluster state: %v", err)
	}

	machineProviderStatus := sshproviderv1.SSHMachineProviderStatus{}
	err = spv1Codec.DecodeFromProviderStatus(machine.Status.ProviderStatus, &machineProviderStatus)
	if err != nil {
		log.Fatalf("Could not decode machine status from cluster state: %v", err)
	}
	for i, member := range clusterProviderStatus.EtcdMembers {
		for _, memberClientURL := range member.ClientURLs {
			for _, machineClientURL := range machineProviderStatus.EtcdMember.ClientURLs {
				if memberClientURL == machineClientURL {
					log.Printf("Removing %s from etcd members in cluster state", machine.Name)
					// Delete element without leaking memory.
					// See https://github.com/golang/go/wiki/SliceTricks
					copy(clusterProviderStatus.EtcdMembers[i:], clusterProviderStatus.EtcdMembers[i+1:])
					clusterProviderStatus.EtcdMembers[len(clusterProviderStatus.EtcdMembers)-1] = sshproviderv1.EtcdMember{}
					clusterProviderStatus.EtcdMembers = clusterProviderStatus.EtcdMembers[:len(clusterProviderStatus.EtcdMembers)-1]
					encodedStatus, err := spv1Codec.EncodeToProviderStatus(&clusterProviderStatus)
					if err != nil {
						log.Fatal("Could not encode etcd member info to persist in cluster state")
					}
					cs.Cluster.Status.ProviderStatus = *encodedStatus
					return
				}
			}
		}
	}
}

func GetProvisionedMachine(cs *common.ClusterState, ip string) *provisionedmachine.ProvisionedMachine {
	for _, pm := range cs.ProvisionedMachines {
		if pm.SSHConfig.Host == ip {
			return &pm
		}
	}
	return nil
}

func DeleteProvisionedMachine(cs *common.ClusterState, ip string) {
	for i, pm := range cs.ProvisionedMachines {
		if pm.SSHConfig.Host == ip {
			// Delete element without leaking memory.
			// See https://github.com/golang/go/wiki/SliceTricks
			copy(cs.ProvisionedMachines[i:], cs.ProvisionedMachines[i+1:])
			cs.ProvisionedMachines[len(cs.ProvisionedMachines)-1] = provisionedmachine.ProvisionedMachine{}
			cs.ProvisionedMachines = cs.ProvisionedMachines[:len(cs.ProvisionedMachines)-1]
			return
		}
	}
}

// DeleteStateFile deletes the cluster state file
func DeleteStateFile() error {
	return os.Remove(STATE_FILE_PATH)
}
