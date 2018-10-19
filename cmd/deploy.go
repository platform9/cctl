package cmd

import (
	"fmt"

	log "github.com/platform9/cctl/pkg/logrus"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy app functionality
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Used to deploy app to the cluster",
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
		fmt.Println("Deploy called")
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
