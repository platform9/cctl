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

package util_test

import (
	"fmt"
	"testing"

	log "github.com/platform9/cctl/pkg/logrus"

	stateutil "github.com/platform9/cctl/pkg/state/util"
)

const (
	V1TestFile = "v1.yaml"
)

func TestVersion(t *testing.T) {
	tcs := []struct {
		name            string
		expectedVersion int
	}{
		{"v0", 0},
		{"v1", 1},
		{"v2", 2},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			testFile := fmt.Sprintf("testdata/%s.yaml", tc.name)
			actualVersion, err := stateutil.VersionFromFile(testFile)
			if err != nil {
				log.Fatalf("unable to get version from state file %s: %v", testFile, err)
			}
			if tc.expectedVersion != actualVersion {
				log.Fatalf("expected version %d, found %d", tc.expectedVersion, actualVersion)
			}
		})
	}
}
