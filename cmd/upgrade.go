package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// upgradeCmd represents the status command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Used to get status of the cluster",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Upgrade called")
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
