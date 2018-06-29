package provisionedmachine

import (
	"fmt"

	"github.com/ghodss/yaml"
	sshproviderv1 "github.com/platform9/ssh-provider/sshproviderconfig/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const configMapKey = `provisionedMachine`

// ProvisionedMachine describes a machine provisioned to accept SSH requests.
type ProvisionedMachine struct {
	// SSHConfig specifies everything needed to ssh to a host
	SSHConfig *sshproviderv1.SSHConfig `json:"sshConfig"`
	// Network interface chosen to create the virtual IP. If it is not specified,
	// the interface of the default gateway is chosen.
	// +optional
	VIPNetworkInterface string `json:"vipNetworkInterface,omitEmpty"`
}

// NewFromConfigMap creates a ProvisionedMachine from a ConfigMap
func NewFromConfigMap(cm *corev1.ConfigMap) (*ProvisionedMachine, error) {
	pmString, ok := cm.Data[configMapKey]
	if !ok {
		return nil, fmt.Errorf("did not find %q key in ConfigMap", configMapKey)
	}
	pm := &ProvisionedMachine{}
	err := yaml.Unmarshal([]byte(pmString), pm)
	if err != nil {
		return nil, err
	}
	return pm, nil
}

// ToConfigMap writes the ProvisionedMachine to a ConfigMap
func (pm *ProvisionedMachine) ToConfigMap(cm *corev1.ConfigMap) error {
	bytes, err := yaml.Marshal(&pm)
	if err != nil {
		return err
	}
	cm.Data[configMapKey] = string(bytes)
	return nil
}
