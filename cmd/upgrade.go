package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Used to upgrade the cluster",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Upgrade called")
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
