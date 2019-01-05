package nodeadm

import (
	"fmt"
	"strconv"

	spconstants "github.com/platform9/ssh-provider/constants"
	spv1 "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	"github.com/platform9/ssh-provider/pkg/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type InitConfiguration struct {
	MasterConfiguration KubeadmMasterConfiguration `json:"masterConfiguration,omitempty"`

	VIPConfiguration *VIPConfiguration `json:"vipConfiguration,omitempty"`
	NetworkBackend   map[string]string `json:"networkBackend,omitempty"`
	KeepAlived       map[string]string `json:"keepAlived,omitempty"`
}

type JoinConfiguration struct {
	NodeConfiguration KubeadmNodeConfiguration `json:"nodeConfiguration"`
}

type VIPConfiguration struct {
	// The virtual IP.
	IP string `json:"ip,omitempty"`
	// The virtual router ID. Must be in the range [0, 254]. Must be unique within
	// a single L2 network domain.
	RouterID int `json:"routerID,omitempty"`
	// Network interface chosen to create the virtual IP. If it is not specified,
	// the interface of the default gateway is chosen.
	NetworkInterface string `json:"networkInterface,omitempty"`
}

func InitConfigurationForMachine(cluster clusterv1.Cluster, machine clusterv1.Machine, pm spv1.ProvisionedMachine) (*InitConfiguration, error) {
	cfg := &InitConfiguration{
		MasterConfiguration: KubeadmMasterConfiguration{
			TypeMeta: metav1.TypeMeta{
				Kind:       "MasterConfiguration",
				APIVersion: "kubeadm.k8s.io/v1alpha2",
			},
		},
	}

	cpc, err := controller.GetClusterSpec(cluster)
	if err != nil {
		return nil, fmt.Errorf("unable to decode cluster spec: %v", err)
	}

	mpc, err := controller.GetMachineSpec(machine)
	if err != nil {
		return nil, fmt.Errorf("unable to decode machine spec: %v", err)
	}

	// NodeRegistrationOptions
	cfg.MasterConfiguration.NodeRegistration.Name = machine.Name
	cfg.MasterConfiguration.NodeRegistration.Taints = machine.Spec.Taints

	// MasterConfiguration
	cfg.MasterConfiguration.KubernetesVersion = mpc.ComponentVersions.KubernetesVersion
	cfg.MasterConfiguration.Etcd.External = &ExternalEtcd{
		Endpoints: []string{"https://127.0.0.1:2379"},
		CAFile:    "/etc/etcd/pki/ca.crt",
		CertFile:  "/etc/etcd/pki/apiserver-etcd-client.crt",
		KeyFile:   "/etc/etcd/pki/apiserver-etcd-client.key",
	}
	if err := validateClusterNetworkingConfiguration(cluster); err != nil {
		return nil, fmt.Errorf("invalid cluster networking configuration: %v", err)
	}
	cfg.MasterConfiguration.Networking.DNSDomain = cluster.Spec.ClusterNetwork.ServiceDomain
	cfg.MasterConfiguration.Networking.PodSubnet = cluster.Spec.ClusterNetwork.Pods.CIDRBlocks[0]
	cfg.MasterConfiguration.Networking.ServiceSubnet = cluster.Spec.ClusterNetwork.Services.CIDRBlocks[0]

	// VIPConfiguration (optional)
	if cpc.VIPConfiguration != nil {
		cfg.VIPConfiguration = &VIPConfiguration{
			IP:               cpc.VIPConfiguration.IP,
			RouterID:         cpc.VIPConfiguration.RouterID,
			NetworkInterface: pm.Spec.VIPNetworkInterface,
		}

		cfg.MasterConfiguration.API.ControlPlaneEndpoint = cpc.VIPConfiguration.IP
		cfg.MasterConfiguration.APIServerCertSANs = []string{cpc.VIPConfiguration.IP}
	}

	// ClusterConfig (optional)
	if cpc.ClusterConfig != nil {
		setInitConfigFromClusterConfig(cfg, cpc.ClusterConfig)
	}

	return cfg, nil
}

// SetKubeAPIServerConfig sets configuration for API Server.
// Depending on the parameter name this function sets
// the MasterConfiguration fields or APIServerExtraArgs
func setKubeAPIServerConfig(cfg *InitConfiguration, clusterConfig *spv1.ClusterConfig) error {
	if clusterConfig.KubeAPIServer != nil {
		// Set fields for API server manually as there is no upstream type yet.
		// BindPort
		bindPortStr, ok := clusterConfig.KubeAPIServer[spconstants.KubeAPIServerSecurePortKey]
		if ok {
			bindPort, err := strconv.ParseInt(bindPortStr, 10, 32)
			if err != nil {
				return fmt.Errorf("unable to parse port value: %s", bindPortStr)
			}
			cfg.MasterConfiguration.API.BindPort = int32(bindPort)
			// delete as it should not be considered as an extra arg
			delete(clusterConfig.KubeAPIServer, spconstants.KubeAPIServerSecurePortKey)
		}
	}
	return nil
}

func setInitConfigFromClusterConfig(cfg *InitConfiguration, clusterConfig *spv1.ClusterConfig) error {
	if err := setKubeAPIServerConfig(cfg, clusterConfig); err != nil {
		return fmt.Errorf("unable to set configurable parameters for api-server: %v", err)
	}
	cfg.MasterConfiguration.ControllerManagerExtraArgs = clusterConfig.KubeControllerManager
	if clusterConfig.KubeProxy != nil {
		cfg.MasterConfiguration.KubeProxy.Config = clusterConfig.KubeProxy.DeepCopy()
	}
	cfg.MasterConfiguration.SchedulerExtraArgs = clusterConfig.KubeScheduler
	if clusterConfig.Kubelet != nil {
		cfg.MasterConfiguration.KubeletConfiguration.BaseConfig = clusterConfig.Kubelet.DeepCopy()
	}
	cfg.NetworkBackend = clusterConfig.NetworkBackend
	cfg.KeepAlived = clusterConfig.KeepAlived
	return nil
}

func JoinConfigurationForMachine(cluster *clusterv1.Cluster, machine *clusterv1.Machine, discoveryTokenAPIServers, discoveryTokenCACertHashes []string, token string) (*JoinConfiguration, error) {
	cfg := &JoinConfiguration{
		NodeConfiguration: KubeadmNodeConfiguration{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NodeConfiguration",
				APIVersion: "kubeadm.k8s.io/v1alpha2",
			},
		},
	}

	// Discovery and TLS bootstrap configuration
	cfg.NodeConfiguration.DiscoveryTokenAPIServers = discoveryTokenAPIServers
	cfg.NodeConfiguration.DiscoveryTokenCACertHashes = discoveryTokenCACertHashes
	cfg.NodeConfiguration.Token = token

	// NodeRegistrationOptions
	cfg.NodeConfiguration.NodeRegistration.Name = machine.Name
	cfg.NodeConfiguration.NodeRegistration.Taints = machine.Spec.Taints

	return cfg, nil
}

func validateClusterNetworkingConfiguration(cluster clusterv1.Cluster) error {
	switch cbl := len(cluster.Spec.ClusterNetwork.Pods.CIDRBlocks); {
	case cbl < 1:
		return fmt.Errorf("cluster %q spec.clusterNetwork.pods.cidrBlocks must contain at least one block", cluster.Name)
	case cbl > 1:
		return fmt.Errorf("cluster %q spec.clusterNetwork.pods.cidrBlocks must contain at most one block", cluster.Name)
	}
	switch cbl := len(cluster.Spec.ClusterNetwork.Services.CIDRBlocks); {
	case cbl < 1:
		return fmt.Errorf("cluster %q spec.clusterNetwork.pods.cidrBlocks must contain at least one block", cluster.Name)
	case cbl > 1:
		return fmt.Errorf("cluster %q spec.clusterNetwork.pods.cidrBlocks must contain at most one block", cluster.Name)
	}
	return nil
}
