package cmd

import (
	log "github.com/platform9/cctl/pkg/logrus"

	"github.com/spf13/cobra"
)

var outputFmt string

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Display one or more resources",
	Args:  cobra.MinimumNArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		InitState()
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Unknown resource %q. Use --help to print available options", args[0])
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.PersistentFlags().StringVar(&outputFmt, "o", "", "Output format yaml|json")
}
