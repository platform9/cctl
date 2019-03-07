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
	"io/ioutil"
	"net"
	"os"
	"strings"
	"text/template"

	log "github.com/platform9/cctl/pkg/logrus"

	"github.com/coreos/go-semver/semver"
	"github.com/ghodss/yaml"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/platform9/cctl/common"
	"github.com/platform9/cctl/pkg/util/clusterapi"
	"github.com/platform9/cctl/pkg/util/secret"
	"github.com/platform9/cctl/semverutil"

	spconstants "github.com/platform9/ssh-provider/constants"
	spv1 "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	sputil "github.com/platform9/ssh-provider/pkg/controller"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clustercommon "sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

var (
	forceDelete bool
	routerID    int
	vip         string
)

// clusterCmd represents the cluster command
var clusterCmdCreate = &cobra.Command{
	Use:   "cluster",
	Short: "Creates clusterspec in the current directory",
	Run: func(cmd *cobra.Command, args []string) {

		if cmd.Flag("file").Changed {
			clusterObjFile := cmd.Flag("f").Value.String()
			clusterObj, err := clusterFromFile(clusterObjFile)
			if err != nil {
				log.Fatalf("Unable to parse cluster object %v", err)
			}

			if _, err := state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Create(clusterObj); err != nil {
				log.Fatalf("Unable to create cluster %q: %v", common.DefaultClusterName, err)
			}
			if err := state.PullFromAPIs(); err != nil {
				log.Fatalf("Unable to sync on-disk state: %v", err)
			}
			log.Println("Cluster created successfully.")
			return
		}

		// If either vip or routerID is defined, then both must be
		if cmd.Flag("vip").Changed != cmd.Flag("router-id").Changed {
			log.Fatalf("Must use both --router-id and --vip, or use neither for a non-HA cluster.")
		}
		var vipConfig *spv1.VIPConfiguration
		if cmd.Flag("vip").Changed && cmd.Flag("router-id").Changed {
			if routerID > 255 || routerID < 0 {
				log.Fatalf("The --router-id %d must be between [0,255].", routerID)
			}
			if parsedIP := net.ParseIP(vip); parsedIP == nil {
				log.Fatalf("The --vip %s must be a valid IP.", vip)
			}
			vipConfig = &spv1.VIPConfiguration{
				RouterID: routerID,
				IP:       vip,
			}
		}

		servicesCIDR := cmd.Flag("service-network").Value.String()
		podsCIDR := cmd.Flag("pod-network").Value.String()
		saPrivateKeyFile := cmd.Flag("sa-private-key").Value.String()
		saPublicKeyFile := cmd.Flag("sa-public-key").Value.String()
		if (len(saPrivateKeyFile) == 0) != (len(saPublicKeyFile) == 0) {
			log.Fatalf("Must specify both sa-private-key and sa-public-key")
		}
		apiServerCACertFile := cmd.Flag("apiserver-ca-cert").Value.String()
		apiServerCAKeyFile := cmd.Flag("apiserver-ca-key").Value.String()
		if (len(apiServerCAKeyFile) == 0) != (len(apiServerCAKeyFile) == 0) {
			log.Fatalf("Must specify both --apiserver-ca-cert and --apiserver-ca-key")
		}
		etcdCACertFile := cmd.Flag("etcd-ca-cert").Value.String()
		etcdCAKeyFile := cmd.Flag("etcd-ca-key").Value.String()
		if (len(etcdCAKeyFile) == 0) != (len(etcdCAKeyFile) == 0) {
			log.Fatalf("Must specify both --etcd-ca-cert and --etcd-ca-key")
		}
		frontProxyCACertFile := cmd.Flag("front-proxy-ca-cert").Value.String()
		frontProxyCAKeyFile := cmd.Flag("front-proxy-ca-key").Value.String()
		if (len(frontProxyCAKeyFile) == 0) != (len(frontProxyCAKeyFile) == 0) {
			log.Fatalf("Must specify both --front-proxy-ca-cert and --front-proxy-ca-key")
		}
		clusterConfig := &spv1.ClusterConfig{}
		var err error
		clusterConfigFile := cmd.Flag("cluster-config").Value.String()
		if len(clusterConfigFile) != 0 {
			clusterConfig, err = parseClusterConfigFromFile(clusterConfigFile)
			if err != nil {
				log.Fatalf("Unable to parse cluster config %v", err)
			}
		}
		setClusterConfigDefaults(clusterConfig)

		newAPIServerCASecret, err := secret.CreateCASecret(common.DefaultAPIServerCASecretName, apiServerCACertFile, apiServerCAKeyFile)
		if err != nil {
			log.Fatalf("Unable to generate API Server CA cert pair: %v", err)
		}
		newEtcdCASecret, err := secret.CreateCASecret(common.DefaultEtcdCASecretName, etcdCACertFile, etcdCAKeyFile)
		if err != nil {
			log.Fatalf("Unable to generate etcd CA cert pair: %v", err)
		}
		newFrontProxyCASecret, err := secret.CreateCASecret(common.DefaultFrontProxyCASecretName, frontProxyCACertFile, frontProxyCAKeyFile)
		if err != nil {
			log.Fatalf("Unable to generate front proxy CA cert pair: %v", err)
		}

		newServiceAccountKeySecret, err := secret.CreateSAKeySecret(common.DefaultServiceAccountKeySecretName, saPrivateKeyFile, saPublicKeyFile)
		if err != nil {
			log.Fatalf("Unable to generate service account key pair: %v", err)
		}
		newBootstrapTokenSecret, err := secret.CreateBootstrapTokenSecret(common.DefaultBootstrapTokenSecretName)
		if err != nil {
			log.Fatalf("Unable to generate bootstrap token secret: %v", err)
		}

		newCluster, err := createCluster(common.DefaultClusterName, podsCIDR, servicesCIDR, vipConfig, clusterConfig)
		if err != nil {
			log.Fatalf("Unable to create cluster: %v", err)
		}
		if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newAPIServerCASecret); err != nil {
			log.Fatalf("Unable to create API server CA secret: %v", err)
		}
		if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newEtcdCASecret); err != nil {
			log.Fatalf("Unable to create etcd CA secret: %v", err)
		}
		if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newFrontProxyCASecret); err != nil {
			log.Fatalf("Unable to create front proxy CA secret: %v", err)
		}
		if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newServiceAccountKeySecret); err != nil {
			log.Fatalf("Unable to create service account secret: %v", err)
		}
		if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newBootstrapTokenSecret); err != nil {
			log.Fatalf("Unable to create bootstrap token secret: %v", err)
		}
		if _, err := state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Create(newCluster); err != nil {
			log.Fatalf("Unable to create cluster %q: %v", common.DefaultClusterName, err)
		}
		if err := state.PullFromAPIs(); err != nil {
			log.Fatalf("Unable to sync on-disk state: %v", err)
		}
		log.Println("Cluster created successfully.")
	},
}

