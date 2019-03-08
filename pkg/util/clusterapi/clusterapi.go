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
