package cmd

import (
	"fmt"

	"github.com/platform9/cctl/common"
	log "github.com/platform9/cctl/pkg/logrus"
	"github.com/platform9/cctl/pkg/util/secret"
	"github.com/spf13/cobra"
)

var secretCmdCreate = &cobra.Command{
	Use:   "secrets",
	Short: "Create default secrets",
	Run: func(cmd *cobra.Command, args []string) {
		err := createSecretDefaults()
		if err != nil {
			log.Fatalf("Unable to create secrets: %v", err)
		}
		log.Println("Secrets created successfully.")
	},
}

func createSecretDefaults() error {
	newAPIServerCASecret, err := secret.CreateCASecretDefault(common.DefaultAPIServerCASecretName)
	if err != nil {
		return fmt.Errorf("unable to generate API server CA secret: %v", err)
	}
	newEtcdCASecret, err := secret.CreateCASecretDefault(common.DefaultEtcdCASecretName)
	if err != nil {
		return fmt.Errorf("unable to generate etcd CA secret: %v", err)
	}
	newFrontProxyCASecret, err := secret.CreateCASecretDefault(common.DefaultFrontProxyCASecretName)
	if err != nil {
		return fmt.Errorf("unable to generate front proxy CA secret: %v", err)
	}

	newServiceAccountKeySecret, err := secret.CreateSAKeySecretDefault(common.DefaultServiceAccountKeySecretName)
	if err != nil {
		return fmt.Errorf("unable to generate service account CA secret: %v", err)
	}
	newBootstrapTokenSecret, err := secret.CreateBootstrapTokenSecret(common.DefaultBootstrapTokenSecretName)
	if err != nil {
		return fmt.Errorf("unable to generate bootstrap token CA secret: %v", err)
	}

	if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newAPIServerCASecret); err != nil {
		return fmt.Errorf("unable to create API server CA secret: %v", err)
	}
	if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newEtcdCASecret); err != nil {
		return fmt.Errorf("unable to create etcd CA secret: %v", err)
	}
	if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newFrontProxyCASecret); err != nil {
		return fmt.Errorf("unable to create front proxy CA secret: %v", err)
	}
	if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newServiceAccountKeySecret); err != nil {
		return fmt.Errorf("unable to create service account secret: %v", err)
	}
	if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newBootstrapTokenSecret); err != nil {
		return fmt.Errorf("unable to create bootstrap token secret: %v", err)
	}
	if err := state.PullFromAPIs(); err != nil {
		return fmt.Errorf("unable to sync on-disk state: %v", err)
	}

	return nil
}

func init() {
	createCmd.AddCommand(secretCmdCreate)
}
