/*
Copyright 2019 The cctl authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
