/*
Copyright 2018 Platform 9 Systems, Inc.
*/

package sshproviderconfig

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SSHMachineProviderConfig struct {
	metav1.TypeMeta `json:",inline"`

	Host          string   `json:"host"`
	Port          string   `json:"port"`
	PublicKeys    []string `json:"publicKeys"`
	SSHSecretName string   `json:"sshSecretName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SSHClusterProviderConfig struct {
	metav1.TypeMeta `json:",inline"`

	CASecretName      string   `json:"caSecretName"`
	APIServerCertSANs []string `json:"apiServerCertSans"`
}
