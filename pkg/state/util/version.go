package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
)

// fakeState is a subset of the state schema used to parse the schema version
type fakeState struct {
	SchemaVersion int `json:"schemaVersion"`
}

// VersionFromFile returns the state version of the file
func VersionFromFile(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, fmt.Errorf("unable to open %q: %v", filename, err)
	}
	defer file.Close()
	return Version(file)
}

// Version returns the state version
func Version(r io.Reader) (int, error) {
	stateBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, fmt.Errorf("unable to read state: %v", err)
	}
	fs := fakeState{}
	if err := yaml.Unmarshal(stateBytes, &fs); err != nil {
		return 0, fmt.Errorf("unable to unmarshal state: %v", err)
	}
	return fs.SchemaVersion, nil
}