func setClusterConfigDefaults(clusterConfig *spv1.ClusterConfig) {
	setKubeAPIServerDefaults(clusterConfig)
	setKubeControllerMgrDefaults(clusterConfig)
	setKubeletConfigDefaults(clusterConfig)
}

func setKubeAPIServerDefaults(clusterConfig *spv1.ClusterConfig) {
	if clusterConfig.KubeAPIServer == nil {
		clusterConfig.KubeAPIServer = make(map[string]string)
	}
	// ServiceNodePortRange
	if _, ok := clusterConfig.KubeAPIServer[spconstants.KubeAPIServerServiceNodePortRangeKey]; !ok {
		clusterConfig.KubeAPIServer[spconstants.KubeAPIServerServiceNodePortRangeKey] = common.KubeAPIServerServiceNodePortRange
	}
}

func setKubeControllerMgrDefaults(clusterConfig *spv1.ClusterConfig) {
	if clusterConfig.KubeControllerManager == nil {
		clusterConfig.KubeControllerManager = make(map[string]string)
	}
	if _, ok := clusterConfig.KubeControllerManager[spconstants.KubeControllerMgrPodEvictionTimeoutKey]; !ok {
		clusterConfig.KubeControllerManager[spconstants.KubeControllerMgrPodEvictionTimeoutKey] = common.KubeControllerMgrPodEvictionTimeout
	}
}

