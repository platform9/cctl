/*
Copyright 2019 The cctl authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"

	setsutil "github.com/platform9/ssh-provider/pkg/util/sets"

	"github.com/platform9/cctl/common"
	log "github.com/platform9/cctl/pkg/logrus"
	"github.com/platform9/cctl/pkg/util/clusterapi"
	kubeadmutil "github.com/platform9/cctl/pkg/util/kubeadm"
	sshutil "github.com/platform9/cctl/pkg/util/ssh"

	spv1 "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	machineActuator "github.com/platform9/ssh-provider/pkg/clusterapi/machine"
	sputil "github.com/platform9/ssh-provider/pkg/controller"
	sshmachine "github.com/platform9/ssh-provider/pkg/machine"

	clustercommon "sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clusterutil "sigs.k8s.io/cluster-api/pkg/util"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	drainTimeout            time.Duration
	drainGracePeriodSeconds int
	drainDeleteLocalData    bool
	drainForce              bool
)

func updateBootstrapToken(masterMachine *clusterv1.Machine, masterProvisionedMachine *spv1.ProvisionedMachine) error {
	log.Println("Getting a bootstrap token from a master")
	newBootstrapTokenSecret, err := bootstrapTokenSecretFromMachine(masterMachine, masterProvisionedMachine)
	if err != nil {
		return fmt.Errorf("Unable to read bootstrap token from master: %v", err)
	}
	if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Get(common.DefaultBootstrapTokenSecretName, metav1.GetOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("Unable to get bootstrap token secret: %v", err)
		}
		if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newBootstrapTokenSecret); err != nil {
			return fmt.Errorf("Unable to create bootstrap token secret: %v", err)
		}
	} else {
		if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Update(newBootstrapTokenSecret); err != nil {
			return fmt.Errorf("Unable to update bootstrap token secret: %v", err)
		}
	}
	return nil
}

func createAdminKubeconfigSecret(machine *clusterv1.Machine, provisionedMachine *spv1.ProvisionedMachine) (*corev1.Secret, error) {
	adminConfigData, err := adminKubeconfigFromMachine(machine, provisionedMachine)
	if err != nil {
		return nil, fmt.Errorf("Unable to get admin kubeconfig data: %v", err)
	}
	adminConfigSecret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              common.DefaultAdminConfigSecretName,
			Namespace:         common.DefaultNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Data: make(map[string][]byte),
	}
	adminConfigSecret.Data[common.DefaultAdminConfigSecretKey] = adminConfigData
	return &adminConfigSecret, nil
}

func copyAdminConfigFromSecret(masterMachine *clusterv1.Machine, masterProvisionedMachine *spv1.ProvisionedMachine,
	newMachine *clusterv1.Machine, newProvisionedMachine *spv1.ProvisionedMachine) error {
	log.Println("Writing admin kubeconfig to machine")
	kubeconfig, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Get(common.DefaultAdminConfigSecretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Unable to get admin kubeconfig from secret: %v", err)
	}
	kubeconfigData, ok := kubeconfig.Data[common.DefaultAdminConfigSecretKey]
	if !ok {
		return fmt.Errorf("unable to find data in admin kubeconfig secret")
	}
	if len(kubeconfigData) == 0 {
		return fmt.Errorf("invalid data in admin kubeconfig secret")
	}
	if err := writeAdminKubeconfigToMachine(kubeconfigData, newMachine, newProvisionedMachine); err != nil {
		return fmt.Errorf("Unable to write admin kubeconfig to machine: %v", err)
	}
	return nil
}

func createAdminKubeConfigSecretIfNotPresent() error {
	machine, provisionedMachine, err := masterMachineAndProvisionedMachine()
	if err != nil {
		return fmt.Errorf("unable to get master machine and provisioned machine: %v", err)
	}
	if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Get(common.DefaultAdminConfigSecretName, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			adminKubeConfigSecret, err := createAdminKubeconfigSecret(machine, provisionedMachine)
			if err != nil {
				return fmt.Errorf("unable to create secret for admin kubeconfig: %v", err)
			}
			if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(adminKubeConfigSecret); err != nil {
				return fmt.Errorf("unable to create secret for admin kubeconfig: %v", err)
			}
		} else {
			return fmt.Errorf("unable to get secret for admin kubeconfig: %v", err)
		}
	}
	return nil
}

func createMachine(ip string, port int, iface string, roleString string, publicKeyFiles []string) {
	role := clustercommon.MachineRole(roleString)
	// TODO(dlipovetsky) Move to master validation code
	if role != clustercommon.MasterRole && role != clustercommon.NodeRole {
		log.Fatalf("Machine role %q is not supported, must be %q or %q.", role, clustercommon.MasterRole, clustercommon.NodeRole)
	}
	var publicKeys []string
	for _, file := range publicKeyFiles {
		publicKey, err := sshutil.PublicKeyFromFile(file)
		if err != nil {
			log.Fatalf("Unable to parse SSH public key from %q: %v", file, err)
		}
		publicKeys = append(publicKeys, string(ssh.MarshalAuthorizedKey(publicKey)))
	}

	cluster, err := state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Get(common.DefaultClusterName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Fatalf("No cluster found. Create a cluster before creating a machine.")
		}
		log.Fatalf("Unable to get cluster: %v", err)
	}

	cspec, err := sputil.GetClusterSpec(*cluster)
	if err != nil {
		log.Fatalf("Unable to decode cluster spec: %v", err)
	}
	// If no vip exists, check if other masters exist before creating a new one.
	if cspec.VIPConfiguration == nil {
		if role == clustercommon.MasterRole {
			_, _, err = masterMachineAndProvisionedMachine()
			if err == nil {
				log.Fatal("Creating a master is not allowed: this cluster already has one master and has no VIP configured.")
			}
		}
	}

	sshCredentialSecret, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Get(common.DefaultSSHCredentialSecretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Fatalf("No SSH credential found. Create a credential before creating a machine.")
		}
		log.Fatalf("Unable to get SSH credential secret: %v", err)
	}

	newSSHConfig := spv1.SSHConfig{
		Host:       ip,
		Port:       port,
		PublicKeys: publicKeys,
		CredentialSecret: corev1.LocalObjectReference{
			Name: sshCredentialSecret.Name,
		},
	}

	newProvisionedMachine, newMachine, err := newProvisionedMachineAndMachine(ip, role, iface, newSSHConfig)
	if _, err := state.SPClient.SshproviderV1alpha1().ProvisionedMachines(common.DefaultNamespace).Create(newProvisionedMachine); err != nil {
		log.Fatalf("Unable to create provisioned machine: %v", err)
	}
	if _, err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).Create(newMachine); err != nil {
		log.Fatalf("Unable to create machine: %v", err)
	}

	var masterMachine *clusterv1.Machine
	var masterProvisionedMachine *spv1.ProvisionedMachine
	if clusterutil.RoleContains(clustercommon.NodeRole, newMachine.Spec.Roles) {
		var err error
		masterMachine, masterProvisionedMachine, err = masterMachineAndProvisionedMachine()
		if err != nil {
			log.Fatalf("Unable to get a master machine and provisioned machine: %v", err)
		}
		if err := updateBootstrapToken(masterMachine, masterProvisionedMachine); err != nil {
			log.Fatalf("Unable to update bootstrap token: %v", err)
		}
	}
	machineClientBuilder := sshmachine.NewClient
	insecureIgnoreHostKey := false
	if len(publicKeys) == 0 {
		insecureIgnoreHostKey = true
		log.Printf("Not able to verify machine SSH identity: No public keys given. Continuing...")
	}
	actuator := machineActuator.NewActuator(
		state.KubeClient,
		state.ClusterClient,
		state.SPClient,
		machineClientBuilder,
		insecureIgnoreHostKey,
		log.LogLevel(),
	)
	if err = actuator.Create(cluster, newMachine); err != nil {
		log.Fatalf("Unable to create machine: %v", err)
	}

	if clusterutil.RoleContains(clustercommon.NodeRole, newMachine.Spec.Roles) {
		if err := createAdminKubeConfigSecretIfNotPresent(); err != nil {
			log.Fatalf("Unable to create admin kubeconfig secret: %v", err)
		}
		if err := copyAdminConfigFromSecret(masterMachine, masterProvisionedMachine, newMachine, newProvisionedMachine); err != nil {
			log.Fatalf("Unable to place admin kubeconfig on the node: %v", err)
		}
	}

	if clusterutil.RoleContains(clustercommon.MasterRole, newMachine.Spec.Roles) {
		log.Println("Updating cluster status")
		// Update cluster etcd members
		machineStatus, err := sputil.GetMachineStatus(*newMachine)
		if err != nil {
			log.Fatalf("Unable to get machine %q status: %v", newMachine.Name, err)
		}
		if machineStatus.EtcdMember != nil {
			if err := insertClusterEtcdMember(*machineStatus.EtcdMember, cluster); err != nil {
				log.Fatalf("Unable to add etcd member to cluster status: %v", err)
			}
		}
		// Update cluster API endpoints
		var apiEndpoint *clusterv1.APIEndpoint
		// Use the controlPlaneEndpoint if it is defined
		apiEndpoint, err = controlPlaneEndpointFromMachine(newMachine, newProvisionedMachine)
		if err != nil {
			if err.Error() != "controlPlaneEndpoint is not defined" {
				log.Fatalf("Unable to get machine %q control plane endpoint: %v", newMachine.Name, err)
			}
			// If control plane endpoint is not defined, use the machine's advertised API address and port
			apiEndpoint, err = apiEndpointFromMachine(newMachine, newProvisionedMachine)
			if err != nil {
				log.Fatalf("Unable to get machine %q advertised API address and port: %v", newMachine.Name, err)
			}
		}

		apiEndpointSet := setsutil.NewAPIEndpointSet(cluster.Status.APIEndpoints...)
		apiEndpointSet.Insert(*apiEndpoint)
		cluster.Status.APIEndpoints = apiEndpointSet.List()

		_, err = state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).UpdateStatus(cluster)
		if err != nil {
			log.Fatalf("Unable to update cluster state: %v", err)
		}
	}

	if err := state.PullFromAPIs(); err != nil {
		log.Fatalf("Unable to sync on-disk state: %v", err)
	}
	log.Println("Machine created successfully.")
}

// machineCmdCreate represents the machine create command
var machineCmdCreate = &cobra.Command{
	Use:   "machine",
	Short: "Adds a machine to the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		ip := cmd.Flag("ip").Value.String()
		iface := cmd.Flag("iface").Value.String()
		role := strings.Title(cmd.Flag("role").Value.String())
		port, err := strconv.Atoi(cmd.Flag("port").Value.String())
		if err != nil {
			log.Fatalf("Invalid port %v", err)
		}
		publicKeyFiles, err := cmd.Flags().GetStringSlice("public-keys")
		if err != nil {
			log.Fatalf("Unable to parse `public-keys`: %v", err)
		}
		createMachine(ip, port, iface, role, publicKeyFiles)
	},
}

func getGoalComponentVersions() *spv1.MachineComponentVersions {
	return &spv1.MachineComponentVersions{
		NodeadmVersion:    common.DefaultNodeadmVersion,
		EtcdadmVersion:    common.DefaultEtcdadmVersion,
		KubernetesVersion: common.DefaultKubernetesVersion,
		CNIVersion:        common.DefaultCNIVersion,
		KeepalivedVersion: common.DefaultKeepalivedVersion,
		FlannelVersion:    common.DefaultFlannelVersion,
		EtcdVersion:       common.DefaultEtcdVersion,
	}
}

func newProvisionedMachineAndMachine(name string, role clustercommon.MachineRole, vipNetworkInterface string, sshConfig spv1.SSHConfig) (*spv1.ProvisionedMachine, *clusterv1.Machine, error) {
	newProvisionedMachine := spv1.ProvisionedMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ProvisionedMachine",
			APIVersion: "sshprovider.platform9.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         common.DefaultNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: spv1.ProvisionedMachineSpec{
			SSHConfig:           &sshConfig,
			VIPNetworkInterface: vipNetworkInterface,
		},
	}

	newMachine := clusterv1.Machine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Machine",
			APIVersion: "cluster.k8s.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         common.DefaultNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: clusterv1.MachineSpec{
			Roles: []clustercommon.MachineRole{role},
		},
		Status: clusterv1.MachineStatus{},
	}

	if role == clustercommon.MasterRole {
		newMachine.Spec.Taints = []corev1.Taint{
			{
				Key:    common.LabelNodeRoleMaster,
				Effect: corev1.TaintEffectPreferNoSchedule,
			},
		}
	}

	machineProviderSpec := spv1.MachineSpec{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sshprovider.platform9.com/v1alpha1",
			Kind:       "MachineSpec",
		},
		ProvisionedMachineName: newProvisionedMachine.Name,
		Roles: []spv1.MachineRole{
			spv1.MachineRole(role),
		},
		ComponentVersions: getGoalComponentVersions(),
	}
	if err := sputil.PutMachineSpec(machineProviderSpec, &newMachine); err != nil {
		return nil, nil, fmt.Errorf("unable to encode machine provider spec: %v", err)
	}

	machineProviderStatus := spv1.MachineStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sshprovider.platform9.com/v1alpha1",
			Kind:       "MachineStatus",
		},
	}
	if err := sputil.PutMachineStatus(machineProviderStatus, &newMachine); err != nil {
		return nil, nil, fmt.Errorf("unable to encode machine provider status: %v", err)
	}

	if err := sputil.BindMachineAndProvisionedMachine(&newMachine, &newProvisionedMachine); err != nil {
		return nil, nil, fmt.Errorf("unable to create bi-directional bind between machine and provisioned machine: %v", err)
	}
	return &newProvisionedMachine, &newMachine, nil
}

func deleteMachine(ip string, force bool, skipDrainDelete bool) {
	targetMachine, err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).Get(ip, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Unable to get machine %q: %v", ip, err)
	}
	targetMachineSpec, err := sputil.GetMachineSpec(*targetMachine)
	if err != nil {
		log.Fatalf("Unable to decode machine %q spec: %v", targetMachine.Name, err)
	}
	targetProvisionedMachine, err := state.SPClient.SshproviderV1alpha1().ProvisionedMachines(common.DefaultNamespace).Get(targetMachineSpec.ProvisionedMachineName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Unable to get provisioned machine %q: %v", targetMachineSpec.ProvisionedMachineName, err)
	}
	cluster, err := state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Get(common.DefaultClusterName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Unable to get cluster: %v", err)
	}

	if force {
		log.Println("--force enabled: skipping node drain, node delete, and commands invoked on the machine")
	} else {
		deleteMustNotOrphanNodes(targetMachine)
		if !skipDrainDelete {
			if err := drainAndDeleteNodeForMachine(targetMachine, targetProvisionedMachine); err != nil {
				log.Fatalf("Unable to drain and delete cluster node for machine %q: %v", targetMachine.Name, err)
			}
		}

		var insecureIgnoreHostKey bool
		if len(targetProvisionedMachine.Spec.SSHConfig.PublicKeys) == 0 {
			insecureIgnoreHostKey = true
			log.Printf("Not able to verify machine SSH identity: No public keys given. Continuing...")
		}
		machineClientBuilder := sshmachine.NewClient
		actuator := machineActuator.NewActuator(
			state.KubeClient,
			state.ClusterClient,
			state.SPClient,
			machineClientBuilder,
			insecureIgnoreHostKey,
			log.LogLevel(),
		)
		log.Println("Deleting machine")
		if err = actuator.Delete(cluster, targetMachine); err != nil {
			log.Fatalf("Unable to delete machine: %v", err)
		}
	}

	log.Println("Updating cluster status")
	machineStatus, err := sputil.GetMachineStatus(*targetMachine)
	if err != nil {
		log.Fatalf("Unable to get machine %q status: %v", targetMachine.Name, err)
	}
	if machineStatus.EtcdMember != nil {
		if err := removeClusterEtcdMember(*machineStatus.EtcdMember, cluster); err != nil {
			log.Fatalf("Unable to delete etcd member from cluster status: %v", err)
		}
	}

	if err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).Delete(targetMachine.Name, &metav1.DeleteOptions{}); err != nil {
		log.Fatalf("unable to delete machine %q: %v", targetMachine.Name, err)
	}
	if err := state.SPClient.SshproviderV1alpha1().ProvisionedMachines(common.DefaultNamespace).Delete(targetProvisionedMachine.Name, &metav1.DeleteOptions{}); err != nil {
		log.Fatalf("unable to delete provisioned machine %q: %v", targetProvisionedMachine.Name, err)
	}

	if clusterutil.RoleContains(clustercommon.MasterRole, targetMachine.Spec.Roles) {
		// Update cluster API endpoints
		machines, err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).List(metav1.ListOptions{})
		if err != nil {
			log.Fatalf("unable to list machines: %v", err)
		}
		masters := clusterapi.MachinesWithRole(machines.Items, clustercommon.MasterRole)
		// It may not possible to identify the endpoint for the machine being
		// deleted, e.g. if the machine is failed and `kubeadm config view`
		// cannot be invoked. For now, assume there is only one endpoint, and
		// delete it after the last master is deleted.
		// See https://github.com/platform9/ssh-provider/issues/67
		if len(masters) == 0 {
			cluster.Status.APIEndpoints = []clusterv1.APIEndpoint{}
		}
		_, err = state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).UpdateStatus(cluster)
		if err != nil {
			log.Fatalf("Unable to update cluster state: %v", err)
		}
	}

	if err := state.PullFromAPIs(); err != nil {
		log.Fatalf("Unable to sync on-disk state: %v", err)
	}

	log.Println("Machine deleted successfully.")
}

var machineCmdDelete = &cobra.Command{
	Use:   "machine",
	Short: "Deletes a machine from the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		ip := cmd.Flag("ip").Value.String()
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			log.Fatalf("Unable to parse `force` flag: %v", err)
		}
		skipDrainDelete, err := cmd.Flags().GetBool("skip-drain-delete")
		if err != nil {
			log.Fatalf("Unable to parse `skip-drain-delete` flag: %v", err)
		}
		deleteMachine(ip, force, skipDrainDelete)
	},
}

func deleteMustNotOrphanNodes(targetMachine *clusterv1.Machine) {
	if clusterutil.RoleContains(clustercommon.MasterRole, targetMachine.Spec.Roles) {
		machineList, err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).List(metav1.ListOptions{})
		if err != nil {
			log.Fatalf("Unable to list machines: %v", err)
		}
		countMasters := 0
		countNodes := 0
		for _, machine := range machineList.Items {
			for _, role := range machine.Spec.Roles {
				switch role {
				case clustercommon.MasterRole:
					countMasters++
				case clustercommon.NodeRole:
					countNodes++
				}
			}
		}
		if countMasters == 1 && countNodes > 0 {
			log.Fatalf("Not deleting last master while %v nodes are in the cluster. Delete the nodes first.", countNodes)
		}
	}
}

func bootstrapTokenSecretFromMachine(machine *clusterv1.Machine, provisionedMachine *spv1.ProvisionedMachine) (*corev1.Secret, error) {
	machineClient, err := sshMachineClientFromSSHConfig(provisionedMachine.Spec.SSHConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create machine client for machine %q: %v", machine.Name, err)
	}
	cmd := "/opt/bin/kubeadm token create --print-join-command"
	stdOut, stdErr, err := machineClient.RunCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("error running %q: %v (%s) (%s)", cmd, err, string(stdOut), string(stdErr))
	}
	token, caHash, err := tokenAndCAHashFromKubeadmJoinCommand(string(stdOut))
	if err != nil {
		return nil, fmt.Errorf("unable to parse bootstrap token from stdout of %q: %q", cmd, stdOut)
	}
	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              common.DefaultBootstrapTokenSecretName,
			Namespace:         common.DefaultNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Data: map[string][]byte{
			"token":  []byte(token),
			"cahash": []byte(caHash),
		},
	}
	return &secret, nil
}

func masterMachineAndProvisionedMachine() (*clusterv1.Machine, *spv1.ProvisionedMachine, error) {
	machineList, err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list machines: %v", err)
	}
	var masterMachine *clusterv1.Machine
	for _, machine := range machineList.Items {
		if clusterutil.RoleContains(clustercommon.MasterRole, machine.Spec.Roles) {
			// Choose first master in the list
			masterMachine = machine.DeepCopy()
			break
		}
	}
	if masterMachine == nil {
		return nil, nil, fmt.Errorf("unable to find any machine with Master role, cannot obtain bootstrap token")
	}
	masterMachineSpec, err := sputil.GetMachineSpec(*masterMachine)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to decode machine spec: %v", err)
	}
	masterProvisionedMachine, err := state.SPClient.SshproviderV1alpha1().ProvisionedMachines(common.DefaultNamespace).Get(masterMachineSpec.ProvisionedMachineName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get provisioned machine: %v", err)
	}
	return masterMachine, masterProvisionedMachine.DeepCopy(), nil
}

// controlPlaneEndpointFromMachine returns the advertised API address and port
// of the API server on the machine
func controlPlaneEndpointFromMachine(machine *clusterv1.Machine, provisionedMachine *spv1.ProvisionedMachine) (*clusterv1.APIEndpoint, error) {
	machineClient, err := sshMachineClientFromSSHConfig(provisionedMachine.Spec.SSHConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create machine client for machine %q: %v", machine.Name, err)
	}

	cmd := `/opt/bin/kubeadm config view`
	stdOut, stdErr, err := machineClient.RunCommand(cmd)
	if err != nil {
		log.Println(stdOut)
		log.Println(stdErr)
		return nil, fmt.Errorf("unable to run %q on %q: %v", cmd, machine.Name, err)
	}
	kcc := kubeadmutil.ClusterConfiguration{}
	err = yaml.Unmarshal(stdOut, &kcc)
	if err != nil {
		return nil, fmt.Errorf("unable to read kubeadm ClusterConfiguration from machine %q:%v", machine.Name, err)
	}
	return kubeadmutil.APIEndpointFromClusterConfiguration(&kcc)
}

// apiEndpointFromMachine returns the advertised API address and port of the API
// server on the machine
func apiEndpointFromMachine(machine *clusterv1.Machine, provisionedMachine *spv1.ProvisionedMachine) (*clusterv1.APIEndpoint, error) {
	machineClient, err := sshMachineClientFromSSHConfig(provisionedMachine.Spec.SSHConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create machine client for machine %q: %v", machine.Name, err)
	}
	cmd := `grep -oP -- '--advertise-address=\K([0-9]{0,3}\.[0-9]{0,3}\.[0-9]{0,3}\.[0-9]{0,3})' /etc/kubernetes/manifests/kube-apiserver.yaml`
	stdOut, stdErr, err := machineClient.RunCommand(cmd)
	if err != nil {
		log.Println(stdOut)
		log.Println(stdErr)
		return nil, fmt.Errorf("unable to run %q on %q: %v", cmd, machine.Name, err)
	}
	apiAddr := net.ParseIP(strings.TrimSpace(string(stdOut)))
	if apiAddr == nil {
		return nil, fmt.Errorf("unable to parse advertised API address from %q", string(stdOut))
	}

	cmd = `grep -oP -- '--secure-port=\K([0-9]{1,5})' /etc/kubernetes/manifests/kube-apiserver.yaml`
	stdOut, stdErr, err = machineClient.RunCommand(cmd)
	if err != nil {
		log.Println(stdOut)
		log.Println(stdErr)
		return nil, fmt.Errorf("unable to run %q on %q: %v", cmd, machine.Name, err)
	}
	apiPort, err := strconv.Atoi(strings.TrimSpace(string(stdOut)))
	if err != nil {
		return nil, fmt.Errorf("unable to parse API secure port from %q", string(stdOut))
	}

	return &clusterv1.APIEndpoint{
		Host: apiAddr.String(),
		Port: apiPort,
	}, nil
}

func tokenAndCAHashFromKubeadmJoinCommand(cmdStdout string) (string, string, error) {
	fields := strings.Fields(cmdStdout)
	//Successful output would be of the type
	//kubeadm join <server:port> --token <token> --discovery-token-ca-cert-hash <sha>
	if len(fields) != 7 { //TODO(puneet) Needs a better way but seems good-enough for now
		return "", "", fmt.Errorf("expected 7 fields, found %v", len(fields))
	}
	token := fields[4]
	caHash := fields[6]
	return token, caHash, nil
}

func adminKubeconfigFromMachine(machine *clusterv1.Machine, provisionedMachine *spv1.ProvisionedMachine) ([]byte, error) {
	machineClient, err := sshMachineClientFromSSHConfig(provisionedMachine.Spec.SSHConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create machine client for machine %q: %v", machine.Name, err)
	}
	// chmod file for read access for all users
	stdOut, stdErr, err := machineClient.RunCommand("chmod 0644 /etc/kubernetes/admin.conf")
	if err != nil {
		log.Println(stdOut)
		log.Println(stdErr)
		return nil, fmt.Errorf("unable to change kubeconfig file permissions on %q: %v", machine.Name, err)
	}
	fileContents, err := machineClient.ReadFile("/etc/kubernetes/admin.conf")
	if err != nil {
		return nil, fmt.Errorf("unable to read kubeconfig from machine %q:%v", machine.Name, err)
	}
	// chmod file to keep it secure
	stdOut, stdErr, err = machineClient.RunCommand("chmod 0600 /etc/kubernetes/admin.conf")
	if err != nil {
		log.Println(stdOut)
		log.Println(stdErr)
		return nil, fmt.Errorf("unable to change kubeconfig file permissions on %q: %v", machine.Name, err)
	}
	return fileContents, nil
}

func writeAdminKubeconfigToMachine(kubeconfig []byte, machine *clusterv1.Machine, provisionedMachine *spv1.ProvisionedMachine) error {
	machineClient, err := sshMachineClientFromSSHConfig(provisionedMachine.Spec.SSHConfig)
	if err != nil {
		return fmt.Errorf("unable to create machine client for machine %q: %v", machine.Name, err)
	}
	// write kubeconfig to /tmp first and then move to /etc
	if err := machineClient.WriteFile("/tmp/admin.conf", 0600, kubeconfig); err != nil {
		return fmt.Errorf("unable to write kubeconfig to machine %q: %v", machine.Name, err)
	}
	// move kubeconfig from /tmp to /etc/kubernetes
	return machineClient.MoveFile("/tmp/admin.conf", "/etc/kubernetes/admin.conf")
}

func drainAndDeleteNodeForMachine(targetMachine *clusterv1.Machine, targetProvisionedMachine *spv1.ProvisionedMachine) error {
	var err error
	targetMachineClient, err := sshMachineClientFromSSHConfig(targetProvisionedMachine.Spec.SSHConfig)
	if err != nil {
		return fmt.Errorf("unable to create machine client for machine %q: %v", targetMachine.Name, err)
	}
	nodeName, err := nodeNameForMachine(targetMachine.Name, targetMachineClient)
	if err != nil {
		return fmt.Errorf("unable to get node name: %v", err)
	}
	if len(nodeName) != 0 {
		log.Printf("Draining cluster node %q for machine %q", nodeName, targetMachine.Name)
		if err := drainNode(nodeName, targetMachineClient); err != nil {
			return fmt.Errorf("unable to drain node: %v", err)
		}
		log.Printf("Deleting cluster node %q for machine %q", nodeName, targetMachine.Name)
		return deleteNode(nodeName, targetMachineClient)

	}
	return nil
}

func sshMachineClientFromSSHConfig(sshConfig *spv1.SSHConfig) (sshmachine.Client, error) {
	sshCredentialSecret, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Get(sshConfig.CredentialSecret.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("unable to find SSH credential %q", sshConfig.CredentialSecret.Name)
		}
		return nil, fmt.Errorf("unable to get SSH credential secret: %v", err)
	}
	username, privateKey, err := sputil.UsernameAndKeyFromSecret(sshCredentialSecret)
	if err != nil {
		return nil, fmt.Errorf("unable to read SSH credential from secret: %v", err)
	}
	var insecureIgnoreHostKey bool
	if len(sshConfig.PublicKeys) == 0 {
		insecureIgnoreHostKey = true
		log.Printf("Not able to verify machine SSH identity: No public keys given. Continuing...")
	}
	return sshmachine.NewClient(sshConfig.Host, sshConfig.Port, username, privateKey, sshConfig.PublicKeys, insecureIgnoreHostKey)
}

var machineCmdGet = &cobra.Command{
	Use:   "machine",
	Short: "Get machine resources",
	Run: func(cmd *cobra.Command, args []string) {
		ip := cmd.Flag("ip").Value.String()
		var machineList *clusterv1.MachineList
		if len(ip) == 0 {
			var err error
			machineList, err = state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).List(metav1.ListOptions{})
			if err != nil {
				log.Fatalf("Unable to list machines: %v", err)
			}
		} else {
			machine, err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).Get(ip, metav1.GetOptions{})
			if err != nil {
				log.Fatalf("Unable to get machine %q: %v", ip, err)
			}
			machineList = &clusterv1.MachineList{
				Items: []clusterv1.Machine{*machine},
			}
		}
		switch outputFmt {
		case "yaml":
			bytes, err := yaml.Marshal(machineList.Items)
			if err != nil {
				log.Fatalf("Unable to marshal cluster spec file to yaml: %s", err)
			}
			os.Stdout.Write(bytes)
		case "json":
			bytes, err := json.Marshal(machineList.Items)
			if err != nil {
				log.Fatalf("Unable to marshal cluster spec file to json: %s", err)
			}
			os.Stdout.Write(bytes)
		case "":
			t := template.Must(template.New("MachineV1PrintTemplate").Parse(common.MachineV1PrintTemplate))
			if err := t.Execute(os.Stdout, machineList.Items); err != nil {
				log.Fatalf("Could not pretty print cluster details: %s", err)
			}
		default:
			log.Fatalf("Unsupported output format %q", outputFmt)
		}
	},
}

type UpgradeRequired struct {
	NodeadmVersion    bool
	EtcdadmVersion    bool
	KubernetesVersion bool
	CNIVersion        bool
	FlannelVersion    bool
	KeepalivedVersion bool
	EtcdVersion       bool
}

func isUpgradeRequired(old *spv1.MachineComponentVersions, cur *spv1.MachineComponentVersions) (bool, UpgradeRequired) {
	if cmp.Equal(old, cur) {
		return false, UpgradeRequired{}
	}

	return true, UpgradeRequired{
		old.NodeadmVersion != cur.NodeadmVersion,
		old.EtcdadmVersion != cur.EtcdadmVersion,
		old.KubernetesVersion != cur.KubernetesVersion,
		old.CNIVersion != cur.CNIVersion,
		old.FlannelVersion != cur.FlannelVersion,
		old.KeepalivedVersion != cur.KeepalivedVersion,
		old.EtcdVersion != cur.EtcdVersion,
	}
}

type instanceStatus *clusterv1.Machine

func getGoalMachine(currentMachine *clusterv1.Machine) (*clusterv1.Machine, error) {
	currentMachineSpec, err := sputil.GetMachineSpec(*currentMachine)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode machine %q spec: %v", currentMachine.Name, err)
	}
	// Prepare goal machine object using current machine
	goalMachine := currentMachine.DeepCopy()

	// Actions required on upgrade
	currentKubernetesVersion := semver.New(currentMachineSpec.ComponentVersions.KubernetesVersion)
	// When upgrading from < 1.11
	if currentKubernetesVersion.LessThan(*semver.New("1.11.0")) {
		// Master machines need a workaround for https://github.com/kubernetes/kubeadm/issues/1358
		if clusterutil.RoleContains(clustercommon.MasterRole, currentMachine.Spec.Roles) && len(currentMachine.Spec.Taints) == 0 {
			goalMachine.Spec.Taints = []corev1.Taint{
				{
					Key:    common.LabelNodeRoleMaster,
					Effect: corev1.TaintEffectPreferNoSchedule,
				},
			}
		}
	}

	goalMachineSpec, err := sputil.GetMachineSpec(*goalMachine)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode machine %q spec: %v", goalMachine.Name, err)
	}
	goalMachineSpec.ComponentVersions = getGoalComponentVersions()
	sputil.PutMachineSpec(*goalMachineSpec, goalMachine)
	// Add current machine as goal machine's annotation
	if currentMachine.ObjectMeta.Annotations == nil {
		currentMachine.ObjectMeta.Annotations = make(map[string]string)
	}
	if _, err := sputil.PutMachineInstanceStatus(goalMachine, currentMachine); err != nil {
		return nil, fmt.Errorf("Unable to set machine instance status %v", err)
	}
	return goalMachine, nil
}

func upgradeMachine(ip string) error {
	log.Printf("Upgrading machine %s\n", ip)
	// Get the current machine
	currentMachine, err := state.ClusterClient.ClusterV1alpha1().
		Machines(common.DefaultNamespace).
		Get(ip, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get machine %q: %v", ip, err)
	}
	currentMachineSpec, err := sputil.GetMachineSpec(*currentMachine)
	if err != nil {
		return fmt.Errorf("unable to decode machine %q spec: %v", currentMachine.Name, err)
	}
	currentProvisionedMachine, err := state.SPClient.SshproviderV1alpha1().
		ProvisionedMachines(common.DefaultNamespace).
		Get(currentMachineSpec.ProvisionedMachineName, metav1.GetOptions{})

	// Check if upgrade is required
	goalComponentVersions := getGoalComponentVersions()
	upgradeRequired, upgrade := isUpgradeRequired(currentMachineSpec.ComponentVersions, goalComponentVersions)
	if !upgradeRequired {
		log.Println("Machine is up to date.")
		return nil
	}

	// If any of the components except for nodeadm/etcdadm were updated, trigger an actuator update
	if upgrade.KubernetesVersion || upgrade.CNIVersion || upgrade.FlannelVersion ||
		upgrade.KeepalivedVersion ||
		upgrade.EtcdVersion {

		targetMachineClient, err := sshMachineClientFromSSHConfig(currentProvisionedMachine.Spec.SSHConfig)
		if err != nil {
			return fmt.Errorf("unable to create machine client for machine %q: %v", currentMachine.Name, err)
		}
		// Prepare goal machine using current machine
		goalMachine, err := getGoalMachine(currentMachine)
		if err != nil {
			return fmt.Errorf("unable to create goal machine object: %v", err)
		}

		// Drain current node
		nodeName, err := nodeNameForMachine(currentMachine.Name, targetMachineClient)
		if err != nil {
			return fmt.Errorf("unable to get node name for machine %s: %v", currentMachine.Name, err)
		}
		if err := drainNode(nodeName, targetMachineClient); err != nil {
			return fmt.Errorf("unable to drain the node %s: %v", nodeName, err)
		}

		// Instantiate actuator
		machineClientBuilder := sshmachine.NewClient
		insecureIgnoreHostKey := false
		if len(currentProvisionedMachine.Spec.SSHConfig.PublicKeys) == 0 {
			insecureIgnoreHostKey = true
			log.Printf("Not able to verify machine SSH identity: No public keys given. Continuing...")
		}
		actuator := machineActuator.NewActuator(
			state.KubeClient,
			state.ClusterClient,
			state.SPClient,
			machineClientBuilder,
			insecureIgnoreHostKey,
			log.LogLevel(),
		)

		// If goal machine is a node we would have to update the token
		// as current token might have expired
		var masterMachine *clusterv1.Machine
		var masterProvisionedMachine *spv1.ProvisionedMachine
		if clusterutil.RoleContains(clustercommon.NodeRole, goalMachine.Spec.Roles) {
			var err error
			masterMachine, masterProvisionedMachine, err = masterMachineAndProvisionedMachine()
			if err != nil {
				return fmt.Errorf("unable to get a master machine and provisioned machine: %v", err)
			}
			if err = updateBootstrapToken(masterMachine, masterProvisionedMachine); err != nil {
				return fmt.Errorf("unable to update bootstrap token for node")
			}
		}

		// Call actuator's update
		cluster, err := state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Get(common.DefaultClusterName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("unable to get cluster %s: %v", common.DefaultClusterName, err)
		}
		currentMachineStatus, err := sputil.GetMachineStatus(*currentMachine)
		if err != nil {
			return fmt.Errorf("unable to get machine status: %v", err)
		}
		// We are deleting etcd member prior to actual delete from actuator
		// this is still valid as delete only needs memberid (available in machine status)
		// if delete called in actuator update succeeds this would allow us to pass correct cluster state to create
		// if delete called in actuator update fails this state would never get persisted
		if clusterutil.RoleContains(clustercommon.MasterRole, goalMachine.Spec.Roles) {
			if err := removeClusterEtcdMember(*currentMachineStatus.EtcdMember, cluster); err != nil {
				return fmt.Errorf("unable to delete etcd member from cluster status")
			}
		}
		if err := actuator.Update(cluster, goalMachine); err != nil {
			return fmt.Errorf("unable to update the node %s: %v", nodeName, err)
		}
		goalMachineStatus, err := sputil.GetMachineStatus(*goalMachine)
		if err != nil {
			return fmt.Errorf("unable to get machine status: %v", err)
		}
		if clusterutil.RoleContains(clustercommon.MasterRole, goalMachine.Spec.Roles) {
			if err := insertClusterEtcdMember(*goalMachineStatus.EtcdMember, cluster); err != nil {
				return fmt.Errorf("unable to add etcd member from cluster status")
			}
		}

		//Uncordon upgraded node
		if clusterutil.RoleContains(clustercommon.NodeRole, goalMachine.Spec.Roles) {
			if err := createAdminKubeConfigSecretIfNotPresent(); err != nil {
				log.Fatalf("Unable to create admin kubeconfig secret: %v", err)
			}
			if err := copyAdminConfigFromSecret(masterMachine, masterProvisionedMachine, goalMachine, currentProvisionedMachine); err != nil {
				return fmt.Errorf("unable to copy admin kubeconfig to node: %v", err)
			}
		}
		if err := uncordonNode(nodeName, targetMachineClient); err != nil {
			return fmt.Errorf("unable to uncordon the node %s: %v", nodeName, err)
		}

		//Reset annotation to empty
		goalMachine.ObjectMeta.Annotations[common.InstanceStatusAnnotationKey] = ""

		currentMachine = goalMachine.DeepCopy()
		log.Println("Machine upgraded successfully.")
	} else {
		// A nodeadm/etcdadm version change does not require an actuator call, just a state file update
		if upgrade.NodeadmVersion || upgrade.EtcdadmVersion {
			currentMachineSpec.ComponentVersions.NodeadmVersion = goalComponentVersions.NodeadmVersion
			currentMachineSpec.ComponentVersions.EtcdadmVersion = goalComponentVersions.EtcdadmVersion
			log.Println("Nodeadm/Etcdadm only change, updating state file.")

			if err := sputil.PutMachineSpec(*currentMachineSpec, currentMachine); err != nil {
				return fmt.Errorf("unable to encode machine provider spec: %v", err)
			}
			log.Println("Machine upgraded successfully.")
		}
	}
	if _, err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).
		Update(currentMachine); err != nil {
		return fmt.Errorf("unable to update machine: %v", err)
	}
	if err := state.PullFromAPIs(); err != nil {
		return fmt.Errorf("unable to sync on-disk state: %v", err)
	}
	return nil
}
func nodeNameForMachine(machineName string, machineClient sshmachine.Client) (string, error) {
	log.Printf("Reading system UUID of machine %q", machineName)
	cmd := fmt.Sprintf("cat %s", common.SystemUUIDFile)
	stdOut, stdErr, err := machineClient.RunCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("error running %q: %v (%s) (%s)", cmd, err, string(stdOut), string(stdErr))
	}
	systemUUID := strings.TrimSpace(string(stdOut))
	// TODO(dlipovetsky) Handle the case when kubectl is not found. Possibly
	// infer that the nodeadm reset ran at least as far as removing the kubectl
	// binary. nodeName includes the object kind, i.e.,

	log.Printf("Identifying node for machine %q", machineName)
	// Requires sudo because the kubelet kubeconfig is readable by only by root.
	cmd = fmt.Sprintf(`%s --kubeconfig=%s get nodes -ojsonpath='{.items[?(@.status.nodeInfo.systemUUID=="%s")].metadata.name}'`, common.KubectlFile, common.KubeletKubeconfig, systemUUID)
	stdOut, stdErr, err = machineClient.RunCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("error running %q: %v (%s) (%s)", cmd, err, string(stdOut), string(stdErr))
	}
	nodeName := strings.TrimSpace(string(stdOut))
	return nodeName, nil
}

func drainNode(nodeName string, machineClient sshmachine.Client) error {
	// Requires sudo because the admin kubeconfig is readable by only by
	// root.
	// Use the admin kubeconfig because admin permissions are required to
	// drain.
	// Use --ignore-daemonsets because any DaemonSet-managed Pods will
	// prevent the drain otherwise, and because all Nodes have DaemonSet
	// Pods (kube-proxy, overlay network).
	cmd := fmt.Sprintf("%s --kubeconfig=%s drain %s --timeout=%v --grace-period=%v --delete-local-data=%v --force=%v --ignore-daemonsets", common.KubectlFile, common.AdminKubeconfig, nodeName, drainTimeout, drainGracePeriodSeconds, drainDeleteLocalData, drainForce)
	stdOut, stdErr, err := machineClient.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("error running %q: %v (%s) (%s)", cmd, err, string(stdOut), string(stdErr))
	}
	log.Println(string(stdOut))
	return nil
}

func deleteNode(nodeName string, machineClient sshmachine.Client) error {
	// Requires sudo because the kubelet kubeconfig is readable by only by
	// root.
	cmd := fmt.Sprintf("%s --kubeconfig=%s delete node %s", common.KubectlFile, common.KubeletKubeconfig, nodeName)
	stdOut, stdErr, err := machineClient.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("error running %q: %v (%s) (%s)", cmd, err, string(stdOut), string(stdErr))
	}
	log.Println(string(stdOut))
	return nil
}

func uncordonNode(nodeName string, machineClient sshmachine.Client) error {
	// Requires sudo because the kubelet kubeconfig is readable by only by
	// root.
	cmd := fmt.Sprintf("%s --kubeconfig=%s uncordon %s", common.KubectlFile, common.AdminKubeconfig, nodeName)
	stdOut, stdErr, err := machineClient.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("error running %q: %v (%s) (%s)", cmd, err, string(stdOut), string(stdErr))
	}
	log.Println(string(stdOut))
	return nil
}

var machineCmdUpgrade = &cobra.Command{
	Use:   "machine",
	Short: "Upgrade machine",
	Run: func(cmd *cobra.Command, args []string) {
		ip := cmd.Flag("ip").Value.String()
		if err := upgradeMachine(ip); err != nil {
			log.Fatalf("Upgrade machine failed with error : %v", err)
		}
	},
}

var machineBundleCmd = &cobra.Command{
	Use:   "machine",
	Short: "Create a support bundle for a node",
	Run: func(cmd *cobra.Command, args []string) {
		ip := cmd.Flag("ip").Value.String()
		targetMachine, err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).Get(ip, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("Unable to get machine %q: %v", ip, err)
		}
		targetMachineSpec, err := sputil.GetMachineSpec(*targetMachine)
		if err != nil {
			log.Fatalf("Unable to decode machine %q spec: %v", targetMachine.Name, err)
		}
		targetProvisionedMachine, err := state.SPClient.SshproviderV1alpha1().ProvisionedMachines(common.DefaultNamespace).Get(targetMachineSpec.ProvisionedMachineName, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("Unable to get provisioned machine %q: %v", targetMachineSpec.ProvisionedMachineName, err)
		}
		targetMachineClient, err := sshMachineClientFromSSHConfig(targetProvisionedMachine.Spec.SSHConfig)
		if err != nil {
			log.Fatalf("unable to create machine client for machine %q: %v", targetMachine.Name, err)
		}
		t := time.Now()

		bundleFileBaseName := fmt.Sprintf("%s-%s-%s.tgz", common.SupportBundleFileNamePrefix, ip, t.Format(time.RFC3339))
		localPath := cmd.Flag("output").Value.String()
		if len(localPath) == 0 {
			localPath = bundleFileBaseName
		}
		remotePath := path.Join(common.DashcamBundleBaseDir, bundleFileBaseName)
		command := fmt.Sprintf("%s bundle --output %s", common.DashcamCommandPath, remotePath)
		log.Printf("Started creating support bundle for %s. This will take a few minutes.", ip)
		stdOut, stdErr, err := targetMachineClient.RunCommand(command)
		if err != nil {
			log.Fatalf("Failed to create support bundle %q: %v (stdout: %q, stderr: %q)", command, err, string(stdOut), string(stdErr))
		}
		defer targetMachineClient.RemoveFile(remotePath)
		if err = downloadRemoteFile(remotePath, localPath, targetMachineClient); err != nil {
			log.Fatalf("Failed to download support bundle: %v", err)
		}
		log.Infof("cctl bundle downloaded to %s ", localPath)
	},
}

func init() {
	createCmd.AddCommand(machineCmdCreate)
	machineCmdCreate.Flags().String("ip", "", "IP of the machine")
	machineCmdCreate.Flags().Int("port", common.DefaultSSHPort, "SSH port")
	machineCmdCreate.Flags().String("role", "", "Role of the machine. Can be master/node")
	machineCmdCreate.Flags().StringSlice("public-keys", []string{}, "The machine's SSH public keys. Provide a comma-separated list, or define multiple flags.")
	machineCmdCreate.Flags().String("iface", "eth0", "Interface that keepalived will bind to in case of master")

	deleteCmd.AddCommand(machineCmdDelete)
	machineCmdDelete.Flags().String("ip", "", "IP of the machine")
	machineCmdDelete.Flags().Bool("force", false, "Force delete the machine")
	machineCmdDelete.Flags().Bool("skip-drain-delete", false, "Do not drain and delete the cluster node for the machine")
	machineCmdDelete.Flags().DurationVar(&drainTimeout, "drain-timeout", common.DrainTimeout, "The length of time to wait before giving up, zero means infinite")
	machineCmdDelete.Flags().IntVar(&drainGracePeriodSeconds, "drain-grace-period", common.DrainGracePeriodSeconds, "Period of time in seconds given to each pod to terminate gracefully. If negative, the default value specified in the pod will be used.")
	machineCmdDelete.Flags().BoolVar(&drainDeleteLocalData, "drain-delete-local-data", common.DrainDeleteLocalData, "Continue even if there are pods using emptyDir (local data that will be deleted when the node is drained).")
	machineCmdDelete.Flags().BoolVar(&drainForce, "drain-force", common.DrainForce, "Continue even if there are pods not managed by a ReplicationController, ReplicaSet, Job, DaemonSet or StatefulSet.")

	machineCmdGet.Flags().String("ip", "", "IP of the machine")
	getCmd.AddCommand(machineCmdGet)

	machineCmdUpgrade.Flags().String("ip", "", "IP of the machine")
	upgradeCmd.AddCommand(machineCmdUpgrade)

	bundleCmd.AddCommand(machineBundleCmd)
	machineBundleCmd.Flags().String("output", "", fmt.Sprintf("File path for bundle tgz file (default \"%s-<ip>-<timestamp>.tgz\" created in current directory)", common.SupportBundleFileNamePrefix))
	machineBundleCmd.Flags().String("ip", "", "IP address of the machine")
	machineBundleCmd.MarkFlagRequired("ip")
}
