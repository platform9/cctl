package nodeadm

import (
	spv1 "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Networking contains elements describing cluster's networking configuration
type Networking struct {
	// ServiceSubnet is the subnet used by k8s services. Defaults to "10.96.0.0/12".
	ServiceSubnet string `json:"serviceSubnet,omitempty"`
	// PodSubnet is the subnet used by pods.
	PodSubnet string `json:"podSubnet,omitempty"`
	// DNSDomain is the dns domain used by k8s services. Defaults to "cluster.local".
	DNSDomain string `json:"dnsDomain,omitempty"`
}

type KubeadmMasterConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	API                        API                  `json:"api,omitempty"`
	APIServerCertSANs          []string             `json:"apiServerCertSANs,omitempty"`
	Etcd                       Etcd                 `json:"etcd,omitempty"`
	KubernetesVersion          string               `json:"kubernetesVersion,omitempty"`
	Networking                 Networking           `json:"networking,omitempty"`
	KubeletConfiguration       KubeletConfiguration `json:"kubeletConfiguration,omitempty"`
	KubeProxy                  KubeProxy            `json:"kubeProxy,omitempty"`
	APIServerExtraArgs         map[string]string    `json:"apiServerExtraArgs,omitempty"`
	ControllerManagerExtraArgs map[string]string    `json:"controllerManagerExtraArgs,omitempty"`
	SchedulerExtraArgs         map[string]string    `json:"schedulerExtraArgs,omitempty"`

	NodeRegistration NodeRegistrationOptions `json:"nodeRegistration"`
}

type KubeadmNodeConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// Token is used for both discovery and TLS bootstrapping.
	Token string `json:"token"`

	// DiscoveryTokenAPIServers is a set of IPs to API servers from which info
	// will be fetched. Currently we only pay attention to one API server but
	// hope to support >1 in the future.
	DiscoveryTokenAPIServers []string `json:"discoveryTokenAPIServers,omitempty"`

	// DiscoveryTokenCACertHashes specifies a set of public key pins to verify
	// when token-based discovery is used. The root CA found during discovery
	// must match one of these values. Specifying an empty set disables root CA
	// pinning, which can be unsafe. Each hash is specified as "<type>:<value>",
	// where the only currently supported type is "sha256". This is a hex-encoded
	// SHA-256 hash of the Subject Public Key Info (SPKI) object in DER-encoded
	// ASN.1. These hashes can be calculated using, for example, OpenSSL:
	// openssl x509 -pubkey -in ca.crt openssl rsa -pubin -outform der 2>&/dev/null | openssl dgst -sha256 -hex
	DiscoveryTokenCACertHashes []string `json:"discoveryTokenCACertHashes,omitempty"`

	// NodeRegistration holds fields that relate to registering the new master node to the cluster
	NodeRegistration NodeRegistrationOptions `json:"nodeRegistration"`
}

// NodeRegistrationOptions holds fields that relate to registering a new master or node to the cluster, either via "kubeadm init" or "kubeadm join"
type NodeRegistrationOptions struct {
	// Name is the `.Metadata.Name` field of the Node API object that will be created in this `kubeadm init` or `kubeadm joiń` operation.
	// This field is also used in the CommonName field of the kubelet's client certificate to the API server.
	// Defaults to the hostname of the node if not provided.
	Name string `json:"name,omitempty"`

	// Taints specifies the taints the Node API object should be registered with. If this field is unset, i.e. nil, in the `kubeadm init` process
	// it will be defaulted to []v1.Taint{'node-role.kubernetes.io/master=""'}. If you don't want to taint your master node, set this field to an
	// empty slice, i.e. `taints: {}` in the YAML file. This field is solely used for Node registration.
	Taints []v1.Taint `json:"taints,omitempty"`

	// KubeletExtraArgs passes through extra arguments to the kubelet. The arguments here are passed to the kubelet command line via the environment file
	// kubeadm writes at runtime for the kubelet to source. This overrides the generic base-level configuration in the kubelet-config-1.X ConfigMap
	// Flags have higher higher priority when parsing. These values are local and specific to the node kubeadm is executing on.
	KubeletExtraArgs map[string]string `json:"kubeletExtraArgs,omitempty"`
}

type API struct {
	AdvertiseAddress     string `json:"advertiseAddress,omitempty"`
	BindPort             int32  `json:"bindPort,omitempty"`
	ControlPlaneEndpoint string `json:"controlPlaneEndpoint"`
}

// Etcd contains elements describing Etcd configuration.
type Etcd struct {
	// External describes how to connect to an external etcd cluster
	// Local and External are mutually exclusive
	External *ExternalEtcd `json:"external,omitempty"`
}

// ExternalEtcd describes an external etcd cluster
type ExternalEtcd struct {

	// Endpoints of etcd members. Useful for using external etcd.
	// If not provided, kubeadm will run etcd in a static pod.
	Endpoints []string `json:"endpoints"`
	// CAFile is an SSL Certificate Authority file used to secure etcd communication.
	CAFile string `json:"caFile"`
	// CertFile is an SSL certification file used to secure etcd communication.
	CertFile string `json:"certFile"`
	// KeyFile is an SSL key file used to secure etcd communication.
	KeyFile string `json:"keyFile"`
}

// KubeletConfiguration contains elements describing initial remote configuration of kubelet.
type KubeletConfiguration struct {
	BaseConfig *spv1.KubeletConfiguration `json:"baseConfig,omitempty"`
}

// KubeProxy contains elements describing the proxy configuration.
type KubeProxy struct {
	Config *spv1.KubeProxyConfiguration `json:"config,omitempty"`
}
