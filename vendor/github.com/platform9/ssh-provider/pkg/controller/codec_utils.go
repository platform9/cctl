package controller

import (
	"bytes"
	"fmt"

	"github.com/platform9/ssh-provider/pkg/api"
	sshprovider "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	clusterapi "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func PutClusterSpec(clusterSpec sshprovider.ClusterSpec, cluster *clusterapi.Cluster) error {
	providerConfig, err := ClusterAPIProviderConfigFromClusterSpec(&clusterSpec)
	if err != nil {
		return fmt.Errorf("unable to decode ProviderConfig from ClusterSpec: %v", err)
	}
	cluster.Spec.ProviderConfig.Value = providerConfig
	return nil
}

func GetClusterSpec(cluster clusterapi.Cluster) (*sshprovider.ClusterSpec, error) {
	clusterSpec, err := ClusterSpecFromClusterAPI(&cluster)
	if err != nil {
		return nil, fmt.Errorf("unable to decode ClusterSpec from ProviderConfig: %v", err)
	}
	return clusterSpec, nil
}

func PutClusterStatus(clusterStatus sshprovider.ClusterStatus, cluster *clusterapi.Cluster) error {
	providerStatus, err := ClusterAPIProviderStatusFromClusterStatus(&clusterStatus)
	if err != nil {
		return fmt.Errorf("unable to decode ProviderStatus from ClusterStatus: %v", err)
	}
	cluster.Status.ProviderStatus = providerStatus
	return nil
}

func GetClusterStatus(cluster clusterapi.Cluster) (*sshprovider.ClusterStatus, error) {
	clusterStatus, err := ClusterStatusFromClusterAPI(&cluster)
	if err != nil {
		return nil, fmt.Errorf("unable to decode ClusterStatus from ProviderStatus: %v", err)
	}
	return clusterStatus, nil
}

func PutMachineSpec(machineSpec sshprovider.MachineSpec, machine *clusterapi.Machine) error {
	providerConfig, err := ClusterAPIProviderConfigFromMachineSpec(&machineSpec)
	if err != nil {
		return fmt.Errorf("unable to decode ProviderConfig from MachineSpec: %v", err)
	}
	machine.Spec.ProviderConfig.Value = providerConfig
	return nil
}

func GetMachineSpec(machine clusterapi.Machine) (*sshprovider.MachineSpec, error) {
	machineSpec, err := MachineSpecFromClusterAPI(&machine)
	if err != nil {
		return nil, fmt.Errorf("unable to decode MachineSpec from ProviderConfig: %v", err)
	}
	return machineSpec, nil
}

func PutMachineStatus(machineStatus sshprovider.MachineStatus, machine *clusterapi.Machine) error {
	providerStatus, err := ClusterAPIProviderStatusFromMachineStatus(&machineStatus)
	if err != nil {
		return fmt.Errorf("unable to decode ProviderStatus from MachineStatus: %v", err)
	}
	machine.Status.ProviderStatus = providerStatus
	return nil
}

func GetMachineStatus(machine clusterapi.Machine) (*sshprovider.MachineStatus, error) {
	machineStatus, err := MachineStatusFromClusterAPI(&machine)
	if err != nil {
		return nil, fmt.Errorf("unable to decode MachineStatus from ProviderStatus: %v", err)
	}
	return machineStatus, nil
}

func ClusterSpecFromClusterAPI(cluster *clusterapi.Cluster) (*sshprovider.ClusterSpec, error) {
	if cluster.Spec.ProviderConfig.Value == nil {
		return &sshprovider.ClusterSpec{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterStatus",
				APIVersion: "sshprovider.platform9.com/v1alpha1",
			},
		}, nil
	}
	obj, gvk, err := api.Codecs.UniversalDecoder(sshprovider.SchemeGroupVersion).Decode([]byte(cluster.Spec.ProviderConfig.Value.Raw), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decode ProviderConfig: %v", err)
	}
	status, ok := obj.(*sshprovider.ClusterSpec)
	if !ok {
		return nil, fmt.Errorf("Unexpected object: %#v", gvk)
	}
	return status, nil
}

