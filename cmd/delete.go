package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteCmd represents the Delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Used to delete resources",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete called")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
