package sshprovider

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Secret.Data["username"] - username used for authentication
	CredentialSecretUsernameKey = "username"
	// Secret.Data["ssh-privatekey"] - private key needed for authentication
	CredentialSecretSSHPrivateKeyKey = "ssh-privatekey"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterSpec defines the desired provider-specific state of the
// cluster
type ClusterSpec struct {
	metav1.TypeMeta

	// EtcdCASecret is the name of the Secret with the etcd CA certificate and
	// private key. If it is not specified, the default name is derived from the
	// cluster name. If the Secret is not present, the provider generates a
	// self-signed one and creates the Secret.
	// +optional
	EtcdCASecret *corev1.LocalObjectReference
	// APIServerCASecret is the name of the Secret with the kube-apiserver CA
	// certificate and private key. If it is not specified, the default name is
	// derived from the cluster name. If the Secret is not present, the provider
	// generates a self-signed one and creates the Secret.
	// +optional
	APIServerCASecret *corev1.LocalObjectReference
	// FrontProxyCASecret is the name of the Secret with the front-proxy CA
	// certificate and private key. If it is not specified, the default name is
	// derived from the cluster name. If the Secret is not present, the provider
	// generates a self-signed one and creates the Secret.
	// +optional
	FrontProxyCASecret *corev1.LocalObjectReference
	// ServiceAccountKeySecret is the name of the Secret with the private and
	// public keys used to generate Service Account tokens. If it is not specified,
	// the default name is derived from the cluster name. If the Secret is not
	// present, the provider generates a self-signed one and creates the Secret.
	// +optional
	ServiceAccountKeySecret *corev1.LocalObjectReference
	// VIPConfiguration is the configuration of the VIP for the API. If it is not
	// specified, the VIP is not created.
	// +optional
	BootstrapTokenSecret *corev1.LocalObjectReference
	// VIPConfiguration is the configuration of the VIP for the API. If it is not
	// specified, the VIP is not created.
	// +optional
	VIPConfiguration *VIPConfiguration
	// ClusterConfig is the set of configurable parameters for the cluster.
	// If not provided, the parameters are set to their default values.
	ClusterConfig *ClusterConfig `json:"clusterConfig,omitempty"`
}

type ClusterConfig struct {
	// generic map[string]string types would eventually be replaced by
	// corresponding structured types as they become available upstream
	KubeAPIServer         map[string]string
	KubeDNS               map[string]string
	KubeControllerManager map[string]string
	KubeScheduler         map[string]string
	KubeProxy             *KubeProxyConfiguration
	Kubelet               *KubeletConfiguration
	NetworkBackend        map[string]string
	KeepAlived            map[string]string
}

// VIPConfiguration specifies the parameters used to provision a virtual IP
// which API servers advertise and accept requests on.
type VIPConfiguration struct {
	// The virtual IP.
	IP string
	// The virtual router ID. Must be in the range [0, 254]. Must be unique within
	// a single L2 network domain.
	RouterID int
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterStatus defines the observed provider-specific state of the
// cluster
type ClusterStatus struct {
	metav1.TypeMeta

	// EtcdMembers defines the observed etcd configuration of the cluster.
	EtcdMembers []EtcdMember
}

// MachineComponentVersions
type MachineComponentVersions struct {
	NodeadmVersion    string
	EtcdadmVersion    string
	KubernetesVersion string
	CNIVersion        string
	FlannelVersion    string
	KeepalivedVersion string
	EtcdVersion       string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineSpec
type MachineSpec struct {
	metav1.TypeMeta

	// A list of roles for this Machine to use.
	Roles []MachineRole

	// ProvisionedMachineName is the binding reference to the Provisioned
	// Machine backing this Machine.
	ProvisionedMachineName string
	// ComponentVersions enumerates versions of all the components
	ComponentVersions *MachineComponentVersions
}

// The MachineRole indicates the purpose of the Machine, and will determine
// what software and configuration will be used when provisioning and managing
// the Machine. A single Machine may have more than one role, and the list and
// definitions of supported roles is expected to evolve over time.
//
// Currently, only two roles are supported: Master and Node. In the future, we
// expect user needs to drive the evolution and granularity of these roles,
// with new additions accommodating common cluster patterns, like dedicated
// etcd Machines.
//
//                 +-----------------------+------------------------+
//                 | Master present        | Master absent          |
// +---------------+-----------------------+------------------------|
// | Node present: | Install control plane | Join the cluster as    |
// |               | and be schedulable    | just a node            |
// |---------------+-----------------------+------------------------|
// | Node absent:  | Install control plane | Invalid configuration  |
// |               | and be unschedulable  |                        |
// +---------------+-----------------------+------------------------+
type MachineRole string

const (
	MasterRole MachineRole = "Master"
	NodeRole   MachineRole = "Node"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineStatus
type MachineStatus struct {
	metav1.TypeMeta

	// SSHConfig is the configuration used to SSH to the machine.
	// +optional
	SSHConfig *SSHConfig

	// Network interface used to create the virtual IP.
	// This field is populated for masters only.
	// +optional
	VIPNetworkInterface string

	// EtcdMember defines the observed etcd configuration of the machine.
	// This field is populated for masters only.
	// +optional
	EtcdMember *EtcdMember
}

// SSHConfig specifies everything needed to ssh to a host
type SSHConfig struct {
	// The IP or hostname used to SSH to the machine
	Host string
	// The used to SSH to the machine
	Port int
	// The SSH public keys of the machine
	PublicKeys []string
	// The Secret with the username and private key used to SSH to the machine
	CredentialSecret corev1.LocalObjectReference
}

// EtcdMember defines the configuration of an etcd member.
type EtcdMember struct {
	// ID is the member ID for this member.
	ID uint64
	// Name is the human-readable name of the member.
	Name string
	// PeerURLs is the list of URLs the member exposes to the cluster for communication.
	PeerURLs []string
	// ClientURLs is the list of URLs the member exposes to clients for communication.
	ClientURLs []string
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProvisionedMachine describes a machine provisioned to accept SSH requests.
type ProvisionedMachine struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ProvisionedMachineSpec
	Status ProvisionedMachineStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProvisionedMachineList is a list of ProvisionedMachines.
type ProvisionedMachineList struct {
	metav1.TypeMeta
	// +optional
	metav1.ListMeta

	Items []ProvisionedMachine
}

// ProvisionedMachineSpec defines the desired state of ProvisionedMachine
type ProvisionedMachineSpec struct {
	// SSHConfig specifies everything needed to ssh to a host
	SSHConfig *SSHConfig
	// Network interface chosen to create the virtual IP. If it is not
	// specified, the interface of the default gateway is chosen.
	// +optional
	VIPNetworkInterface string
}

// ProvisionedMachineStatus defines the observed state of ProvisionedMachine
type ProvisionedMachineStatus struct {
	// MachineRef is part of a bi-directional binding between
	// ProvisionedMachine and Machine.
	// MachineRef is expected to be non-nil when bound.
	// provisionedmachine.MachineRef is the authoritative bind between
	// ProvisionedMachine and Machine.
	// When set to non-nil value, Machine.Spec.Selector of the referenced
	// Machine is ignored, i.e. labels of this ProvisionedMachine do not
	// need to match Machine selector.
	// +optional
	MachineRef *corev1.LocalObjectReference
}
