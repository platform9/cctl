package cmd

import (
	"bytes"
	"encoding/gob"
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

func DecodeMigratedState(any []byte) statePkg.State {
	buf := bytes.NewBuffer(any)
	dec := gob.NewDecoder(buf)
	var thisState statePkg.State
	err := dec.Decode(&thisState)
	if err != nil {
		log.Fatal("decode:", err)
	}
	return thisState
}

func _migrate(stateBytes *[]byte, version statePkg.SchemaVersion) ([]byte, error) {
	return migrator.MigrateV0toV1(stateBytes, version)
}

func Migrate() {
	log.Printf("Migrating state file to new schema")
	file, err := os.OpenFile(state.Filename, os.O_RDONLY|os.O_CREATE, statePkg.FileMode)
	if err != nil {
		log.Fatal("Unable to open %q: %v", state.Filename, err)
	}

	defer file.Close()
	stateBytes, err := ioutil.ReadAll(file)

	migratedBytes, err := _migrate(&stateBytes, statePkg.Version)
	if err != nil {
		log.Fatal(err)
	}

	newState := DecodeMigratedState(migratedBytes)
	newState.KubeClient = state.KubeClient
	newState.ClusterClient = state.ClusterClient
	newState.SPClient = state.SPClient
	newState.Filename = stateFilename

	err = statePkg.CreateObjects(&newState)
	if err != nil {
		log.Fatalf("Unable to sync API objects: %v", err)
	}

	err = newState.PullFromAPIs()
	if err != nil {
		log.Fatalf("Unable to write to on-disk state: %v", err)
	}
	log.Printf("Finished migrating state file to new schema")

}
