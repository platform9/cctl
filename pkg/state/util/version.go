package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
)

type fakeState struct {
	SchemaVersion int `json:"schemaVersion"`
}

func VersionFromFile(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, fmt.Errorf("unable to open %q: %v", filename, err)
	}
	defer file.Close()
	return Version(file)
}

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