func setKubeletConfigDefaults(clusterConfig *spv1.ClusterConfig) {
	if clusterConfig.Kubelet == nil {
		clusterConfig.Kubelet = &spv1.KubeletConfiguration{}
	}
	if clusterConfig.Kubelet.KubeAPIQPS == nil {
		clusterConfig.Kubelet.KubeAPIQPS = &common.KubeletKubeAPIQPS
	}
	if clusterConfig.Kubelet.KubeAPIBurst == 0 {
		clusterConfig.Kubelet.KubeAPIBurst = common.KubeletKubeAPIBurst
	}
	if clusterConfig.Kubelet.MaxPods == 0 {
		clusterConfig.Kubelet.MaxPods = common.KubeletMaxPods
	}
	if clusterConfig.Kubelet.FailSwapOn == nil {
		clusterConfig.Kubelet.FailSwapOn = &common.KubeletFailSwapOn
	}
	if clusterConfig.Kubelet.FeatureGates == nil {
		clusterConfig.Kubelet.FeatureGates = common.KubeletFeatureGates
	}
	if clusterConfig.Kubelet.EvictionHard == nil {
		clusterConfig.Kubelet.EvictionHard = common.KubeletEvictionHard
	}
	if clusterConfig.KubeAPIServer == nil {
		clusterConfig.KubeAPIServer = common.DefaultKubeAPIServerExtraArgs
	}
	if clusterConfig.KubeControllerManager == nil {
		clusterConfig.KubeControllerManager = common.DefaultKubeControllerManagerExtraArgs
	}
	if clusterConfig.KubeScheduler == nil {
		clusterConfig.KubeScheduler = common.DefaultKubeSchedulerExtraArgs
	}
}

func parseClusterConfigFromFile(file string) (*spv1.ClusterConfig, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read cluster config file: %s", file)
	}
	clusterConfig := spv1.ClusterConfig{}
	if err = yaml.Unmarshal(data, &clusterConfig); err != nil {
		return nil, fmt.Errorf("unable to decode cluster config: %v", err)
	}
	return &clusterConfig, nil
}

func clusterFromFile(file string) (*clusterv1.Cluster, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read cluster object: %s", file)
	}
	clusterObj := clusterv1.Cluster{}
	if err = yaml.Unmarshal(data, &clusterObj); err != nil {
		return nil, fmt.Errorf("unable to decode cluster object: %v", err)
	}
	return &clusterObj, nil
}

