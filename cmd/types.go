package cmd

import (
	sshconfig "github.com/platform9/ssh-provider/sshproviderconfig"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type ClusterSpec struct {
	Name           string `yaml:"name"`
	ServiceNetwork string `yaml:"serviceNetwork"`
	PodNetwork     string `yaml:"podNetwork"`
	Vip            string `yaml:"vip"`
	Cacert         string `yaml:"cacert"`
	Cakey          string `yaml:"cakey"`
	Token          string `yaml:"token"`
	Version        string `yaml:"version"`
}

type ClusterState struct {
	SSHConfig sshconfig.SSHMachineProviderConfig
	Cluster   clusterv1.Cluster
	Machines  []clusterv1.Machine
}
