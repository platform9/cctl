package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// backupCmd represents the status command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Used to obtain backup of the cluster state including etcd state",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Backup cluster called")
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