func createCluster(clusterName, podsCIDR, servicesCIDR string, vipConfig *spv1.VIPConfiguration, clusterConfig *spv1.ClusterConfig) (*clusterv1.Cluster, error) {
	newCluster := clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.k8s.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              clusterName,
			Namespace:         common.DefaultNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: clusterv1.ClusterSpec{
			ClusterNetwork: clusterv1.ClusterNetworkingConfig{
				Services: clusterv1.NetworkRanges{
					CIDRBlocks: []string{
						servicesCIDR,
					},
				},
				Pods: clusterv1.NetworkRanges{
					CIDRBlocks: []string{
						podsCIDR,
					},
				},
				ServiceDomain: "cluster.local",
			},
		},
	}

	spClusterSpec := spv1.ClusterSpec{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sshprovider.platform9.com/v1alpha1",
			Kind:       "ClusterSpec",
		},
		APIServerCASecret: &corev1.LocalObjectReference{
			Name: common.DefaultAPIServerCASecretName,
		},
		EtcdCASecret: &corev1.LocalObjectReference{
			Name: common.DefaultEtcdCASecretName,
		},
		FrontProxyCASecret: &corev1.LocalObjectReference{
			Name: common.DefaultFrontProxyCASecretName,
		},
		ServiceAccountKeySecret: &corev1.LocalObjectReference{
			Name: common.DefaultServiceAccountKeySecretName,
		},
		BootstrapTokenSecret: &corev1.LocalObjectReference{
			Name: common.DefaultBootstrapTokenSecretName,
		},
		ClusterConfig:    clusterConfig,
		VIPConfiguration: vipConfig,
	}

	spClusterStatus := spv1.ClusterStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sshprovider.platform9.com/v1alpha1",
			Kind:       "ClusterStatus",
		},
		EtcdMembers: []spv1.EtcdMember{},
	}

	sputil.PutClusterSpec(spClusterSpec, &newCluster)
	sputil.PutClusterStatus(spClusterStatus, &newCluster)

	return &newCluster, nil
}

var clusterCmdDelete = &cobra.Command{
	Use:   "cluster",
	Short: "Deletes a node from a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Running cluster delete")

		cluster, err := state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Get(common.DefaultClusterName, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("Unable to get cluster: %v", err)
		}

		machineList, err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).List(metav1.ListOptions{})
		if err != nil {
			log.Fatalf("Unable to list machines: %v", err)
		}

		if len(machineList.Items) > 0 {
			var machineNames []string
			for _, machine := range machineList.Items {
				machineNames = append(machineNames, machine.Name)
			}
			if forceDelete {
				log.Printf("Machines [%s] part of cluster. Deleting them from the state.", machineNames)
				for _, machine := range machineList.Items {
					if err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).Delete(machine.Name, &metav1.DeleteOptions{}); err != nil {
						if !apierrors.IsNotFound(err) {
							log.Fatalf("Unable to delete machine %q: %v", machine.Name, err)
						}
					}
				}
			} else {
				log.Fatalf("Machines [%s] part of cluster. Delete them before deleting the cluster.", machineNames)
			}
		}

		clusterProviderSpec, err := sputil.GetClusterSpec(*cluster)
		if err != nil {
			log.Fatalf("Unable to decode cluster spec: %v", err)
		}
		if clusterProviderSpec.APIServerCASecret != nil {
			if err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Delete(clusterProviderSpec.APIServerCASecret.Name, &metav1.DeleteOptions{}); err != nil {
				if !apierrors.IsNotFound(err) {
					log.Fatalf("Unable to delete API server CA secret: %v", err)
				}
			}
		}
		if clusterProviderSpec.EtcdCASecret != nil {
			if err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Delete(clusterProviderSpec.EtcdCASecret.Name, &metav1.DeleteOptions{}); err != nil {
				if !apierrors.IsNotFound(err) {
					log.Fatalf("Unable to delete etcd CA secret: %v", err)
				}
			}
		}
		if clusterProviderSpec.FrontProxyCASecret != nil {
			if err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Delete(clusterProviderSpec.FrontProxyCASecret.Name, &metav1.DeleteOptions{}); err != nil {
				if !apierrors.IsNotFound(err) {
					log.Fatalf("Unable to delete front proxy CA secret: %v", err)
				}
			}
		}
		if clusterProviderSpec.ServiceAccountKeySecret != nil {
			if err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Delete(clusterProviderSpec.ServiceAccountKeySecret.Name, &metav1.DeleteOptions{}); err != nil {
				if !apierrors.IsNotFound(err) {
					log.Fatalf("Unable to delete service account key secret: %v", err)
				}
			}
		}
		if clusterProviderSpec.BootstrapTokenSecret != nil {
			if err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Delete(clusterProviderSpec.BootstrapTokenSecret.Name, &metav1.DeleteOptions{}); err != nil {
				if !apierrors.IsNotFound(err) {
					log.Fatalf("Unable to delete bootstrap token secret: %v", err)
				}
			}
		}

		if err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Delete(common.DefaultAdminConfigSecretName, &metav1.DeleteOptions{}); err != nil {
			if !apierrors.IsNotFound(err) {
				log.Fatalf("Unable to delete admin kubeconfig secret: %v", err)
			}
		}

		if err := state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Delete(cluster.Name, &metav1.DeleteOptions{}); err != nil {
			if !apierrors.IsNotFound(err) {
				log.Fatalf("Unable to delete cluster: %v", err)
			}
		}

		if err := state.PullFromAPIs(); err != nil {
			log.Fatalf("Unable to sync on-disk state: %v", err)
		}
		log.Println("Cluster deleted successfully")
	},
}

