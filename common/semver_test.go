package common

import (
	"testing"

	"github.com/coreos/go-semver/semver"
)

func TestCompareMajorMinorVersions(t *testing.T) {
	tcs := []struct {
		name    string
		a       string
		b       string
		compare int
	}{
		{
			// We expect 0 here as patch versions are not compared
			name:    "equal",
			a:       "0.0.0-9+8d7d5693ad4ec9",
			b:       "0.0.2-10+g61d9a1a",
			compare: 0,
		},
		{
			name:    "lower",
			a:       "0.0.2-9+8d7d5693ad4ec9",
			b:       "0.1.2-9+8d7d5693ad4ec9",
			compare: -1,
		},
		{
			name:    "higher",
			a:       "0.1.3-9+8d7d5693ad4ec9",
			b:       "0.0.1-9+8d7d5693ad4ec9",
			compare: 1,
		},
	}
	for _, tc := range tcs {
		a := semver.New(tc.a)
		b := semver.New(tc.b)
		actual := CompareMajorMinorVersions(*a, *b)
		if actual != tc.compare {
			t.Errorf("Testcase %s failed while comparing %s and %s, expected = %d actual = %d", tc.name, a, b, tc.compare, actual)
		}
	}
}
