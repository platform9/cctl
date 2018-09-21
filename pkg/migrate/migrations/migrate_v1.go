package migrations

import (
	"fmt"
	"github.com/platform9/cctl/pkg/state"
	spv1 "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/ghodss/yaml"
	spclient "github.com/platform9/ssh-provider/pkg/client/clientset_generated/clientset"
	"k8s.io/client-go/kubernetes"
	clusterclient "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

type StateV0toV1 struct {
	SchemaVersion state.SchemaVersion     `json:"schemaVersion"`
	Filename      string                  `json:"-"`
	KubeClient    kubernetes.Interface    `json:"-"`
	ClusterClient clusterclient.Interface `json:"-"`
	SPClient      spclient.Interface      `json:"-"`

	SecretList             corev1.SecretList           `json:"secretList,omitempty"`
	ClusterList            clusterv1.ClusterList       `json:"clusterList,omitempty"`
	MachineList            clusterv1.MachineList       `json:"machineList,omitempty"`
	ProvisionedMachineList spv1.ProvisionedMachineList `json:"provisionedMachineList,omitempty"`
}

// MigrateV0toV1 adds a schemaVersion field to the state file
func MigrateV0toV1(stateBytes *[]byte, version state.SchemaVersion) (interface{}, error) {
	tempState := new(StateV0toV1)
	if err := yaml.Unmarshal(*stateBytes, tempState); err != nil {
		return tempState, fmt.Errorf("unable to unmarshal state from YAML: %v", err)
	}
	switch tempState.SchemaVersion {
	case 0:
		tempState.SchemaVersion = 1
	case 1:
		return tempState, nil
	default:
		return tempState, fmt.Errorf("unable to migrate state file to schemaVersion 1")
	}
	return tempState, nil
}
