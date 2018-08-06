// +build !ignore_autogenerated

/*
Copyright 2018 Platform 9 Systems, Inc.

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

// Code generated by conversion-gen. DO NOT EDIT.

// This file was autogenerated by conversion-gen. Do not edit it manually!

package v1alpha1

import (
	unsafe "unsafe"

	sshprovider "github.com/platform9/ssh-provider/pkg/apis/sshprovider"
	v1 "k8s.io/api/core/v1"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(scheme *runtime.Scheme) error {
	return scheme.AddGeneratedConversionFuncs(
		Convert_v1alpha1_ClusterSpec_To_sshprovider_ClusterSpec,
		Convert_sshprovider_ClusterSpec_To_v1alpha1_ClusterSpec,
		Convert_v1alpha1_ClusterStatus_To_sshprovider_ClusterStatus,
		Convert_sshprovider_ClusterStatus_To_v1alpha1_ClusterStatus,
		Convert_v1alpha1_EtcdMember_To_sshprovider_EtcdMember,
		Convert_sshprovider_EtcdMember_To_v1alpha1_EtcdMember,
		Convert_v1alpha1_MachineComponentVersions_To_sshprovider_MachineComponentVersions,
		Convert_sshprovider_MachineComponentVersions_To_v1alpha1_MachineComponentVersions,
		Convert_v1alpha1_MachineSpec_To_sshprovider_MachineSpec,
		Convert_sshprovider_MachineSpec_To_v1alpha1_MachineSpec,
		Convert_v1alpha1_MachineStatus_To_sshprovider_MachineStatus,
		Convert_sshprovider_MachineStatus_To_v1alpha1_MachineStatus,
		Convert_v1alpha1_ProvisionedMachine_To_sshprovider_ProvisionedMachine,
		Convert_sshprovider_ProvisionedMachine_To_v1alpha1_ProvisionedMachine,
		Convert_v1alpha1_ProvisionedMachineList_To_sshprovider_ProvisionedMachineList,
		Convert_sshprovider_ProvisionedMachineList_To_v1alpha1_ProvisionedMachineList,
		Convert_v1alpha1_ProvisionedMachineSpec_To_sshprovider_ProvisionedMachineSpec,
		Convert_sshprovider_ProvisionedMachineSpec_To_v1alpha1_ProvisionedMachineSpec,
		Convert_v1alpha1_ProvisionedMachineStatus_To_sshprovider_ProvisionedMachineStatus,
		Convert_sshprovider_ProvisionedMachineStatus_To_v1alpha1_ProvisionedMachineStatus,
		Convert_v1alpha1_SSHConfig_To_sshprovider_SSHConfig,
		Convert_sshprovider_SSHConfig_To_v1alpha1_SSHConfig,
		Convert_v1alpha1_VIPConfiguration_To_sshprovider_VIPConfiguration,
		Convert_sshprovider_VIPConfiguration_To_v1alpha1_VIPConfiguration,
	)
}

func autoConvert_v1alpha1_ClusterSpec_To_sshprovider_ClusterSpec(in *ClusterSpec, out *sshprovider.ClusterSpec, s conversion.Scope) error {
	out.EtcdCASecret = (*v1.LocalObjectReference)(unsafe.Pointer(in.EtcdCASecret))
	out.APIServerCASecret = (*v1.LocalObjectReference)(unsafe.Pointer(in.APIServerCASecret))
	out.FrontProxyCASecret = (*v1.LocalObjectReference)(unsafe.Pointer(in.FrontProxyCASecret))
	out.ServiceAccountKeySecret = (*v1.LocalObjectReference)(unsafe.Pointer(in.ServiceAccountKeySecret))
	out.BootstrapTokenSecret = (*v1.LocalObjectReference)(unsafe.Pointer(in.BootstrapTokenSecret))
	out.VIPConfiguration = (*sshprovider.VIPConfiguration)(unsafe.Pointer(in.VIPConfiguration))
	return nil
}

// Convert_v1alpha1_ClusterSpec_To_sshprovider_ClusterSpec is an autogenerated conversion function.
func Convert_v1alpha1_ClusterSpec_To_sshprovider_ClusterSpec(in *ClusterSpec, out *sshprovider.ClusterSpec, s conversion.Scope) error {
	return autoConvert_v1alpha1_ClusterSpec_To_sshprovider_ClusterSpec(in, out, s)
}

func autoConvert_sshprovider_ClusterSpec_To_v1alpha1_ClusterSpec(in *sshprovider.ClusterSpec, out *ClusterSpec, s conversion.Scope) error {
	out.EtcdCASecret = (*v1.LocalObjectReference)(unsafe.Pointer(in.EtcdCASecret))
	out.APIServerCASecret = (*v1.LocalObjectReference)(unsafe.Pointer(in.APIServerCASecret))
	out.FrontProxyCASecret = (*v1.LocalObjectReference)(unsafe.Pointer(in.FrontProxyCASecret))
	out.ServiceAccountKeySecret = (*v1.LocalObjectReference)(unsafe.Pointer(in.ServiceAccountKeySecret))
	out.BootstrapTokenSecret = (*v1.LocalObjectReference)(unsafe.Pointer(in.BootstrapTokenSecret))
	out.VIPConfiguration = (*VIPConfiguration)(unsafe.Pointer(in.VIPConfiguration))
	return nil
}

// Convert_sshprovider_ClusterSpec_To_v1alpha1_ClusterSpec is an autogenerated conversion function.
func Convert_sshprovider_ClusterSpec_To_v1alpha1_ClusterSpec(in *sshprovider.ClusterSpec, out *ClusterSpec, s conversion.Scope) error {
	return autoConvert_sshprovider_ClusterSpec_To_v1alpha1_ClusterSpec(in, out, s)
}

func autoConvert_v1alpha1_ClusterStatus_To_sshprovider_ClusterStatus(in *ClusterStatus, out *sshprovider.ClusterStatus, s conversion.Scope) error {
	out.EtcdMembers = *(*[]sshprovider.EtcdMember)(unsafe.Pointer(&in.EtcdMembers))
	return nil
}

// Convert_v1alpha1_ClusterStatus_To_sshprovider_ClusterStatus is an autogenerated conversion function.
func Convert_v1alpha1_ClusterStatus_To_sshprovider_ClusterStatus(in *ClusterStatus, out *sshprovider.ClusterStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_ClusterStatus_To_sshprovider_ClusterStatus(in, out, s)
}

func autoConvert_sshprovider_ClusterStatus_To_v1alpha1_ClusterStatus(in *sshprovider.ClusterStatus, out *ClusterStatus, s conversion.Scope) error {
	out.EtcdMembers = *(*[]EtcdMember)(unsafe.Pointer(&in.EtcdMembers))
	return nil
}

// Convert_sshprovider_ClusterStatus_To_v1alpha1_ClusterStatus is an autogenerated conversion function.
func Convert_sshprovider_ClusterStatus_To_v1alpha1_ClusterStatus(in *sshprovider.ClusterStatus, out *ClusterStatus, s conversion.Scope) error {
	return autoConvert_sshprovider_ClusterStatus_To_v1alpha1_ClusterStatus(in, out, s)
}

func autoConvert_v1alpha1_EtcdMember_To_sshprovider_EtcdMember(in *EtcdMember, out *sshprovider.EtcdMember, s conversion.Scope) error {
	out.ID = in.ID
	out.Name = in.Name
	out.PeerURLs = *(*[]string)(unsafe.Pointer(&in.PeerURLs))
	out.ClientURLs = *(*[]string)(unsafe.Pointer(&in.ClientURLs))
	return nil
}

// Convert_v1alpha1_EtcdMember_To_sshprovider_EtcdMember is an autogenerated conversion function.
func Convert_v1alpha1_EtcdMember_To_sshprovider_EtcdMember(in *EtcdMember, out *sshprovider.EtcdMember, s conversion.Scope) error {
	return autoConvert_v1alpha1_EtcdMember_To_sshprovider_EtcdMember(in, out, s)
}

func autoConvert_sshprovider_EtcdMember_To_v1alpha1_EtcdMember(in *sshprovider.EtcdMember, out *EtcdMember, s conversion.Scope) error {
	out.ID = in.ID
	out.Name = in.Name
	out.PeerURLs = *(*[]string)(unsafe.Pointer(&in.PeerURLs))
	out.ClientURLs = *(*[]string)(unsafe.Pointer(&in.ClientURLs))
	return nil
}

// Convert_sshprovider_EtcdMember_To_v1alpha1_EtcdMember is an autogenerated conversion function.
func Convert_sshprovider_EtcdMember_To_v1alpha1_EtcdMember(in *sshprovider.EtcdMember, out *EtcdMember, s conversion.Scope) error {
	return autoConvert_sshprovider_EtcdMember_To_v1alpha1_EtcdMember(in, out, s)
}

func autoConvert_v1alpha1_MachineComponentVersions_To_sshprovider_MachineComponentVersions(in *MachineComponentVersions, out *sshprovider.MachineComponentVersions, s conversion.Scope) error {
	out.NodeadmVersion = in.NodeadmVersion
	out.EtcdadmVersion = in.EtcdadmVersion
	out.KubernetesVersion = in.KubernetesVersion
	out.CNIVersion = in.CNIVersion
	out.FlannelVersion = in.FlannelVersion
	out.KeepalivedVersion = in.KeepalivedVersion
	out.EtcdVersion = in.EtcdVersion
	return nil
}

// Convert_v1alpha1_MachineComponentVersions_To_sshprovider_MachineComponentVersions is an autogenerated conversion function.
func Convert_v1alpha1_MachineComponentVersions_To_sshprovider_MachineComponentVersions(in *MachineComponentVersions, out *sshprovider.MachineComponentVersions, s conversion.Scope) error {
	return autoConvert_v1alpha1_MachineComponentVersions_To_sshprovider_MachineComponentVersions(in, out, s)
}

func autoConvert_sshprovider_MachineComponentVersions_To_v1alpha1_MachineComponentVersions(in *sshprovider.MachineComponentVersions, out *MachineComponentVersions, s conversion.Scope) error {
	out.NodeadmVersion = in.NodeadmVersion
	out.EtcdadmVersion = in.EtcdadmVersion
	out.KubernetesVersion = in.KubernetesVersion
	out.CNIVersion = in.CNIVersion
	out.FlannelVersion = in.FlannelVersion
	out.KeepalivedVersion = in.KeepalivedVersion
	out.EtcdVersion = in.EtcdVersion
	return nil
}

// Convert_sshprovider_MachineComponentVersions_To_v1alpha1_MachineComponentVersions is an autogenerated conversion function.
func Convert_sshprovider_MachineComponentVersions_To_v1alpha1_MachineComponentVersions(in *sshprovider.MachineComponentVersions, out *MachineComponentVersions, s conversion.Scope) error {
	return autoConvert_sshprovider_MachineComponentVersions_To_v1alpha1_MachineComponentVersions(in, out, s)
}

func autoConvert_v1alpha1_MachineSpec_To_sshprovider_MachineSpec(in *MachineSpec, out *sshprovider.MachineSpec, s conversion.Scope) error {
	out.Roles = *(*[]sshprovider.MachineRole)(unsafe.Pointer(&in.Roles))
	out.ProvisionedMachineName = in.ProvisionedMachineName
	out.ComponentVersions = (*sshprovider.MachineComponentVersions)(unsafe.Pointer(in.ComponentVersions))
	return nil
}

// Convert_v1alpha1_MachineSpec_To_sshprovider_MachineSpec is an autogenerated conversion function.
func Convert_v1alpha1_MachineSpec_To_sshprovider_MachineSpec(in *MachineSpec, out *sshprovider.MachineSpec, s conversion.Scope) error {
	return autoConvert_v1alpha1_MachineSpec_To_sshprovider_MachineSpec(in, out, s)
}

func autoConvert_sshprovider_MachineSpec_To_v1alpha1_MachineSpec(in *sshprovider.MachineSpec, out *MachineSpec, s conversion.Scope) error {
	out.Roles = *(*[]MachineRole)(unsafe.Pointer(&in.Roles))
	out.ProvisionedMachineName = in.ProvisionedMachineName
	out.ComponentVersions = (*MachineComponentVersions)(unsafe.Pointer(in.ComponentVersions))
	return nil
}

// Convert_sshprovider_MachineSpec_To_v1alpha1_MachineSpec is an autogenerated conversion function.
func Convert_sshprovider_MachineSpec_To_v1alpha1_MachineSpec(in *sshprovider.MachineSpec, out *MachineSpec, s conversion.Scope) error {
	return autoConvert_sshprovider_MachineSpec_To_v1alpha1_MachineSpec(in, out, s)
}

func autoConvert_v1alpha1_MachineStatus_To_sshprovider_MachineStatus(in *MachineStatus, out *sshprovider.MachineStatus, s conversion.Scope) error {
	out.SSHConfig = (*sshprovider.SSHConfig)(unsafe.Pointer(in.SSHConfig))
	out.VIPNetworkInterface = in.VIPNetworkInterface
	out.EtcdMember = (*sshprovider.EtcdMember)(unsafe.Pointer(in.EtcdMember))
	return nil
}

// Convert_v1alpha1_MachineStatus_To_sshprovider_MachineStatus is an autogenerated conversion function.
func Convert_v1alpha1_MachineStatus_To_sshprovider_MachineStatus(in *MachineStatus, out *sshprovider.MachineStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_MachineStatus_To_sshprovider_MachineStatus(in, out, s)
}

func autoConvert_sshprovider_MachineStatus_To_v1alpha1_MachineStatus(in *sshprovider.MachineStatus, out *MachineStatus, s conversion.Scope) error {
	out.SSHConfig = (*SSHConfig)(unsafe.Pointer(in.SSHConfig))
	out.VIPNetworkInterface = in.VIPNetworkInterface
	out.EtcdMember = (*EtcdMember)(unsafe.Pointer(in.EtcdMember))
	return nil
}

// Convert_sshprovider_MachineStatus_To_v1alpha1_MachineStatus is an autogenerated conversion function.
func Convert_sshprovider_MachineStatus_To_v1alpha1_MachineStatus(in *sshprovider.MachineStatus, out *MachineStatus, s conversion.Scope) error {
	return autoConvert_sshprovider_MachineStatus_To_v1alpha1_MachineStatus(in, out, s)
}

func autoConvert_v1alpha1_ProvisionedMachine_To_sshprovider_ProvisionedMachine(in *ProvisionedMachine, out *sshprovider.ProvisionedMachine, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1alpha1_ProvisionedMachineSpec_To_sshprovider_ProvisionedMachineSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1alpha1_ProvisionedMachineStatus_To_sshprovider_ProvisionedMachineStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1alpha1_ProvisionedMachine_To_sshprovider_ProvisionedMachine is an autogenerated conversion function.
func Convert_v1alpha1_ProvisionedMachine_To_sshprovider_ProvisionedMachine(in *ProvisionedMachine, out *sshprovider.ProvisionedMachine, s conversion.Scope) error {
	return autoConvert_v1alpha1_ProvisionedMachine_To_sshprovider_ProvisionedMachine(in, out, s)
}

func autoConvert_sshprovider_ProvisionedMachine_To_v1alpha1_ProvisionedMachine(in *sshprovider.ProvisionedMachine, out *ProvisionedMachine, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_sshprovider_ProvisionedMachineSpec_To_v1alpha1_ProvisionedMachineSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_sshprovider_ProvisionedMachineStatus_To_v1alpha1_ProvisionedMachineStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_sshprovider_ProvisionedMachine_To_v1alpha1_ProvisionedMachine is an autogenerated conversion function.
func Convert_sshprovider_ProvisionedMachine_To_v1alpha1_ProvisionedMachine(in *sshprovider.ProvisionedMachine, out *ProvisionedMachine, s conversion.Scope) error {
	return autoConvert_sshprovider_ProvisionedMachine_To_v1alpha1_ProvisionedMachine(in, out, s)
}

func autoConvert_v1alpha1_ProvisionedMachineList_To_sshprovider_ProvisionedMachineList(in *ProvisionedMachineList, out *sshprovider.ProvisionedMachineList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]sshprovider.ProvisionedMachine)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_v1alpha1_ProvisionedMachineList_To_sshprovider_ProvisionedMachineList is an autogenerated conversion function.
func Convert_v1alpha1_ProvisionedMachineList_To_sshprovider_ProvisionedMachineList(in *ProvisionedMachineList, out *sshprovider.ProvisionedMachineList, s conversion.Scope) error {
	return autoConvert_v1alpha1_ProvisionedMachineList_To_sshprovider_ProvisionedMachineList(in, out, s)
}

func autoConvert_sshprovider_ProvisionedMachineList_To_v1alpha1_ProvisionedMachineList(in *sshprovider.ProvisionedMachineList, out *ProvisionedMachineList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]ProvisionedMachine)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_sshprovider_ProvisionedMachineList_To_v1alpha1_ProvisionedMachineList is an autogenerated conversion function.
func Convert_sshprovider_ProvisionedMachineList_To_v1alpha1_ProvisionedMachineList(in *sshprovider.ProvisionedMachineList, out *ProvisionedMachineList, s conversion.Scope) error {
	return autoConvert_sshprovider_ProvisionedMachineList_To_v1alpha1_ProvisionedMachineList(in, out, s)
}

func autoConvert_v1alpha1_ProvisionedMachineSpec_To_sshprovider_ProvisionedMachineSpec(in *ProvisionedMachineSpec, out *sshprovider.ProvisionedMachineSpec, s conversion.Scope) error {
	out.SSHConfig = (*sshprovider.SSHConfig)(unsafe.Pointer(in.SSHConfig))
	out.VIPNetworkInterface = in.VIPNetworkInterface
	return nil
}

// Convert_v1alpha1_ProvisionedMachineSpec_To_sshprovider_ProvisionedMachineSpec is an autogenerated conversion function.
func Convert_v1alpha1_ProvisionedMachineSpec_To_sshprovider_ProvisionedMachineSpec(in *ProvisionedMachineSpec, out *sshprovider.ProvisionedMachineSpec, s conversion.Scope) error {
	return autoConvert_v1alpha1_ProvisionedMachineSpec_To_sshprovider_ProvisionedMachineSpec(in, out, s)
}

func autoConvert_sshprovider_ProvisionedMachineSpec_To_v1alpha1_ProvisionedMachineSpec(in *sshprovider.ProvisionedMachineSpec, out *ProvisionedMachineSpec, s conversion.Scope) error {
	out.SSHConfig = (*SSHConfig)(unsafe.Pointer(in.SSHConfig))
	out.VIPNetworkInterface = in.VIPNetworkInterface
	return nil
}

// Convert_sshprovider_ProvisionedMachineSpec_To_v1alpha1_ProvisionedMachineSpec is an autogenerated conversion function.
func Convert_sshprovider_ProvisionedMachineSpec_To_v1alpha1_ProvisionedMachineSpec(in *sshprovider.ProvisionedMachineSpec, out *ProvisionedMachineSpec, s conversion.Scope) error {
	return autoConvert_sshprovider_ProvisionedMachineSpec_To_v1alpha1_ProvisionedMachineSpec(in, out, s)
}

func autoConvert_v1alpha1_ProvisionedMachineStatus_To_sshprovider_ProvisionedMachineStatus(in *ProvisionedMachineStatus, out *sshprovider.ProvisionedMachineStatus, s conversion.Scope) error {
	out.MachineRef = (*v1.LocalObjectReference)(unsafe.Pointer(in.MachineRef))
	return nil
}

// Convert_v1alpha1_ProvisionedMachineStatus_To_sshprovider_ProvisionedMachineStatus is an autogenerated conversion function.
func Convert_v1alpha1_ProvisionedMachineStatus_To_sshprovider_ProvisionedMachineStatus(in *ProvisionedMachineStatus, out *sshprovider.ProvisionedMachineStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_ProvisionedMachineStatus_To_sshprovider_ProvisionedMachineStatus(in, out, s)
}

func autoConvert_sshprovider_ProvisionedMachineStatus_To_v1alpha1_ProvisionedMachineStatus(in *sshprovider.ProvisionedMachineStatus, out *ProvisionedMachineStatus, s conversion.Scope) error {
	out.MachineRef = (*v1.LocalObjectReference)(unsafe.Pointer(in.MachineRef))
	return nil
}

// Convert_sshprovider_ProvisionedMachineStatus_To_v1alpha1_ProvisionedMachineStatus is an autogenerated conversion function.
func Convert_sshprovider_ProvisionedMachineStatus_To_v1alpha1_ProvisionedMachineStatus(in *sshprovider.ProvisionedMachineStatus, out *ProvisionedMachineStatus, s conversion.Scope) error {
	return autoConvert_sshprovider_ProvisionedMachineStatus_To_v1alpha1_ProvisionedMachineStatus(in, out, s)
}

func autoConvert_v1alpha1_SSHConfig_To_sshprovider_SSHConfig(in *SSHConfig, out *sshprovider.SSHConfig, s conversion.Scope) error {
	out.Host = in.Host
	out.Port = in.Port
	out.PublicKeys = *(*[]string)(unsafe.Pointer(&in.PublicKeys))
	out.CredentialSecret = in.CredentialSecret
	return nil
}

// Convert_v1alpha1_SSHConfig_To_sshprovider_SSHConfig is an autogenerated conversion function.
func Convert_v1alpha1_SSHConfig_To_sshprovider_SSHConfig(in *SSHConfig, out *sshprovider.SSHConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_SSHConfig_To_sshprovider_SSHConfig(in, out, s)
}

func autoConvert_sshprovider_SSHConfig_To_v1alpha1_SSHConfig(in *sshprovider.SSHConfig, out *SSHConfig, s conversion.Scope) error {
	out.Host = in.Host
	out.Port = in.Port
	out.PublicKeys = *(*[]string)(unsafe.Pointer(&in.PublicKeys))
	out.CredentialSecret = in.CredentialSecret
	return nil
}

// Convert_sshprovider_SSHConfig_To_v1alpha1_SSHConfig is an autogenerated conversion function.
func Convert_sshprovider_SSHConfig_To_v1alpha1_SSHConfig(in *sshprovider.SSHConfig, out *SSHConfig, s conversion.Scope) error {
	return autoConvert_sshprovider_SSHConfig_To_v1alpha1_SSHConfig(in, out, s)
}

func autoConvert_v1alpha1_VIPConfiguration_To_sshprovider_VIPConfiguration(in *VIPConfiguration, out *sshprovider.VIPConfiguration, s conversion.Scope) error {
	out.IP = in.IP
	out.RouterID = in.RouterID
	return nil
}

// Convert_v1alpha1_VIPConfiguration_To_sshprovider_VIPConfiguration is an autogenerated conversion function.
func Convert_v1alpha1_VIPConfiguration_To_sshprovider_VIPConfiguration(in *VIPConfiguration, out *sshprovider.VIPConfiguration, s conversion.Scope) error {
	return autoConvert_v1alpha1_VIPConfiguration_To_sshprovider_VIPConfiguration(in, out, s)
}

func autoConvert_sshprovider_VIPConfiguration_To_v1alpha1_VIPConfiguration(in *sshprovider.VIPConfiguration, out *VIPConfiguration, s conversion.Scope) error {
	out.IP = in.IP
	out.RouterID = in.RouterID
	return nil
}

// Convert_sshprovider_VIPConfiguration_To_v1alpha1_VIPConfiguration is an autogenerated conversion function.
func Convert_sshprovider_VIPConfiguration_To_v1alpha1_VIPConfiguration(in *sshprovider.VIPConfiguration, out *VIPConfiguration, s conversion.Scope) error {
	return autoConvert_sshprovider_VIPConfiguration_To_v1alpha1_VIPConfiguration(in, out, s)
}
