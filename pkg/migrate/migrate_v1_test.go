package migrate

import (
	"github.com/ghodss/yaml"
	"github.com/platform9/cctl/pkg/util/migrate"
	"log"
	"strings"
	"testing"
)

// MigrateV0toV1 adds a schemaVersion field to the state file
func TestMigrateV0toV1(t *testing.T) {
	testSchemaVersionUpdate(t)
	testSchemaVersionCannotUpdate(t)
}

func testSchemaVersionUpdate(t *testing.T) {
	t.Run("Schema Version updated to 1", func(t *testing.T) {
		testYaml := StateV0toV1{
			SchemaVersion: 0,
		}
		stateBytes, err := yaml.Marshal(testYaml)
		if err != nil {
			log.Fatal(err)
		}

		migratedBytes, err := MigrateV0toV1(&stateBytes)
		if err != nil {
			log.Fatal(err)
		}
		newState := util.DecodeMigratedState(migratedBytes)
		if newState.SchemaVersion != 1 {
			log.Fatal("Migration failed.")
		}
	})
}

func testSchemaVersionCannotUpdate(t *testing.T) {
	t.Run("Schema Version updated to 1", func(t *testing.T) {
		testYaml := StateV0toV1{
			SchemaVersion: 5,
		}
		stateBytes, err := yaml.Marshal(testYaml)
		if err != nil {
			log.Fatal(err)
		}

		_, err = MigrateV0toV1(&stateBytes)
		if err != nil {
			if !strings.Contains(err.Error(), "unable to migrate state file to schemaVersion 1: schemaVersion is 5") {
				log.Fatal("Migration failed")
			}
		}
	})
}