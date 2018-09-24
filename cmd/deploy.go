package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deployCmd represents the deploy app functionality
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Used to deploy app to the cluster",
	Args:  cobra.MinimumNArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		InitState()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Deploy called")
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
