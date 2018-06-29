/*
Copyright 2018 Platform 9 Systems, Inc.
*/

package v1alpha1

import (
	"bytes"
	"fmt"

	"github.com/platform9/ssh-provider/sshproviderconfig"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// +k8s:deepcopy-gen=false
type SSHProviderCodec struct {
	encoder runtime.Encoder
	decoder runtime.Decoder
}

const GroupName = "sshproviderconfig"

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

var (
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)

func init() {
	localSchemeBuilder.Register(addKnownTypes)
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&SSHMachineProviderConfig{},
	)
	scheme.AddKnownTypes(SchemeGroupVersion,
		&SSHClusterProviderConfig{},
	)
	scheme.AddKnownTypes(SchemeGroupVersion,
		&SSHMachineProviderStatus{},
	)
	scheme.AddKnownTypes(SchemeGroupVersion,
		&SSHClusterProviderStatus{},
	)
	return nil
}

func NewScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := sshproviderconfig.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

func NewCodec() (*SSHProviderCodec, error) {
	scheme, err := NewScheme()
	if err != nil {
		return nil, err
	}
	codecFactory := serializer.NewCodecFactory(scheme)
	encoder, err := newEncoder(&codecFactory)
	if err != nil {
		return nil, err
	}
	codec := SSHProviderCodec{
		encoder: encoder,
		decoder: codecFactory.UniversalDecoder(SchemeGroupVersion),
	}
	return &codec, nil
}

func (codec *SSHProviderCodec) DecodeFromProviderConfig(providerConfig clusterv1.ProviderConfig, out runtime.Object) error {
	_, _, err := codec.decoder.Decode(providerConfig.Value.Raw, nil, out)
	if err != nil {
		return fmt.Errorf("decoding failed: %v", err)
	}
	return nil
}

func (codec *SSHProviderCodec) EncodeToProviderConfig(in runtime.Object) (*clusterv1.ProviderConfig, error) {
	var buf bytes.Buffer
	if err := codec.encoder.Encode(in, &buf); err != nil {
		return nil, fmt.Errorf("encoding failed: %v", err)
	}
	return &clusterv1.ProviderConfig{
		Value: &runtime.RawExtension{Raw: buf.Bytes()},
	}, nil
}

func (codec *SSHProviderCodec) DecodeFromProviderStatus(providerStatus clusterv1.ProviderStatus, out runtime.Object) error {
	_, _, err := codec.decoder.Decode(providerStatus.Value.Raw, nil, out)
	if err != nil {
		return fmt.Errorf("decoding failed: %v", err)
	}
	return nil
}

func (codec *SSHProviderCodec) EncodeToProviderStatus(in runtime.Object) (*clusterv1.ProviderStatus, error) {
	var buf bytes.Buffer
	if err := codec.encoder.Encode(in, &buf); err != nil {
		return nil, fmt.Errorf("encoding failed: %v", err)
	}
	return &clusterv1.ProviderStatus{
		Value: &runtime.RawExtension{Raw: buf.Bytes()},
	}, nil
}

func newEncoder(codecFactory *serializer.CodecFactory) (runtime.Encoder, error) {
	serializerInfos := codecFactory.SupportedMediaTypes()
	if len(serializerInfos) == 0 {
		return nil, fmt.Errorf("unable to find any serializers")
	}
	encoder := codecFactory.EncoderForVersion(serializerInfos[0].Serializer, SchemeGroupVersion)
	return encoder, nil
}
