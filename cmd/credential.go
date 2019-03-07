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
	"io/ioutil"

	log "github.com/platform9/cctl/pkg/logrus"

	"github.com/platform9/cctl/common"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var credentialCmdCreate = &cobra.Command{
	Use:   "credential",
	Short: "Create new SSH credential",
	Run: func(cmd *cobra.Command, args []string) {
		privateKeyFilename := cmd.Flag("private-key").Value.String()
		privateKeyBytes, err := ioutil.ReadFile(privateKeyFilename)
		if err != nil {
			log.Fatalf("Failed to read private key from %q: %v", privateKeyFilename, err)
		}
		secret := corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              common.DefaultSSHCredentialSecretName,
				Namespace:         common.DefaultNamespace,
				CreationTimestamp: metav1.Now(),
			},
			Data: map[string][]byte{
				"username":       []byte(cmd.Flag("user").Value.String()),
				"ssh-privatekey": privateKeyBytes,
			},
		}
		if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(&secret); err != nil {
			if apierrors.IsAlreadyExists(err) {
				log.Fatalf("Credential already exists. To create a new credential, first delete the existing one.")
			}
			log.Fatalf("Unable to create ssh credential secret: %v", err)
		}
		log.Printf("Created ssh credential: user %q and private key %q", cmd.Flag("user").Value.String(), cmd.Flag("private-key").Value.String())
		if err := state.PullFromAPIs(); err != nil {
			log.Fatalf("Unable to sync on-disk state: %v", err)
		}
	},
}

var credentialCmdDelete = &cobra.Command{
	Use:   "credential",
	Short: "Delete SSH credential",
	Run: func(cmd *cobra.Command, args []string) {
		if err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Delete(common.DefaultSSHCredentialSecretName, &metav1.DeleteOptions{}); err != nil {
			if apierrors.IsNotFound(err) {
				log.Fatal("SSH credential dooes not exist.")
			}
			log.Fatalf("Unable to delete ssh credential secret: %v", err)
		}
		log.Println("Deleted ssh credential")
		if err := state.PullFromAPIs(); err != nil {
			log.Fatalf("Unable to sync on-disk state: %v", err)
		}
	},
}

func init() {
	createCmd.AddCommand(credentialCmdCreate)
	credentialCmdCreate.Flags().String("user", "root", "SSH username")
	credentialCmdCreate.Flags().String("private-key", "", "SSH privateKey file location")
	credentialCmdCreate.MarkFlagRequired("private-key")

	deleteCmd.AddCommand(credentialCmdDelete)
}
