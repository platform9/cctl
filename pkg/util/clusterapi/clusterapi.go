package clusterapi

import (
	clustercommon "sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// MachinesWithRole returns every machine in the list that has the role.
func MachinesWithRole(machines []clusterv1.Machine, role clustercommon.MachineRole) []clusterv1.Machine {
	mwr := make([]clusterv1.Machine, 0)
	for _, m := range machines {
		for _, r := range m.Spec.Roles {
			if r == role {
				mwr = append(mwr, m)
			}
		}
	}
	return mwr
}
