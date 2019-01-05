/*
Copyright The Kubernetes Authors.

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

// Package sets implements various Set abstract data types. It is based on
// k8s.io/apimachinery/pkg/util/sets.
package sets

import (
	"fmt"
	"sort"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type APIEndpointSet map[clusterv1.APIEndpoint]struct{}

func NewAPIEndpointSet(items ...clusterv1.APIEndpoint) APIEndpointSet {
	s := APIEndpointSet{}
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

func (s APIEndpointSet) Insert(items ...clusterv1.APIEndpoint) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s APIEndpointSet) Delete(items ...clusterv1.APIEndpoint) {
	for _, item := range items {
		delete(s, item)
	}
}

func (s APIEndpointSet) Has(item clusterv1.APIEndpoint) bool {
	_, contained := s[item]
	return contained
}

type sortableSliceOfAPIEndpoint []clusterv1.APIEndpoint

func (s sortableSliceOfAPIEndpoint) Len() int           { return len(s) }
func (s sortableSliceOfAPIEndpoint) Less(i, j int) bool { return lessAPIEndpoint(s[i], s[j]) }
func (s sortableSliceOfAPIEndpoint) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// List returns the contents as a slice sorted by etcd member ID.
func (s APIEndpointSet) List() []clusterv1.APIEndpoint {
	res := make(sortableSliceOfAPIEndpoint, 0, len(s))
	for v := range s {
		res = append(res, v)
	}
	sort.Sort(res)
	return []clusterv1.APIEndpoint(res)
}

func lessAPIEndpoint(lhs, rhs clusterv1.APIEndpoint) bool {
	return fmt.Sprintf("%s:%d", lhs.Host, lhs.Port) < fmt.Sprintf("%s:%d", rhs.Host, rhs.Port)
}
