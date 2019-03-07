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
	log "github.com/platform9/cctl/pkg/logrus"

	"github.com/platform9/cctl/pkg/state/v1"

	"github.com/spf13/cobra"

	spclientfake "github.com/platform9/ssh-provider/pkg/client/clientset_generated/clientset/fake"
	kubeclientfake "k8s.io/client-go/kubernetes/fake"
	clusterclientfake "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/fake"

	stateutil "github.com/platform9/cctl/pkg/state/util"
	"github.com/platform9/cctl/pkg/state/v0"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate the state file to the current version",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// PersistentPreRuns are not chained https://github.com/spf13/cobra/issues/216
		// Therefore LogLevel must be set in all the PersistentPreRuns
		if err := log.SetLogLevelUsingString(LogLevel); err != nil {
			log.Fatalf("Unable to parse log level %s", LogLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		version, err := stateutil.VersionFromFile(stateFilename)
		if err != nil {
			log.Fatalf("Error determining version of state file: %v", err)
		}
		switch version {
		case 0:
			log.Println("Migrating from v0")
			kubeClient := kubeclientfake.NewSimpleClientset()
			clusterClient := clusterclientfake.NewSimpleClientset()
			spClient := spclientfake.NewSimpleClientset()
			stateV0 := v0.NewWithFile(stateFilename, kubeClient, clusterClient, spClient)
			if err := stateV0.PushToAPIs(); err != nil {
				log.Fatalf("Error reading from state: %v", err)
			}
			stateV1 := stateutil.StateV1FromStateV0(stateV0)
			stateV2 := stateutil.StateV2FromStateV1(stateV1)
			if err := stateV2.PullFromAPIs(); err != nil {
				log.Fatalf("Error writing to state: %v", err)
			}
		case 1:
			log.Println("Migrating from v1")
			kubeClient := kubeclientfake.NewSimpleClientset()
			clusterClient := clusterclientfake.NewSimpleClientset()
			spClient := spclientfake.NewSimpleClientset()
			stateV1 := v1.NewWithFile(stateFilename, kubeClient, clusterClient, spClient)
			if err := stateV1.PushToAPIs(); err != nil {
				log.Fatalf("Error reading from state: %v", err)
			}
			stateV2 := stateutil.StateV2FromStateV1(stateV1)
			if err := stateV2.PullFromAPIs(); err != nil {
				log.Fatalf("Error writing to state: %v", err)
			}
		case 2:
			log.Println("No migration needed: already at v2")
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
