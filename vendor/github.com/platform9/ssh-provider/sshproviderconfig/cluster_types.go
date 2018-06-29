/*
Copyright 2018 Platform 9 Systems, Inc.
*/

package sshproviderconfig

import (
	"net"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SSHClusterProviderConfig defines the desired provider-specific state of the
// cluster
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SSHClusterProviderConfig struct {
	metav1.TypeMeta `json:",inline"`

	// APIServerCASecret is the name of the Secret with the kube-apiserver CA
	// certificate and private key. If it is not specified, the default name is
	// derived from the cluster name. If the Secret is not present, the provider
	// generates a self-signed one and creates the Secret.
	// +optional
	APIServerCASecret string `json:"apiServerCASecret"`
	// FrontProxyCASecret is the name of the Secret with the front-proxy CA
	// certificate and private key. If it is not specified, the default name is
	// derived from the cluster name. If the Secret is not present, the provider
	// generates a self-signed one and creates the Secret.
	// +optional
	FrontProxyCASecret string `json:"frontProxyCASecret"`
	// ServiceAccountKeySecret is the name of the Secret with the private and
	// public keys used to generate Service Account tokens. If it is not specified,
	// the default name is derived from the cluster name. If the Secret is not
	// present, the provider generates a self-signed one and creates the Secret.
	// +optional
	ServiceAccountKeySecret string `json:"serviceAccountKeySecret"`
	// VIPConfiguration is the configuration of the VIP for the API. If it is not
	// specified, the VIP is not created.
	// +optional
	VIPConfiguration *VIPConfiguration `json:"vipConfiguration,omitempty"`
}

// VIPConfiguration specifies the parameters used to provision a virtual IP
// which API servers advertise and accept requests on.
type VIPConfiguration struct {
	// The virtual IP.
	IP net.IP `json:"ip"`
	// The virtual router ID. Must be in the range [0, 254]. Must be unique within
	// a single L2 network domain.
	RouterID int `json:"routerID"`
}

// SSHClusterProviderStatus defines the observed provider-specific state of the
// cluster
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SSHClusterProviderStatus struct {
	metav1.TypeMeta `json:",inline"`

	// EtcdMembers defines the observed etcd configuration of the cluster.
	EtcdMembers []EtcdMember `json:"etcdMembers,omitempty"`
}
