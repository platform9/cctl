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

	"github.com/spf13/cobra"

	"github.com/platform9/cctl/pkg/util/archive"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create an archive with the current cctl state and an etcd snapshot from the cluster.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		InitState()
		// PersistentPreRuns are not chained https://github.com/spf13/cobra/issues/216
		// Therefore LogLevel must be set in all the PersistentPreRuns
		if err := log.SetLogLevelUsingString(LogLevel); err != nil {
			log.Fatalf("Unable to parse log level %s", LogLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		archivePath, err := cmd.Flags().GetString("archive")
		if err != nil {
			log.Fatalf("Unable to parse `archive`: %v", err)
		}
		snapshotPath, err := cmd.Flags().GetString("snapshot")
		if err != nil {
			log.Fatalf("Unable to parse `snapshot`: %v", err)
		}
		if err := archive.Create(archivePath, stateFilename, snapshotPath); err != nil {
			log.Fatalf("Unable to create archive: %v", err)
		}
		log.Printf("[backup] Created archive %q", archivePath)
	},
}

func init() {
	backupCmd.Flags().String("archive", "", "Path of the archive to be created.")
	backupCmd.Flags().String("snapshot", "", "Path of the etcd snapshot to include in the archive.")
	rootCmd.AddCommand(backupCmd)
}