var clusterCmdGet = &cobra.Command{
	Use:   "cluster",
	Short: "Get the cluster details",
	Run: func(cmd *cobra.Command, args []string) {
		cluster, err := state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Get(common.DefaultClusterName, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("Unable to get cluster: %v", err)
		}
		switch outputFmt {
		case "yaml":
			bytes, err := yaml.Marshal(cluster)
			if err != nil {
				log.Fatalf("Unable to marshal cluster spec file to yaml: %s", err)
			}
			os.Stdout.Write(bytes)
		case "json":
			bytes, err := json.Marshal(cluster)
			if err != nil {
				log.Fatalf("Unable to marshal cluster spec file to json: %s", err)
			}
			os.Stdout.Write(bytes)
		case "":
			// Pretty print cluster details
			clusterProviderSpec, err := sputil.GetClusterSpec(*cluster)
			if err != nil {
				log.Fatalf("Could not decode cluster provider spec: %v", err)
			}
			data := struct {
				Cluster             *clusterv1.Cluster
				ClusterProviderSpec *spv1.ClusterSpec
			}{
				Cluster:             cluster,
				ClusterProviderSpec: clusterProviderSpec,
			}
			t := template.Must(template.New("ClusterV1PrintTemplate").Parse(common.ClusterV1PrintTemplate))
			if err := t.Execute(os.Stdout, &data); err != nil {
				log.Fatalf("Could not pretty print cluster details: %s", err)
			}
		default:
			log.Fatalf("Unsupported output format %q", outputFmt)
		}
	},
}

func createLocalCopyOfAdminKubeConfig() (string, error) {
	kubeconfig, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Get(common.DefaultAdminConfigSecretName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("unable to get admin kubeconfig from secret: %v", err)
	}
	tmpKubeConfig, err := ioutil.TempFile("", common.TmpKubeConfigNamePrefix)
	if err != nil {
		return "", fmt.Errorf("unable to create temporary file : %v", err)
	}
	kubeconfigData, ok := kubeconfig.Data[common.DefaultAdminConfigSecretKey]
	if !ok {
		return "", fmt.Errorf("unable to find data in admin kubeconfig secret")
	}
	if len(kubeconfigData) == 0 {
		return "", fmt.Errorf("invalid data in admin kubeconfig secret")
	}
	err = ioutil.WriteFile(tmpKubeConfig.Name(), kubeconfigData, os.FileMode(os.O_RDONLY))
	if err != nil {
		return "", fmt.Errorf("unable to write kubeconfig to file : %v", err)
	}
	return tmpKubeConfig.Name(), nil
}

func checkClusterHealth() error {
	kubeconfig, err := createLocalCopyOfAdminKubeConfig()
	defer os.Remove(kubeconfig)
	if err != nil {
		return fmt.Errorf("unable to create local copy of kubeconfig : %v", err)
	}
	log.Print("Checking if all masters are in ready state")
	if err = common.MasterNodesReady(kubeconfig); err != nil {
		return err
	}
	log.Print("Checking if all control plane pods are in ready state")
	if err = common.ControlPlaneReady(kubeconfig); err != nil {
		return err
	}
	return nil
}

