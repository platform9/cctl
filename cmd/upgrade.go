package cmd

import (
	"fmt"

	log "github.com/platform9/cctl/pkg/logrus"
	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Used to upgrade the cluster",
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
		fmt.Println("Upgrade called")
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
