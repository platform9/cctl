package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// addonCmd represents the addon command
var addonCmd = &cobra.Command{
	Use:   "addon",
	Short: "Used to deploy addons",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Addon called")
	},
}

func init() {
	deployCmd.AddCommand(addonCmd)
	addonCmd.Flags().String("type", "metallb", "Deploy addon specified by type")
	addonCmd.Flags().String("yaml", "yamls/metallb.yaml", "File path respresenting the YAML spec")
}
