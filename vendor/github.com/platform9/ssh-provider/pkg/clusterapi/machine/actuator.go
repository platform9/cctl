/*
Copyright 2018 Platform 9 Systems, Inc.
*/

package machine

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/coreos/go-semver/semver"
	log "github.com/platform9/ssh-provider/pkg/logrus"
	"github.com/sirupsen/logrus"

	"github.com/platform9/ssh-provider/pkg/controller"
	"github.com/platform9/ssh-provider/pkg/machine"
	semverutil "github.com/platform9/ssh-provider/pkg/util/semver"

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
	CachePath         = "/var/cache/ssh-provider"
)

type machineClientBuilder func(host string, port int, username string, privateKey string, publicKeys []string, insecureIgnoreHostKey bool) (machine.Client, error)

type Actuator struct {
	InsecureIgnoreHostKey bool

	kubeClient           kubernetes.Interface
	clusterClient        clusterclient.Interface
	spClient             spclient.Interface
	machineClientBuilder machineClientBuilder
	logLevel             logrus.Level
}

func NewActuator(kubeClient kubernetes.Interface, clusterClient clusterclient.Interface, spClient spclient.Interface, machineClientBuilder machineClientBuilder, insecureIgnoreHostKey bool, logLevel logrus.Level) *Actuator {
	log.SetLogLevel(logLevel)
	return &Actuator{
		InsecureIgnoreHostKey: insecureIgnoreHostKey,

		kubeClient:           kubeClient,
		clusterClient:        clusterClient,
		spClient:             spClient,
		machineClientBuilder: machineClientBuilder,
		logLevel:             logLevel,
	}
}

func trimVFromVersion(version string) string {
	return strings.TrimPrefix(version, "v")
}

//installEtcdadm installs etcdadm on the machine
func installEtcdadm(version string, machineClient machine.Client) error {
	return installComponent(EtcdadmPath, version, "etcdadm", machineClient)
}

//installNodeadm installs nodeadm on the machine
func installNodeadm(version string, machineClient machine.Client) error {
	return installComponent(NodeadmPath, version, "nodeadm", machineClient)
}

func installComponent(componentInstallPath, desiredVersion, componentName string, machineClient machine.Client) error {
	log.Printf("Installing %q.", componentName)

	log.Printf("Checking %q desired version.", componentName)
	parsedDesiredVersion, err := semver.NewVersion(trimVFromVersion(desiredVersion))
	if err != nil {
		return fmt.Errorf("unable to parse %q desired version %q: %v", componentName, desiredVersion, err)
	}

	exists, err := machineClient.Exists(componentInstallPath)
	if err != nil {
		return fmt.Errorf("unable to check if %q installed at %q: %v", componentName, componentInstallPath, err)
	}
	if exists {
		log.Printf("%q is already installed. Checking version.", componentName)
		installedVersionBytes, _, err := machineClient.RunCommand(fmt.Sprintf("%s version --short", componentInstallPath))
		if err != nil {
			return fmt.Errorf("unable to check %q installed version: %v", componentName, err)
		}
		installedVersion := trimVFromVersion(strings.TrimSpace(string(installedVersionBytes)))
		parsedInstalledVersion, err := semver.NewVersion(installedVersion)
		if err != nil {
			return fmt.Errorf("unable to parse %q installed version %q: %v", componentName, installedVersion, err)
		}
		log.Printf("Found %q version %q.", componentName, parsedInstalledVersion)
		if semverutil.EqualMajorMinorPatchVersions(*parsedDesiredVersion, *parsedInstalledVersion) {
			log.Printf("Using %q that is already installed. The installed and desired versions match on major.minor.patch.", componentName)
			return nil
		}
	}

	componentCachePath := filepath.Join(CachePath, componentName, desiredVersion, componentName)
	log.Printf("Checking for %q version %q in the cache %q.", componentName, desiredVersion, componentCachePath)
	exists, err = machineClient.Exists(componentCachePath)
	if err != nil {
		return fmt.Errorf("unable to check if %q exists: %v", componentCachePath, err)
	}
	if !exists {
		//TODO(puneet) Try download from a hosted location
		return fmt.Errorf("unable to find %q in the cache %q", componentName, componentCachePath)
	}
	log.Printf("Installing %q version %q from cache %q", componentName, desiredVersion, componentCachePath)
	if err := machineClient.CopyFile(componentCachePath, componentInstallPath); err != nil {
		return fmt.Errorf("unable to copy file from %q to %q: %v", componentCachePath, componentInstallPath, err)
	}
	return nil
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

func (a *Actuator) Update(cluster *clusterv1.Cluster, goalMachine *clusterv1.Machine) error {
	currentMachine, err := controller.GetMachineInstanceStatus(goalMachine)
	if err != nil {
		return fmt.Errorf("unable to get current machine from annotation %v", err)
	}
	requiresUpdate, err := requiresUpdate(goalMachine, currentMachine)
	if err != nil {
		return fmt.Errorf("unable to compare goal and current machines: %v", err)
	}
	if !requiresUpdate {
		log.Println("Current machine is already at required versions")
		return nil
	}
	if err := a.Delete(cluster, currentMachine); err != nil {
		return fmt.Errorf("unable to delete machine %v", err)
	}
	return a.Create(cluster, goalMachine)
}

func (a *Actuator) Exists(cluster *clusterv1.Cluster, machine *clusterv1.Machine) (bool, error) {
	return false, nil
}

// The two machines differ in a way that requires an update
func requiresUpdate(a *clusterv1.Machine, b *clusterv1.Machine) (bool, error) {
	// Do not want status changes. Do want changes that impact machine provisioning
	aSpec, err := controller.GetMachineSpec(*a)
	if err != nil {
		return false, fmt.Errorf("unable to decode machine spec: %v", err)
	}
	bSpec, err := controller.GetMachineSpec(*b)
	if err != nil {
		return false, fmt.Errorf("unable to decode machine spec: %v", err)
	}
	return (aSpec.ComponentVersions.CNIVersion != bSpec.ComponentVersions.CNIVersion ||
		aSpec.ComponentVersions.EtcdVersion != bSpec.ComponentVersions.EtcdVersion ||
		aSpec.ComponentVersions.FlannelVersion != bSpec.ComponentVersions.FlannelVersion ||
		aSpec.ComponentVersions.KeepalivedVersion != bSpec.ComponentVersions.KeepalivedVersion ||
		aSpec.ComponentVersions.KubernetesVersion != bSpec.ComponentVersions.KubernetesVersion), nil
}
