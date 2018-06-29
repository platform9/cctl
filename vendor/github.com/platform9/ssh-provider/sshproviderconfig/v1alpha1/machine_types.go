/*
Copyright 2018 Platform 9 Systems, Inc.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeletconfigv1alpha1 "k8s.io/kubernetes/pkg/kubelet/apis/kubeletconfig/v1alpha1"
	kubeproxyconfigv1alpha1 "k8s.io/kubernetes/pkg/proxy/apis/kubeproxyconfig/v1alpha1"
)

// SSHMachineProviderConfig defines the desired provider-specific state of the
// machine.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SSHMachineProviderConfig struct {
	metav1.TypeMeta `json:",inline"`

	// KubeletConfiguration is the kubelet configuration.
	// +optional
	KubeletConfiguration *kubeletconfigv1alpha1.KubeletConfiguration

	// KubeProxyConfiguration is the kube-proxy configuration
	// +optional
	KubeProxyConfiguration *kubeproxyconfigv1alpha1.KubeProxyConfiguration
}

// SSHMachineProviderStatus defines the observed provider-specific state of the
// machine.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SSHMachineProviderStatus struct {
	metav1.TypeMeta `json:",inline"`

	// SSHConfig is the configuration used to SSH to the machine.
	// +optional
	SSHConfig *SSHConfig `json:"sshConfig"`

	// EtcdMember defines the observed etcd configuration of the machine.
	// This field is populated for masters only.
	// +optional
	EtcdMember *EtcdMember `json:"etcdMember,omitempty"`
}

// SSHConfig specifies everything needed to ssh to a host
type SSHConfig struct {
	// The IP or hostname used to SSH to the machine
	Host string `json:"host"`
	// The used to SSH to the machine
	Port int `json:"port"`
	// The SSH public keys of the machine
	PublicKeys []string `json:"publicKeys"`
	// The Secret with the username and private key used to SSH to the machine
	SecretName string `json:"secretName"`
}
