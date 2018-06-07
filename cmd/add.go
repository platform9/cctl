package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Used to add resources",
	Args:  cobra.MinimumNArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("add called")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
