package common

import (
	"log"
	"net"

	sshconfigv1 "github.com/platform9/ssh-provider/sshproviderconfig/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	clustercommon "sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func CreateSSHClusterProviderConfig(routerID int, vip string) (*clusterv1.ProviderConfig, error) {
	SSHClusterProviderConfig := sshconfigv1.SSHClusterProviderConfig{
		TypeMeta: v1.TypeMeta{
			APIVersion: "sshproviderconfig/v1alpha1",
			Kind:       "SSHClusterProviderConfig",
		},
		VIPConfiguration: &sshconfigv1.VIPConfiguration{
			IP:       net.ParseIP(vip),
			RouterID: routerID,
		},
	}

	sshClusterProviderConfigCodec, err := sshconfigv1.NewCodec()
	if err != nil {
		log.Fatal(err)
	}
	sshProviderConfig, err := sshClusterProviderConfigCodec.EncodeToProviderConfig(&SSHClusterProviderConfig)

	return sshProviderConfig, err
}

func CreateSSHMachineProviderConfig(cmd *cobra.Command) (*clusterv1.ProviderConfig, error) {
	//port, err := strconv.Atoi(cmd.Flag("port").Value.String())
	//keys := strings.Split(cmd.Flag("publicKeys").Value.String(), ",")
	SSHMachineProviderConfig := sshconfigv1.SSHMachineProviderConfig{
		TypeMeta: v1.TypeMeta{
			APIVersion: "sshproviderconfig/v1alpha1",
			Kind:       "SSHMachineProviderConfig",
		},
		//Host:          cmd.Flag("ip").Value.String(),
		//Port:          port,
		//PublicKeys:    keys,
		//SSHSecretName: cmd.Flag("sshSecretName").Value.String(),
	}
	//
	SSHMachineProviderConfigCodec, err := sshconfigv1.NewCodec()
	if err != nil {
		log.Fatal(err)
	}

	sshProviderConfig, err := SSHMachineProviderConfigCodec.EncodeToProviderConfig(&SSHMachineProviderConfig)

	return sshProviderConfig, err
}

func DecodeSSHClusterProviderConfig(pc clusterv1.ProviderConfig) sshconfigv1.SSHClusterProviderConfig {
	var config sshconfigv1.SSHClusterProviderConfig
	sshProviderConfigCodec, err := sshconfigv1.NewCodec()
	if err != nil {
		log.Fatal(err)
	}
	sshProviderConfigCodec.DecodeFromProviderConfig(pc, &config)
	return config
}

func DecodeSSHMachineProviderStatus(machineProviderStatus clusterv1.ProviderStatus) sshconfigv1.SSHMachineProviderStatus {
	config := sshconfigv1.SSHMachineProviderStatus{}

	sshProviderConfigCodec, err := sshconfigv1.NewCodec()
	if err != nil {
		log.Fatal(err)
	}
	if machineProviderStatus.Value != nil {
		sshProviderConfigCodec.DecodeFromProviderStatus(machineProviderStatus, &config)
	}
	return config
}

func EncodeSSHMachineProviderStatus(machineProviderStatus sshconfigv1.SSHMachineProviderStatus) (*clusterv1.ProviderStatus, error) {
	sshProviderConfigCodec, err := sshconfigv1.NewCodec()
	if err != nil {
		log.Fatal(err)
	}
	return sshProviderConfigCodec.EncodeToProviderStatus(machineProviderStatus.DeepCopyObject())
}

func EncodeSSHClusterProviderConfig(providerConfig sshconfigv1.SSHClusterProviderConfig) (*clusterv1.ProviderConfig, error) {
	sshProviderConfigCodec, err := sshconfigv1.NewCodec()
	if err != nil {
		log.Fatal(err)
	}
	return sshProviderConfigCodec.EncodeToProviderConfig(providerConfig.DeepCopyObject())
}

func DecodeSSHClusterProviderStatus(clusterProviderStatus clusterv1.ProviderStatus) sshconfigv1.SSHClusterProviderStatus {
	config := sshconfigv1.SSHClusterProviderStatus{}

	sshProviderConfigCodec, err := sshconfigv1.NewCodec()
	if err != nil {
		log.Fatal(err)
	}
	if clusterProviderStatus.Value != nil {
		sshProviderConfigCodec.DecodeFromProviderStatus(clusterProviderStatus, &config)
	}

	return config
}

func EncodeSSHClusterProviderStatus(clusterProviderStatus sshconfigv1.SSHClusterProviderStatus) (*clusterv1.ProviderStatus, error) {
	sshProviderConfigCodec, err := sshconfigv1.NewCodec()
	if err != nil {
		log.Fatal(err)
	}
	return sshProviderConfigCodec.EncodeToProviderStatus(clusterProviderStatus.DeepCopyObject())
}

func IsMaster(machine clusterv1.Machine) bool {
	for _, r := range machine.Spec.Roles {
		if r == clustercommon.MasterRole {
			return true
		}
	}
	return false
}