func trimVFromVersion(version string) string {
	return strings.TrimPrefix(version, "v")
}

func checkVersionSkew() error {
	machines, err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to get list of machines in the cluster")
	}
	// TODO(puneet) doing this check for every machine seems expensive
	// should we have a set of versions at cluster level as well?
	for _, machine := range machines.Items {
		machineSpec, err := sputil.GetMachineSpec(machine)
		if err != nil {
			return fmt.Errorf("unable to decode machine spec: %v", err)
		}
		machineK8sVersion, err := semver.NewVersion(machineSpec.ComponentVersions.KubernetesVersion)
		if err != nil {
			return fmt.Errorf("unable to parse kubernetes version for machine %s", machine.Name)
		}
		// minimum K8s version that we can upgrade from
		minimumK8sVersion, err := semver.NewVersion(trimVFromVersion(common.MinimumControlPlaneVersion))
		if err != nil {
			return fmt.Errorf("unable to parse kubernetes version %s", minimumK8sVersion)
		}
		if semverutil.CompareMajorMinorVersions(*machineK8sVersion, *minimumK8sVersion) < 0 {
			return fmt.Errorf("cannot upgrade machine %s. Minimum supported version for upgrade %s. Machine is currently at %s", machine.Name, minimumK8sVersion, machineK8sVersion)
		}
	}
	return nil
}

func upgradeMachines(machines []clusterv1.Machine) error {
	for _, machine := range machines {
		machineSpec, err := sputil.GetMachineSpec(machine)
		if err != nil {
			return fmt.Errorf("unable to decode machine spec: %v", err)
		}
		currentProvisionedMachine, err := state.SPClient.SshproviderV1alpha1().
			ProvisionedMachines(common.DefaultNamespace).
			Get(machineSpec.ProvisionedMachineName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("unable to decode provisioned machine spec: %v", err)
		}
		if err = upgradeMachine(currentProvisionedMachine.Spec.SSHConfig.Host); err != nil {
			return fmt.Errorf("Cluster upgrade failed with error: %v", err)
		}
	}
	return nil
}

var clusterCmdUpgrade = &cobra.Command{
	Use:   "cluster",
	Short: "Upgrade the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if err := createAdminKubeConfigSecretIfNotPresent(); err != nil {
			log.Fatalf("Unable to create admin kubeconfig secret: %v", err)
		}
		log.Print("[pre-flight] Running preflight checks for cluster upgrade")
		if err := checkVersionSkew(); err != nil {
			log.Fatalf("[pre-flight] Preflight check failed with error: %v", err)
		}
		if err := checkClusterHealth(); err != nil {
			log.Fatalf("[pre-flight] Preflight check failed with error: %v", err)
		}
		log.Print("[pre-flight] Preflight check passed")
		log.Print("Starting cluster upgrade")

		cluster, err := state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Get(common.DefaultClusterName, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("unable to get cluster %s: %v", common.DefaultClusterName, err)
		}
		clusterSpec, err := sputil.GetClusterSpec(*cluster)
		if err != nil {
			log.Fatalf("unable to get cluster spec %s: %v", common.DefaultClusterName, err)
		}
		if clusterSpec.ClusterConfig == nil {
			clusterSpec.ClusterConfig = &spv1.ClusterConfig{}
			setClusterConfigDefaults(clusterSpec.ClusterConfig)
		}
		if err := sputil.PutClusterSpec(*clusterSpec, cluster); err != nil {
			log.Fatalf("Unable to update cluster spec %s: %v", common.DefaultClusterName, err)
		}
		if _, err = state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Update(cluster); err != nil {
			log.Fatalf("unable to update cluster spec %s: %v", common.DefaultClusterName, err)
		}
		machines, err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).List(metav1.ListOptions{})
		if err != nil {
			log.Fatalf("unable to get list of machines in the cluster")
		}
		masters := clusterapi.MachinesWithRole(machines.Items, clustercommon.MasterRole)
		nodes := clusterapi.MachinesWithRole(machines.Items, clustercommon.NodeRole)
		log.Printf("Upgrading cluster masters")
		if err = upgradeMachines(masters); err != nil {
			log.Fatalf("Cluster upgrade failed with error: %v", err)
		}
		log.Printf("Upgrading cluster nodes")
		if err = upgradeMachines(nodes); err != nil {
			log.Fatalf("Cluster upgrade failed with error: %v", err)
		}
		if err := state.PullFromAPIs(); err != nil {
			log.Fatalf("Unable to sync on-disk state: %v", err)
		}
		log.Printf("Cluster upgraded successfully")
	},
}

