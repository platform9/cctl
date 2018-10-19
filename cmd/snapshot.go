package cmd

import (
	log "github.com/platform9/cctl/pkg/logrus"
	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Used to get a snapshot",
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
	},
}

func init() {
	rootCmd.AddCommand(snapshotCmd)
}
