package common

import "time"

const (
	DefaultAPIServerPort                = 6443
	DrainTimeout                        = 5 * time.Minute
	DrainGracePeriodSeconds             = -1
	DrainDeleteLocalData                = false
	DrainForce                          = false
	MasterRole                          = "master"
	NodeRole                            = "node"
	DefaultSSHPort                      = 22
	DefaultNamespace                    = "default"
	DefaultClusterName                  = "cctl-cluster"
	DefaultSSHCredentialSecretName      = "ssh-credential"
	DefaultCommonCASecretName           = "common-ca"
	DefaultEtcdCASecretName             = "etcd-ca"
	DefaultAPIServerCASecretName        = "apiserver-ca"
	DefaultFrontProxyCASecretName       = "front-proxy-ca"
	DefaultServiceAccountKeySecretName  = "serviceaccount-key"
	DefaultBootstrapTokenSecretName     = "bootstrap-token"
	SystemUUIDFile                      = "/sys/class/dmi/id/product_uuid"
	KubectlFile                         = "/opt/bin/kubectl"
	AdminKubeconfig                     = "/etc/kubernetes/admin.conf"
	KubeletKubeconfig                   = "/etc/kubernetes/kubelet.conf"
	DefaultNodeadmVersion               = "v0.0.5"
	DefaultEtcdadmVersion               = "v0.0.6"
	DefaultKubernetesVersion            = "1.10.4"
	DefaultCNIVersion                   = "v0.6.0"
	DefaultFlannelVersion               = "v0.10.0"
	DefaultKeepalivedVersion            = "v2.0.4"
	DefaultEtcdVersion                  = "v3.3.8"
	DockerKubeAPIServerNameFilter       = "name=k8s_kube-apiserver.*kube-system.*"
	DockerRunningStatusFilter           = "status=running"
	InstanceStatusAnnotationKey         = "instance-status"
	KubeAPIServer                       = "kube-apiserver"
	KubeControllerManager               = "kube-controller-manager"
	KubeScheduler                       = "kube-scheduler"
	KubeSystemNamespace                 = "kube-system"
	MinimumControlPlaneVersion          = "v1.10.0"
	TmpKubeConfigNamePrefix             = "kubeconfig"
	DefaultAdminConfigSecretName        = "admin-kubeconfig"
	DefaultAdminConfigSecretKey         = "data"
	KubeAPIServerServiceNodePortRange   = "80-32767"
	KubeAPIServerAllowPrivileged        = "true"
	KubeControllerMgrPodEvictionTimeout = "20s"
	DashcamCommandPath                  = "/opt/bin/dashcam"
	SupportBundleFileNamePrefix         = "support-bundle"
	ClusterV1PrintTemplate              = `Cluster Information
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

var (
	// TODO(dlipovetsky) Move fields to configuration
	KubeletFailSwapOn   = false
	KubeletMaxPods      = int32(500)
	KubeletKubeAPIQPS   = int32(20)
	KubeletKubeAPIBurst = int32(40)
)
var MasterComponents = []string{KubeAPIServer, KubeControllerManager, KubeScheduler}
