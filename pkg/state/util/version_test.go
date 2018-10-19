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
