package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/platform9/cctl/pkg/util/archive"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the cctl state and etcd snapshot from an archive.",
	Run: func(cmd *cobra.Command, args []string) {
		archivePath, err := cmd.Flags().GetString("archive")
		if err != nil {
			log.Fatalf("Unable to parse `archive`: %v", err)
		}
		snapshotPath, err := cmd.Flags().GetString("snapshot")
		if err != nil {
			log.Fatalf("Unable to parse `snapshot`: %v", err)
		}
		if err := archive.Extract(archivePath, stateFilename, snapshotPath); err != nil {
			log.Fatalf("Unable to extract archive: %v", err)
		}
		log.Printf("[restore] Extracted etcd snapshot to %q", snapshotPath)
		log.Printf("[restore] Extracted cctl state to %q", stateFilename)
	},
}

func init() {
	restoreCmd.Flags().String("archive", "", "Path of the archive to be extracted.")
	restoreCmd.Flags().String("snapshot", "", "Path of the etcd snapshot to include in the archive.")
	rootCmd.AddCommand(restoreCmd)
}
