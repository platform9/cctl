package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Used to get version of pf9-clusteradm",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version called")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
