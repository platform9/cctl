package semver

import (
	"github.com/coreos/go-semver/semver"
)

// EqualMajorMinorPatchVersions tests if the major, minor, and patch
// versions of a and b are equal, ignoring the pre-release identifier.
func EqualMajorMinorPatchVersions(a, b semver.Version) bool {
	a.PreRelease = semver.PreRelease("")
	b.PreRelease = semver.PreRelease("")
	return a.Equal(b)
}
