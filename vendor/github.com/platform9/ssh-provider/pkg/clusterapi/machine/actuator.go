/*
Copyright 2018 Platform 9 Systems, Inc.
*/

package machine

import (
	"fmt"

	"github.com/platform9/ssh-provider/pkg/controller"
	"github.com/platform9/ssh-provider/pkg/machine"

	"k8s.io/client-go/kubernetes"

	spclient "github.com/platform9/ssh-provider/pkg/client/clientset_generated/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clusterclient "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
	clusterutil "sigs.k8s.io/cluster-api/pkg/util"
)

const (
	EtcdadmPath       = "/opt/bin/etcdadm"
	NodeadmPath       = "/opt/bin/nodeadm"
	NodeadmConfigPath = "/etc/nodeadm.yaml"
)

type machineClientBuilder func(host string, port int, username string, privateKey string, publicKeys []string, insecureIgnoreHostKey bool) (machine.Client, error)

type Actuator struct {
	InsecureIgnoreHostKey bool

	kubeClient           kubernetes.Interface
	clusterClient        clusterclient.Interface
	spClient             spclient.Interface
	machineClientBuilder machineClientBuilder
}

func NewActuator(kubeClient kubernetes.Interface, clusterClient clusterclient.Interface, spClient spclient.Interface, machineClientBuilder machineClientBuilder, insecureIgnoreHostKey bool) *Actuator {
	return &Actuator{
		InsecureIgnoreHostKey: insecureIgnoreHostKey,

		kubeClient:           kubeClient,
		clusterClient:        clusterClient,
		spClient:             spClient,
		machineClientBuilder: machineClientBuilder,
	}
}

func (a *Actuator) Create(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	machineSpec, err := controller.GetMachineSpec(*machine)
	if err != nil {
		return fmt.Errorf("unable to decode spec of machine %q: %v", machine.Name, err)
	}
	pm, err := a.spClient.SshproviderV1alpha1().ProvisionedMachines(machine.Namespace).Get(machineSpec.ProvisionedMachineName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get provider machine %q bound to machine %q: %v", machineSpec.ProvisionedMachineName, machine.Name, err)
	}
	// TODO(dlipovetsky) validate machine-provisioned machine binding
	credentialSecret, err := a.kubeClient.CoreV1().Secrets(machine.Namespace).Get(pm.Spec.SSHConfig.CredentialSecret.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get ssh credential of provisioned machine %q bound to machine %q: %v", pm.Name, machine.Name, err)
	}
	username, privateKey, err := controller.UsernameAndKeyFromSecret(credentialSecret)
	if err != nil {
		return fmt.Errorf("unable to get ssh credential for machine %q: %v", machine.Name, err)
	}
	machineClient, err := a.machineClientBuilder(
		pm.Spec.SSHConfig.Host,
		pm.Spec.SSHConfig.Port,
		username,
		privateKey,
		pm.Spec.SSHConfig.PublicKeys,
		a.InsecureIgnoreHostKey,
	)
	if err != nil {
		return fmt.Errorf("error creating client for machine %q: %s", machine.Name, err)
	}
	if clusterutil.IsMaster(machine) {
		if err := a.createMaster(cluster, machine, pm, machineClient); err != nil {
			return fmt.Errorf("error creating machine %q: %s", machine.Name, err)
		}
	} else {
		if err := a.createNode(cluster, machine, machineClient); err != nil {
			return fmt.Errorf("error creating machine %q: %s", machine.Name, err)
		}
	}

	if _, err := a.clusterClient.ClusterV1alpha1().Machines(machine.Namespace).UpdateStatus(machine); err != nil {
		return fmt.Errorf("error updating machine %q: %v", machine.Name, err)
	}
	return nil
}

func (a *Actuator) Delete(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	machineSpec, err := controller.GetMachineSpec(*machine)
	if err != nil {
		return fmt.Errorf("unable to decode spec of machine %q: %v", machine.Name, err)
	}
	pm, err := a.spClient.SshproviderV1alpha1().ProvisionedMachines(machine.Namespace).Get(machineSpec.ProvisionedMachineName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get provider machine %q bound to machine %q: %v", machineSpec.ProvisionedMachineName, machine.Name, err)
	}
	// TODO(dlipovetsky) validate machine-provisioned machine binding
	credentialSecret, err := a.kubeClient.CoreV1().Secrets(machine.Namespace).Get(pm.Spec.SSHConfig.CredentialSecret.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get ssh credential of provisioned machine %q bound to machine %q: %v", pm.Name, machine.Name, err)
	}
	username, privateKey, err := controller.UsernameAndKeyFromSecret(credentialSecret)
	if err != nil {
		return fmt.Errorf("unable to get ssh credential for machine %q: %v", machine.Name, err)
	}
	machineClient, err := a.machineClientBuilder(
		pm.Spec.SSHConfig.Host,
		pm.Spec.SSHConfig.Port,
		username,
		privateKey,
		pm.Spec.SSHConfig.PublicKeys,
		a.InsecureIgnoreHostKey,
	)
	if err != nil {
		return fmt.Errorf("error creating client for machine %q: %s", machine.Name, err)
	}
	if clusterutil.IsMaster(machine) {
		if err := a.deleteMaster(machine, machineClient); err != nil {
			return fmt.Errorf("error deleting machine %q: %s", machine.Name, err)
		}
	} else {
		if err := a.deleteNode(machine, machineClient); err != nil {
			return fmt.Errorf("error deleting machine %q: %s", machine.Name, err)
		}
	}

	if _, err := a.clusterClient.ClusterV1alpha1().Machines(machine.Namespace).UpdateStatus(machine); err != nil {
		return fmt.Errorf("error updating machine %q: %v", machine.Name, err)
	}
	return nil
}

func (a *Actuator) Update(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	return nil
}

func (a *Actuator) Exists(cluster *clusterv1.Cluster, machine *clusterv1.Machine) (bool, error) {
	return false, nil
}
