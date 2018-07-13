package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var outputFmt string

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Used to get resources",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Unknown object %q to get. Use --help to print available options", args[0])
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.PersistentFlags().StringVar(&outputFmt, "o", "", "Output format yaml|json")
}
