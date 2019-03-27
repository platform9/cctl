/*
Copyright 2019 The cctl authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	DefaultNodeadmVersion               = "v0.2.0"
	DefaultEtcdadmVersion               = "v0.1.1"
	DefaultKubernetesVersion            = "1.11.7"
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
	KubeControllerMgrPodEvictionTimeout = "20s"
	DashcamBundleBaseDir                = "/var/tmp"
	DashcamCommandPath                  = "/opt/bin/dashcam"
	SupportBundleFileNamePrefix         = "cctl-bundle"
	ClusterV1PrintTemplate              = `Cluster Information
------- ------------
Cluster Name       : {{ .Cluster.ObjectMeta.Name}}
Creation Timestamp : {{ .Cluster.ObjectMeta.CreationTimestamp }}

Networking

	Pod CIDR     : {{ .Cluster.Spec.ClusterNetwork.Pods.CIDRBlocks }}
	Service CIDR : {{ .Cluster.Spec.ClusterNetwork.Services.CIDRBlocks }}
	{{  if .ClusterProviderSpec.VIPConfiguration  }}
	VIP          : {{ .ClusterProviderSpec.VIPConfiguration.IP  }}
	RouterID     : {{ .ClusterProviderSpec.VIPConfiguration.RouterID }}
	{{- else  }}
	VIP          : None configured.
	{{  end  }}
`
	MachineV1PrintTemplate = `Machine Information
------- -----------
Machine IP             Creation Timestamp                      Role
{{ range $machine := .}}{{ $machine.ObjectMeta.Name }}           {{ $machine.ObjectMeta.CreationTimestamp }}           {{ $machine.Spec.Roles }}
{{ end }}
`
	// LabelNodeRoleMaster specifies that a node is a master
	LabelNodeRoleMaster = "node-role.kubernetes.io/master"
)

var (
	// TODO(dlipovetsky) Move fields to configuration
	KubeletFailSwapOn   = false
	KubeletMaxPods      = int32(500)
	KubeletKubeAPIQPS   = int32(20)
	KubeletKubeAPIBurst = int32(40)
	KubeletEvictionHard = map[string]string{
		"memory.available": "600Mi",
		"nodefs.available": "10%",
	}
	KubeletFeatureGates = map[string]bool{
		"PodPriority": true,
	}
	DefaultKubeAPIServerExtraArgs         = map[string]string{}
	DefaultKubeControllerManagerExtraArgs = map[string]string{}
	DefaultKubeSchedulerExtraArgs         = map[string]string{}
)
var MasterComponents = []string{KubeAPIServer, KubeControllerManager, KubeScheduler}
