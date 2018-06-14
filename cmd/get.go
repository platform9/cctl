package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Used to get resources",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Get called")
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
