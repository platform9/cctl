package cmd

import (
	"fmt"

	log "github.com/platform9/cctl/pkg/logrus"
	"github.com/spf13/cobra"
)

// deleteCmd represents the Delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Used to delete resources",
	Args:  cobra.MinimumNArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		InitState()
		// PersistentPreRuns are not chained https://github.com/spf13/cobra/issues/216
		// Therefore LogLevel must be set in all the PersistentPreRuns
		if err := log.SetLogLevelUsingString(LogLevel); err != nil {
			log.Fatalf("Unable to parse log level %s", LogLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete called")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
