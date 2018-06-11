package common

import (
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	sshconfigv1 "github.com/platform9/ssh-provider/sshproviderconfig/v1alpha1"
)

func CreateSSHClusterProviderConfig(cmd *cobra.Command )  (*clusterv1.ProviderConfig, error) {
	SSHClusterProviderConfig := sshconfigv1.SSHClusterProviderConfig{
		TypeMeta: v1.TypeMeta{
			APIVersion: "sshproviderconfig/v1alpha1",
			Kind:       "SSHClusterProviderConfig",
		},
		CASecretName:"",
		APIServerCertSANs: []string{cmd.Flag("vip").Value.String()},
	}

	sshProviderConfigCodec, err := sshconfigv1.NewCodec()
	if err != nil {
		log.Fatal(err)
	}
	sshProviderConfig, err := sshProviderConfigCodec.EncodeToProviderConfig(&SSHClusterProviderConfig)
	return sshProviderConfig, err;
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
