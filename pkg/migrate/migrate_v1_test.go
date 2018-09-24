package migrate

import (
	"github.com/ghodss/yaml"
	"github.com/platform9/cctl/pkg/util/migrate"
	"log"
	"strings"
	"testing"
)

func marshal(s StateV0toV1) []byte {
	stateBytes, err := yaml.Marshal(s)
	if err != nil {
		log.Fatal(err)
	}
	return stateBytes
}

func decode(b []byte) {
	newState := util.DecodeMigratedState(b)
	if newState.SchemaVersion != 1 {
		log.Fatal("Migration failed.")
	}
}

// MigrateV0toV1 adds a schemaVersion field to the state file
func TestMigrateV0toV1(t *testing.T) {
	testSchemaVersionUpdate(t)
	testHigherSchemaVersion(t)
	testSameSchemaVersion(t)
}

func testSchemaVersionUpdate(t *testing.T) {
	t.Run("Update schemaVersion to 1", func(t *testing.T) {
		testYaml := StateV0toV1{
			SchemaVersion: 0,
		}
		stateBytes := marshal(testYaml)

		migratedBytes, err := MigrateV0toV1(&stateBytes)
		if err != nil {
			log.Fatal(err)
		}

		decode(migratedBytes)
	})
}

func testHigherSchemaVersion(t *testing.T) {
	t.Run("Cannot update schemaVersion > 1", func(t *testing.T) {
		testYaml := StateV0toV1{
			SchemaVersion: 5,
		}
		stateBytes := marshal(testYaml)

		_, err := MigrateV0toV1(&stateBytes)
		if err != nil {
			if !strings.Contains(err.Error(), "unable to migrate state file to schemaVersion 1: schemaVersion is 5") {
				log.Fatal("Migration failed")
			}
		}
	})
}

func testSameSchemaVersion(t *testing.T) {
	t.Run("Update same schemaVersion", func(t *testing.T) {
		testYaml := StateV0toV1{
			SchemaVersion: 1,
		}
		stateBytes := marshal(testYaml)

		migratedBytes, err := MigrateV0toV1(&stateBytes)
		if err != nil {
			log.Fatal(err)
		}

		decode(migratedBytes)
	})
}
