package cmd

import (
	"github.com/spf13/cobra"
)

// recoverCmd represents the status command
var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Used to recover the cluster",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(recoverCmd)
}
