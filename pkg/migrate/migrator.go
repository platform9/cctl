package migrate

import (
	"bytes"
	"encoding/gob"
	"github.com/platform9/cctl/pkg/migrate/migrations"
	"github.com/platform9/cctl/pkg/state"
)

func ToBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Migrate(stateBytes *[]byte, version state.SchemaVersion) (interface{}, error) {
	return migrations.MigrateV0toV1(stateBytes, version)
}
