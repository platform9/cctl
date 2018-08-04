package common

import "time"

const (
	DefaultKubernetesVersion = "1.10.4"
	DefaultApiserverPort     = 6443
	DrainTimeout             = 5 * time.Minute
	DrainGracePeriodSeconds  = -1

	MasterRole = "master"
	NodeRole   = "node"

	DefaultSSHPort                     = 22
	DefaultNamespace                   = "default"
	DefaultClusterName                 = "cctl-cluster"
	DefaultSSHCredentialSecretName     = "ssh-credential"
	DefaultCommonCASecretName          = "common-ca"
	DefaultEtcdCASecretName            = "etcd-ca"
	DefaultAPIServerCASecretName       = "apiserver-ca"
	DefaultFrontProxyCASecretName      = "front-proxy-ca"
	DefaultServiceAccountKeySecretName = "serviceaccount-key"
	DefaultBootstrapTokenSecretName    = "bootstrap-token"

	SystemUUIDFile    = "/sys/class/dmi/id/product_uuid"
	KubectlFile       = "/opt/bin/kubectl"
	AdminKubeconfig   = "/etc/kubernetes/admin.conf"
	KubeletKubeconfig = "/etc/kubernetes/kubelet.conf"

	ClusterV1PrintTemplate = `Cluster Information
------- ------------
Cluster Name       : {{ .Cluster.ObjectMeta.Name}}
Creation Timestamp : {{ .Cluster.ObjectMeta.CreationTimestamp }}

Networking

	Pod CIDR     : {{ .Cluster.Spec.ClusterNetwork.Pods.CIDRBlocks }}
	Service CIDR : {{ .Cluster.Spec.ClusterNetwork.Services.CIDRBlocks }}
	VIP          : {{ .ClusterProviderSpec.VIPConfiguration.IP  }}
	RouterID     : {{ .ClusterProviderSpec.VIPConfiguration.RouterID }}
`
	MachineV1PrintTemplate = `Machine Information
------- -----------
Machine IP             Creation Timestamp                      Role
{{ range $machine := .}}{{ $machine.ObjectMeta.Name }}           {{ $machine.ObjectMeta.CreationTimestamp }}           {{ $machine.Spec.Roles }}
{{ end }}
`
)
