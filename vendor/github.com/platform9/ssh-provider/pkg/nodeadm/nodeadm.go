package nodeadm

import (
	"fmt"

	spv1 "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	"github.com/platform9/ssh-provider/pkg/controller"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type InitConfiguration struct {
	MasterConfiguration KubeadmInitConfiguration `json:"masterConfiguration,omitempty"`
	Networking          Networking               `json:"networking,omitempty"`
	VIPConfiguration    VIPConfiguration         `json:"vipConfiguration,omitempty"`
	ClusterConfig       spv1.ClusterConfig       `json:"clusterConfiguration,omitempty"`
}

type JoinConfiguration struct {
	Networking    Networking         `json:"networking,omitempty"`
	ClusterConfig spv1.ClusterConfig `json:"clusterConfiguration,omitempty"`
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

// Networking contains elements describing cluster's networking configuration
type Networking struct {
	// ServiceSubnet is the subnet used by k8s services. Defaults to "10.96.0.0/12".
	ServiceSubnet string `json:"serviceSubnet,omitempty"`
	// PodSubnet is the subnet used by pods.
	PodSubnet string `json:"podSubnet,omitempty"`
	// DNSDomain is the dns domain used by k8s services. Defaults to "cluster.local".
	DNSDomain string `json:"dnsDomain,omitempty"`
}
type KubeadmInitConfiguration struct {
	API               API        `json:"api,omitempty"`
	APIServerCertSANs []string   `json:"apiServerCertSANs,omitempty"`
	Etcd              Etcd       `json:"etcd,omitempty"`
	KubernetesVersion string     `json:"kubernetesVersion,omitempty"`
	Networking        Networking `json:"networking,omitempty"`
}

type API struct {
	AdvertiseAddress     string `json:"advertiseAddress,omitempty"`
	BindPort             int32  `json:"bindPort,omitempty"`
	ControlPlaneEndpoint string `json:"controlPlaneEndpoint"`
}

type Etcd struct {
	Endpoints []string `json:"endpoints,omitempty"`
	CAFile    string   `json:"caFile,omitempty"`
	CertFile  string   `json:"certFile,omitempty"`
	KeyFile   string   `json:"keyFile,omitempty"`
}

func InitConfigurationForMachine(cluster clusterv1.Cluster, machine clusterv1.Machine, pm spv1.ProvisionedMachine) (*InitConfiguration, error) {
	cfg := &InitConfiguration{}

	cpc, err := controller.GetClusterSpec(cluster)
	if err != nil {
		return nil, fmt.Errorf("unable to decode cluster spec: %v", err)
	}

	// MasterConfiguration
	if cpc.VIPConfiguration != nil {
		cfg.MasterConfiguration.API.ControlPlaneEndpoint = cpc.VIPConfiguration.IP
		cfg.MasterConfiguration.APIServerCertSANs = []string{cpc.VIPConfiguration.IP}
	} // else: kubeadm will set defaults
	cfg.MasterConfiguration.KubernetesVersion = machine.Spec.Versions.ControlPlane
	cfg.MasterConfiguration.Etcd.Endpoints = []string{"https://127.0.0.1:2379"}
	cfg.MasterConfiguration.Etcd.CAFile = "/etc/etcd/pki/ca.crt"
	cfg.MasterConfiguration.Etcd.CertFile = "/etc/etcd/pki/apiserver-etcd-client.crt"
	cfg.MasterConfiguration.Etcd.KeyFile = "/etc/etcd/pki/apiserver-etcd-client.key"
	if cpc.ClusterConfig != nil {
		cfg.ClusterConfig = *cpc.ClusterConfig
	} else {
		cfg.ClusterConfig = spv1.ClusterConfig{}
	}
	// Networking
	switch len(cluster.Spec.ClusterNetwork.Pods.CIDRBlocks) {
	case 0:
		return nil, fmt.Errorf("cluster %q spec.clusterNetwork.pods.cidrBlocks must contain at least one block", cluster.Name)
	case 1:
		cfg.Networking.PodSubnet = cluster.Spec.ClusterNetwork.Pods.CIDRBlocks[0]
	case 2:
		return nil, fmt.Errorf("cluster %q spec.clusterNetwork.pods.cidrBlocks must contain at most one block", cluster.Name)
	}
	switch len(cluster.Spec.ClusterNetwork.Services.CIDRBlocks) {
	case 0:
		return nil, fmt.Errorf("cluster %q spec.clusterNetwork.pods.cidrBlocks must contain at least one block", cluster.Name)
	case 1:
		cfg.Networking.ServiceSubnet = cluster.Spec.ClusterNetwork.Services.CIDRBlocks[0]
	case 2:
		return nil, fmt.Errorf("cluster %q spec.clusterNetwork.pods.cidrBlocks must contain at most one block", cluster.Name)
	}
	cfg.Networking.DNSDomain = cluster.Spec.ClusterNetwork.ServiceDomain

	// VIPConfiguration
	if cpc.VIPConfiguration != nil {
		cfg.VIPConfiguration.IP = cpc.VIPConfiguration.IP
		cfg.VIPConfiguration.RouterID = cpc.VIPConfiguration.RouterID
	}
	cfg.VIPConfiguration.NetworkInterface = pm.Spec.VIPNetworkInterface

	return cfg, nil
}

func JoinConfigurationForMachine(cluster *clusterv1.Cluster, machine *clusterv1.Machine) (*JoinConfiguration, error) {
	cfg := &JoinConfiguration{}

	cpc, err := controller.GetClusterSpec(*cluster)
	if err != nil {
		return nil, fmt.Errorf("unable to decode cluster spec: %v", err)
	}

	// Networking
	switch len(cluster.Spec.ClusterNetwork.Pods.CIDRBlocks) {
	case 0:
		return nil, fmt.Errorf("cluster %q spec.clusterNetwork.pods.cidrBlocks must contain at least one block", cluster.Name)
	case 1:
		cfg.Networking.PodSubnet = cluster.Spec.ClusterNetwork.Pods.CIDRBlocks[0]
	default:
		return nil, fmt.Errorf("cluster %q spec.clusterNetwork.pods.cidrBlocks must contain at most one block", cluster.Name)
	}
	switch len(cluster.Spec.ClusterNetwork.Services.CIDRBlocks) {
	case 0:
		return nil, fmt.Errorf("cluster %q spec.clusterNetwork.pods.cidrBlocks must contain at least one block", cluster.Name)
	case 1:
		cfg.Networking.ServiceSubnet = cluster.Spec.ClusterNetwork.Services.CIDRBlocks[0]
	default:
		return nil, fmt.Errorf("cluster %q spec.clusterNetwork.pods.cidrBlocks must contain at most one block", cluster.Name)
	}
	cfg.Networking.DNSDomain = cluster.Spec.ClusterNetwork.ServiceDomain
	if cpc.ClusterConfig != nil {
		cfg.ClusterConfig = *cpc.ClusterConfig
	} else {
		cfg.ClusterConfig = spv1.ClusterConfig{}
	}
	return cfg, nil
}
