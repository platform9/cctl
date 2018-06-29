package machine

import (
	"fmt"

	"github.com/platform9/ssh-provider/provisionedmachine"
	sshconfigv1 "github.com/platform9/ssh-provider/sshproviderconfig/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

const (
	AnnotationMachineName            = "sshprovider.platform9.com/machine-name"
	AnnotationProvisionedMachineName = "sshprovider.platform9.com/provisionedmachine-name"
)

func (sa *SSHActuator) ReserveProvisionedMachine(machine *clusterv1.Machine) (*corev1.ConfigMap, error) {
	available := availableProvisionedMachines(machine, sa.provisionedMachineConfigMaps)
	if len(available) == 0 {
		return nil, fmt.Errorf("no available provisioned machines")
	}
	compatible := compatibleProvisionedMachines(machine, available)
	if len(compatible) == 0 {
		return nil, fmt.Errorf("no compatible provisioned machines")
	}
	reserved := compatible[0]
	linkProvisionedMachineWithMachine(reserved, machine)
	if err := sa.updateMachineStatus(reserved, machine); err != nil {
		return nil, fmt.Errorf("error updating machine status: %s", err)
	}

	return reserved, nil
}

func availableProvisionedMachines(machine *clusterv1.Machine, cms []*corev1.ConfigMap) []*corev1.ConfigMap {
	available := make([]*corev1.ConfigMap, 0)
	for _, cm := range cms {
		if cm.Annotations != nil {
			if _, ok := cm.Annotations[AnnotationMachineName]; ok {
				// This ProvisionedMachine is in use by a Machine
				continue
			}
		}
		available = append(available, cm)
	}
	return available
}

// TODO(dlipovetsky) Implement
func compatibleProvisionedMachines(machine *clusterv1.Machine, cms []*corev1.ConfigMap) []*corev1.ConfigMap {
	return cms
}

func (sa *SSHActuator) updateMachineStatus(cm *corev1.ConfigMap, machine *clusterv1.Machine) error {
	pm, err := provisionedmachine.NewFromConfigMap(cm)
	if err != nil {
		return fmt.Errorf("error parsing ProvisionedMachine from ConfigMap %q: %s", cm.Name, err)
	}
	sshProviderStatus := &sshconfigv1.SSHMachineProviderStatus{
		SSHConfig: pm.SSHConfig,
	}
	providerStatus, err := sa.sshProviderCodec.EncodeToProviderStatus(sshProviderStatus)
	if err != nil {
		return fmt.Errorf("error creating machine ProviderStatus: %s", err)
	}
	machine.Status.ProviderStatus = *providerStatus
	return nil
}

func linkProvisionedMachineWithMachine(cm *corev1.ConfigMap, machine *clusterv1.Machine) {
	if cm.Annotations == nil {
		cm.Annotations = map[string]string{}
	}
	cm.Annotations[AnnotationMachineName] = machine.Name

	if machine.Annotations == nil {
		machine.Annotations = map[string]string{}
	}
	machine.Annotations[AnnotationProvisionedMachineName] = cm.Name
}
