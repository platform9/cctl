package machine

import (
	"fmt"

	log "github.com/platform9/ssh-provider/pkg/logrus"

	"github.com/ghodss/yaml"
	"github.com/platform9/ssh-provider/pkg/controller"
	"github.com/platform9/ssh-provider/pkg/machine"
	"github.com/platform9/ssh-provider/pkg/nodeadm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func (a *Actuator) createNode(cluster *clusterv1.Cluster, machine *clusterv1.Machine, machineClient machine.Client) error {
	machineSpec, err := controller.GetMachineSpec(*machine)
	if err != nil {
		return fmt.Errorf("unable to decode spec of machine %q: %v", machine.Name, err)
	}
	// Install correct version of nodeadm
	installNodeadm(machineSpec.ComponentVersions.NodeadmVersion, machineClient)
	clusterSpec, err := controller.GetClusterSpec(*cluster)
	if err != nil {
		return fmt.Errorf("unable to decode cluster spec: %v", err)
	}
	bootstrapTokenSecret, err := a.kubeClient.CoreV1().Secrets(cluster.Namespace).Get(clusterSpec.BootstrapTokenSecret.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get bootstrap token secret %q: %v", clusterSpec.BootstrapTokenSecret.Name, err)
	}
	if err := deployKubernetesNode(cluster, machine, machineClient, bootstrapTokenSecret); err != nil {
		return fmt.Errorf("unable to deploy kubernetes: %v", err)
	}
	return nil
}

func deployKubernetesNode(cluster *clusterv1.Cluster, machine *clusterv1.Machine, machineClient machine.Client, bootstrapTokenSecret *corev1.Secret) error {
	machineSpec, err := controller.GetMachineSpec(*machine)
	if err != nil {
		return fmt.Errorf("unable to decode spec of machine %q: %v", machine.Name, err)
	}
	// Install correct version of nodeadm
	installNodeadm(machineSpec.ComponentVersions.NodeadmVersion, machineClient)
	if len(cluster.Status.APIEndpoints) == 0 {
		return fmt.Errorf("no API endpoints found")
	}
	// The first API endpoint should work as well as any other
	apiEndpoint := cluster.Status.APIEndpoints[0]
	bootstrapToken, ok := bootstrapTokenSecret.Data["token"]
	if !ok {
		return fmt.Errorf("bootstrap token secret missing %q key", "token")
	}
	caHash, ok := bootstrapTokenSecret.Data["cahash"]
	if !ok {
		return fmt.Errorf("bootstrap token secret missing %q key", "cahash")
	}
	discoveryTokenAPIServers := []string{
		fmt.Sprintf("%s:%d", apiEndpoint.Host, apiEndpoint.Port),
	}
	discoveryTokenCAHashes := []string{
		string(caHash),
	}
	joinConfig, err := nodeadm.JoinConfigurationForMachine(cluster, machine, discoveryTokenAPIServers, discoveryTokenCAHashes, string(bootstrapToken))
	if err != nil {
		return fmt.Errorf("error creating nodeadm join configuration: %v", err)
	}
	joinConfigBytes, err := yaml.Marshal(joinConfig)
	if err != nil {
		return fmt.Errorf("error marshalling nodeadm join configuration to YAML: %v", err)
	}
	log.Println("writing nodeadm configuration")
	tmpNodeadmConfigPath := "/tmp/nodeadm.yaml"
	if err := machineClient.WriteFile(tmpNodeadmConfigPath, 0600, joinConfigBytes); err != nil {
		return fmt.Errorf("error writing nodeadm join configuration to %q: %v", NodeadmConfigPath, err)
	}
	if err := machineClient.MoveFile(tmpNodeadmConfigPath, NodeadmConfigPath); err != nil {
		return fmt.Errorf("error moving file from %q to %q:%v", tmpNodeadmConfigPath, NodeadmConfigPath, err)
	}
	cmd := fmt.Sprintf("%s join --cfg %s",
		NodeadmPath,
		NodeadmConfigPath)
	log.Println("deploying kubernetes. this might take a few minutes..")
	stdOut, stdErr, err := machineClient.RunCommand(cmd)
	if err != nil {
		log.Println(string(stdOut))
		log.Println(string(stdErr))
		return fmt.Errorf("error running %q: %v", cmd, err)
	}
	log.Println(string(stdOut))
	return nil
}

func (a *Actuator) deleteNode(machine *clusterv1.Machine, machineClient machine.Client) error {
	machineSpec, err := controller.GetMachineSpec(*machine)
	if err != nil {
		return fmt.Errorf("unable to decode spec of machine %q: %v", machine.Name, err)
	}
	// Install correct version of nodeadm
	if err := installNodeadm(machineSpec.ComponentVersions.NodeadmVersion, machineClient); err != nil {
		return fmt.Errorf("unable to install the correct version of nodeadm: %v", err)
	}
	log.Println("resetting kubernetes on node")
	if err := resetKubernetes(machineClient); err != nil {
		return fmt.Errorf("unable to reset kubernetes: %v", err)
	}
	return nil
}
