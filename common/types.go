package common

import (
	sshProvider "github.com/platform9/ssh-provider/provisionedmachine"
	sshproviderconfigv1 "github.com/platform9/ssh-provider/sshproviderconfig/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type ClusterState struct {
	SSHCredentials      *corev1.Secret                        `yaml:"sshCredentials"`
	EtcdCA              *corev1.Secret                        `yaml:"etcdCA"`
	APIServerCA         *corev1.Secret                        `yaml:"apiServerCA"`
	FrontProxyCA        *corev1.Secret                        `yaml:"frontProxyCA"`
	ServiceAccountKey   *corev1.Secret                        `yaml:"serviceAccountKey"`
	Cluster             clusterv1.Cluster                     `yaml:"cluster"`
	Machines            []clusterv1.Machine                   `yaml:"machines"`
	VIPConfiguration    *sshproviderconfigv1.VIPConfiguration `yaml:"vipConfiguration"`
	K8sVersion          string                                `yaml:"k8sVersion"`
	ProvisionedMachines []sshProvider.ProvisionedMachine      `yaml:"provisionedMachines"`
}

type VIPConfigurationType struct {
	// The virtual IP.
	IP string `json:"ip"`
	// The virtual router ID. Must be in the range [0, 254]. Must be unique within
	// a single L2 network domain.
	RouterID int `json:"routerID"`
}

type SSHSecret struct {
	User       string
	PrivateKey string
}
