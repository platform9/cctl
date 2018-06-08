/*
Copyright 2018 Platform 9 Systems, Inc.
*/

package sshproviderconfig

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	SchemeBuilder      runtime.SchemeBuilder
	AddToScheme        = SchemeBuilder.AddToScheme
	localSchemeBuilder = &SchemeBuilder
)

func init() {
	localSchemeBuilder.Register(addKnownTypes)
}

const GroupName = "sshproviderconfig"

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: runtime.APIVersionInternal}

func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&SSHMachineProviderConfig{},
	)
	scheme.AddKnownTypes(SchemeGroupVersion,
		&SSHClusterProviderConfig{},
	)
	return nil
}
