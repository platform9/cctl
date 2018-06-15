package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Used to get status of the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Status called")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
