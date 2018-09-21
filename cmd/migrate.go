package cmd

import (
	"github.com/ghodss/yaml"
	migrator "github.com/platform9/cctl/pkg/migrate"
	statePkg "github.com/platform9/cctl/pkg/state"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
)

// upgradeCmd represents the upgrade command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate state file to a newer schema",
	Run: func(cmd *cobra.Command, args []string) {
		// This is a noop. The actual command is called in root.go
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func Migrate() {
	log.Printf("Migrating state file to new schema")
	file, err := os.OpenFile(state.Filename, os.O_RDONLY|os.O_CREATE, statePkg.FileMode)
	if err != nil {
		log.Fatal("Unable to open %q: %v", state.Filename, err)
	}

	defer file.Close()
	stateBytes, err := ioutil.ReadAll(file)

	migratedState, err := migrator.Migrate(&stateBytes, statePkg.Version)
	if err != nil {
		log.Fatal(err)
	}

	migratedBytes, err := migrator.ToBytes(migratedState)
	if err != nil {
		log.Fatal(err)
	}
	yaml.Unmarshal(migratedBytes, state)
	if err != nil {
		log.Fatal("Error unmarshalling migrated state file")
	}
	log.Printf("Finished migrating state file to new schema")
	state.PullFromAPIs()
}
