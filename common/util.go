package common

import (
	"log"
	"net"

	sshProvider "github.com/platform9/ssh-provider/sshproviderconfig/v1alpha1"
	sshconfigv1 "github.com/platform9/ssh-provider/sshproviderconfig/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func CreateSSHClusterProviderConfig(routerID int, vip string) (*clusterv1.ProviderConfig, error) {
	SSHClusterProviderConfig := sshconfigv1.SSHClusterProviderConfig{
		TypeMeta: v1.TypeMeta{
			APIVersion: "sshproviderconfig/v1alpha1",
			Kind:       "SSHClusterProviderConfig",
		},
		VIPConfiguration: &sshProvider.VIPConfiguration{
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
