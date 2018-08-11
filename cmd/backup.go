package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/platform9/cctl/pkg/util/archive"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create an archive with the current cctl state and an etcd snapshot from the cluster.",
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
