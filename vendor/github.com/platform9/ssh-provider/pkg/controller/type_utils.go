package controller

import (
	"fmt"

	spv1 "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func HasRole(a spv1.MachineRole, roles []spv1.MachineRole) bool {
	for _, b := range roles {
		if b == a {
			return true
		}
	}
	return false
}

func IsMaster(machineSpec *spv1.MachineSpec) bool {
	return HasRole(spv1.MasterRole, machineSpec.Roles)
}

func IsNode(machineSpec *spv1.MachineSpec) bool {
	return HasRole(spv1.NodeRole, machineSpec.Roles)
}

func UsernameAndKeyFromSecret(sshCredentialSecret *corev1.Secret) (string, string, error) {
	username, ok := sshCredentialSecret.Data["username"]
	if !ok {
		return "", "", fmt.Errorf("unable to find `username` key in secret %q", sshCredentialSecret.Name)
	}
	privateKey, ok := sshCredentialSecret.Data["ssh-privatekey"]
	if !ok {
		return "", "", fmt.Errorf("unable to find `ssh-privatekey` key in secret %q", sshCredentialSecret.Name)
	}
	return string(username), string(privateKey), nil
}

// BindMachineAndProvisionedMachine creates a bi-directional bind between
// machine and provisioned machine.
func BindMachineAndProvisionedMachine(machine *clusterv1.Machine, pm *spv1.ProvisionedMachine) error {
	// Bind the provisioned machine to the machine
	if pm.Status.MachineRef == nil {
		pm.Status.MachineRef = &corev1.LocalObjectReference{}
	}
	pm.Status.MachineRef.Name = machine.Name

	// Bind the machine to the provisioned machine
	machineSpec, err := GetMachineSpec(*machine)
	if err != nil {
		return fmt.Errorf("unable to decode machine spec: %v", err)
	}
	machineSpec.ProvisionedMachineName = pm.Name
	if err := PutMachineSpec(*machineSpec, machine); err != nil {
		return fmt.Errorf("unable to encode machine spec: %v", err)
	}

	// Update machine status SSHConfig
	machineStatus, err := GetMachineStatus(*machine)
	if err != nil {
		return fmt.Errorf("unable to decode machine status: %v", err)
	}
	machineStatus.SSHConfig = pm.Spec.SSHConfig.DeepCopy()
	if err := PutMachineStatus(*machineStatus, machine); err != nil {
		return fmt.Errorf("unable to encode machine status: %v", err)
	}

	return nil
}
