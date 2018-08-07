package common

import "time"

const (
	DefaultApiserverPort               = 6443
	DrainTimeout                       = 5 * time.Minute
	DrainGracePeriodSeconds            = -1
	MasterRole                         = "master"
	NodeRole                           = "node"
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
	SystemUUIDFile                     = "/sys/class/dmi/id/product_uuid"
	KubectlFile                        = "/opt/bin/kubectl"
	AdminKubeconfig                    = "/etc/kubernetes/admin.conf"
	KubeletKubeconfig                  = "/etc/kubernetes/kubelet.conf"
	DefaultNodeadmVersion              = "v0.0.1-alpha"
	DefaultEtcdadmVersion              = "v0.0.1-alpha"
	DefaultKubernetesVersion           = "1.10.4"
	DefaultCNIVersion                  = "v0.6.0"
	DefaultFlannelVersion              = "v0.10.0"
	DefaultKeepalivedVersion           = "v2.0.4"
	DefaultEtcdVersion                 = "v3.3.8"
	ClusterV1PrintTemplate             = `Cluster Information
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
