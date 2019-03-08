/*
Copyright 2019 The cctl authors.

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
		if err = state.PullFromAPIs(); err != nil {
			log.Fatalf("unable to sync on-disk state: %v", err)
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
	return nil
}

func init() {
	createCmd.AddCommand(secretCmdCreate)
}