func init() {
	createCmd.AddCommand(clusterCmdCreate)
	clusterCmdCreate.Flags().String("service-network", "10.1.0.0/16", "Network CIDR for services e.g. 10.1.0.0/16")
	clusterCmdCreate.Flags().String("pod-network", "10.2.0.0/16", "Network CIDR for pods e.g. 10.2.0.0.16")
	clusterCmdCreate.Flags().StringVar(&vip, "vip", "", "Virtual IP to be used for multi master setup")
	clusterCmdCreate.Flags().IntVar(&routerID, "router-id", -1, "Virtual router ID for keepalived for multi master setup. Must be in the range [0, 254]. Must be unique within a single L2 network domain.")
	clusterCmdCreate.Flags().String("apiserver-ca-cert", "", "The API Server CA certificate. Used to sign kubelet certificate requests and verify client certificates.")
	clusterCmdCreate.Flags().String("apiserver-ca-key", "", "The API Server CA certificate key.")
	clusterCmdCreate.Flags().String("etcd-ca-cert", "", "The etcd CA certificate. Used to sign and verify client and peer certificates.")
	clusterCmdCreate.Flags().String("etcd-ca-key", "", "The etcd CA certificate key.")
	clusterCmdCreate.Flags().String("front-proxy-ca-cert", "", "The front proxy CA certificate. Used to verify client certificates on incoming requests.")
	clusterCmdCreate.Flags().String("front-proxy-ca-key", "", "The front proxy CA certificate key.")
	clusterCmdCreate.Flags().String("sa-private-key", "", "Location of file containing private key used for signing service account tokens")
	clusterCmdCreate.Flags().String("sa-public-key", "", "Location of file containing public key used for signing service account tokens")
	clusterCmdCreate.Flags().String("cluster-config", "", "Location of file containing configurable parameters for the cluster")
	clusterCmdCreate.Flags().StringP("file", "f", "", "Location of file containing a cluster object")
	//clusterCmdCreate.Flags().String("version", "1.10.2", "Kubernetes version")

	deleteCmd.AddCommand(clusterCmdDelete)
	clusterCmdDelete.Flags().BoolVar(&forceDelete, "force", false, "Force delete a cluster")

	getCmd.AddCommand(clusterCmdGet)
	upgradeCmd.AddCommand(clusterCmdUpgrade)
	clusterCmdUpgrade.Flags().DurationVar(&drainTimeout, "drain-timeout", common.DrainTimeout, "The length of time to wait before giving up, zero means infinite")
	clusterCmdUpgrade.Flags().IntVar(&drainGracePeriodSeconds, "drain-grace-period", common.DrainGracePeriodSeconds, "Period of time in seconds given to each pod to terminate gracefully. If negative, the default value specified in the pod will be used.")
	clusterCmdUpgrade.Flags().BoolVar(&drainDeleteLocalData, "drain-delete-local-data", common.DrainDeleteLocalData, "Continue even if there are pods using emptyDir (local data that will be deleted when the node is drained).")
	clusterCmdUpgrade.Flags().BoolVar(&drainForce, "drain-force", common.DrainForce, "Continue even if there are pods not managed by a ReplicationController, ReplicaSet, Job, DaemonSet or StatefulSet.")
}
