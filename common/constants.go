package common

import "time"

const (
	K8S_VERSION                = "1.10.4"
	DEFAULT_APISERVER_PORT     = 6443
	DRAIN_TIMEOUT              = 5 * time.Minute
	DRAIN_GRACE_PERIOD_SECONDS = -1

	ClusterV1PrintTemplate = `Cluster Information
------- ------------
Cluster Name       : {{ .Cluster.ObjectMeta.Name}}
Creation Timestamp : {{ .Cluster.ObjectMeta.CreationTimestamp }}
Kubernetes Version : {{ .K8sVersion }}

Networking

	Pod CIDR     : {{ .Cluster.Spec.ClusterNetwork.Pods.CIDRBlocks }}
	Service CIDR : {{ .Cluster.Spec.ClusterNetwork.Services.CIDRBlocks }}
	VIP          : {{ .VIPConfiguration.IP  }}
	RouterID     : {{ .VIPConfiguration.RouterID }}
`
	MachineV1PrintTemplate = `Machine Information
------- -----------
Machine IP             Creation Timestamp                      Role
{{ range $machine := .}}{{ $machine.ObjectMeta.Name }}           {{ $machine.ObjectMeta.CreationTimestamp }}           {{ $machine.Spec.Roles }}
{{ end }}
`
)
