package common

import (
	"github.com/coreos/go-semver/semver"
)

// CompareMajorMinorVersions compares Major and Minor portions of semver versions
func CompareMajorMinorVersions(a, b semver.Version) int {
	a.PreRelease = semver.PreRelease("")
	b.PreRelease = semver.PreRelease("")
	//Ignore
	a.Patch = 0
	b.Patch = 0
	return a.Compare(b)
}
