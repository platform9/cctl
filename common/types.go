package common

import (
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type ExtraOpts struct {
	Masters     []string
	Nodes       []string
	Vip         string
	MetallbPool string
	K8sVersion  string
}

type ClusterState struct {
	SSHSecret []SSHSecret
	Cluster   clusterv1.Cluster   `yaml:"cluster"`
	Machines  []clusterv1.Machine `yaml:"machines"`
	Extra     ExtraOpts           `yaml:"extra"`
}

type SSHSecret struct {
	Name           string
	SSHCredentials SSHCredentials
}
type SSHCredentials struct {
	User       string
	PrivateKey string
}