func ClusterStatusFromClusterAPI(cluster *clusterapi.Cluster) (*sshprovider.ClusterStatus, error) {
	if cluster.Status.ProviderStatus == nil {
		return &sshprovider.ClusterStatus{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterStatus",
				APIVersion: "sshprovider.platform9.com/v1alpha1",
			},
		}, nil
	}
	obj, gvk, err := api.Codecs.UniversalDecoder(sshprovider.SchemeGroupVersion).Decode([]byte(cluster.Status.ProviderStatus.Raw), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decode ProviderStatus: %v", err)
	}
	status, ok := obj.(*sshprovider.ClusterStatus)
	if !ok {
		return nil, fmt.Errorf("Unexpected object: %#v", gvk)
	}
	return status, nil
}

func MachineSpecFromClusterAPI(machine *clusterapi.Machine) (*sshprovider.MachineSpec, error) {
	if machine.Spec.ProviderConfig.Value == nil {
		return &sshprovider.MachineSpec{
			TypeMeta: metav1.TypeMeta{
				Kind:       "MachineStatus",
				APIVersion: "sshprovider.platform9.com/v1alpha1",
			},
		}, nil
	}
	obj, gvk, err := api.Codecs.UniversalDecoder(sshprovider.SchemeGroupVersion).Decode([]byte(machine.Spec.ProviderConfig.Value.Raw), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decode ProviderConfig: %v", err)
	}
	status, ok := obj.(*sshprovider.MachineSpec)
	if !ok {
		return nil, fmt.Errorf("Unexpected object: %#v", gvk)
	}
	return status, nil
}

func MachineStatusFromClusterAPI(machine *clusterapi.Machine) (*sshprovider.MachineStatus, error) {
	if machine.Status.ProviderStatus == nil {
		return &sshprovider.MachineStatus{
			TypeMeta: metav1.TypeMeta{
				Kind:       "MachineStatus",
				APIVersion: "sshprovider.platform9.com/v1alpha1",
			},
		}, nil
	}
	obj, gvk, err := api.Codecs.UniversalDecoder(sshprovider.SchemeGroupVersion).Decode([]byte(machine.Status.ProviderStatus.Raw), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decode ProviderStatus: %v", err)
	}
	status, ok := obj.(*sshprovider.MachineStatus)
	if !ok {
		return nil, fmt.Errorf("Unexpected object: %#v", gvk)
	}
	return status, nil
}

func ClusterAPIProviderStatusFromClusterStatus(clusterStatus *sshprovider.ClusterStatus) (*runtime.RawExtension, error) {
	serializer := jsonserializer.NewSerializer(jsonserializer.DefaultMetaFactory, api.Scheme, api.Scheme, false)
	var buffer bytes.Buffer
	err := serializer.Encode(clusterStatus, &buffer)
	if err != nil {
		return nil, err
	}
	return &runtime.RawExtension{
		Raw: bytes.TrimSpace(buffer.Bytes()),
	}, nil
}

func ClusterAPIProviderStatusFromMachineStatus(machineStatus *sshprovider.MachineStatus) (*runtime.RawExtension, error) {
	serializer := jsonserializer.NewSerializer(jsonserializer.DefaultMetaFactory, api.Scheme, api.Scheme, false)
	var buffer bytes.Buffer
	err := serializer.Encode(machineStatus, &buffer)
	if err != nil {
		return nil, err
	}
	return &runtime.RawExtension{
		Raw: bytes.TrimSpace(buffer.Bytes()),
	}, nil
}

func ClusterAPIProviderConfigFromClusterSpec(clusterSpec *sshprovider.ClusterSpec) (*runtime.RawExtension, error) {
	serializer := jsonserializer.NewSerializer(jsonserializer.DefaultMetaFactory, api.Scheme, api.Scheme, false)
	var buffer bytes.Buffer
	err := serializer.Encode(clusterSpec, &buffer)
	if err != nil {
		return nil, err
	}
	return &runtime.RawExtension{
		Raw: bytes.TrimSpace(buffer.Bytes()),
	}, nil
}

func ClusterAPIProviderConfigFromMachineSpec(machineSpec *sshprovider.MachineSpec) (*runtime.RawExtension, error) {
	serializer := jsonserializer.NewSerializer(jsonserializer.DefaultMetaFactory, api.Scheme, api.Scheme, false)
	var buffer bytes.Buffer
	err := serializer.Encode(machineSpec, &buffer)
	if err != nil {
		return nil, err
	}
	return &runtime.RawExtension{
		Raw: bytes.TrimSpace(buffer.Bytes()),
	}, nil
}
