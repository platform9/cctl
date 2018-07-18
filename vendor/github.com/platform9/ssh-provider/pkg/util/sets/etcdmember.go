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
	"sort"

	spv1 "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
)

type EtcdMemberSet map[uint64]spv1.EtcdMember

func NewEtcdMemberSet(items ...spv1.EtcdMember) EtcdMemberSet {
	s := EtcdMemberSet{}
	for _, item := range items {
		s[item.ID] = item
	}
	return s
}

func (s EtcdMemberSet) Insert(items ...spv1.EtcdMember) {
	for _, item := range items {
		s[item.ID] = item
	}
}

func (s EtcdMemberSet) Delete(items ...spv1.EtcdMember) {
	for _, item := range items {
		delete(s, item.ID)
	}
}

func (s EtcdMemberSet) Has(item spv1.EtcdMember) bool {
	_, contained := s[item.ID]
	return contained
}

type sortableSliceOfEtcdMember []spv1.EtcdMember

func (s sortableSliceOfEtcdMember) Len() int           { return len(s) }
func (s sortableSliceOfEtcdMember) Less(i, j int) bool { return lessEtcdMember(s[i], s[j]) }
func (s sortableSliceOfEtcdMember) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// List returns the contents as a slice sorted by etcd member ID.
func (s EtcdMemberSet) List() []spv1.EtcdMember {
	res := make(sortableSliceOfEtcdMember, 0, len(s))
	for _, v := range s {
		res = append(res, v)
	}
	sort.Sort(res)
	return []spv1.EtcdMember(res)
}

func lessEtcdMember(lhs, rhs spv1.EtcdMember) bool {
	return lhs.ID < rhs.ID
}
