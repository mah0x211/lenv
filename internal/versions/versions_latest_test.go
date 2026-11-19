package versions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func latestVersions(t *testing.T, kind string) *Versions {
	t.Helper()
	versions := NewVersions()
	switch kind {
	case "prerelease":
		require.True(t, versions.Add("lua-5.5.0-rc1.tar.gz", "5.5.0-rc1", "", ".tar.gz"))
	case "branch":
		versions.AddBranch("v2.1", "origin")
	case "empty":
	default:
		t.Fatalf("unknown fixture %q", kind)
	}
	return versions
}

func TestLatestWithoutStableCandidates(t *testing.T) {
	for _, kind := range []string{"empty", "prerelease", "branch"} {
		t.Run(kind, func(t *testing.T) {
			versions := latestVersions(t, kind)
			require.NotPanics(t, func() { require.Nil(t, versions.PickItem("latest")) })
			require.Nil(t, versions.PickItem("5."))
			require.Nil(t, versions.PickItem("missing"))
			if kind == "prerelease" {
				require.NotNil(t, versions.PickItem("5.5.0-rc1"))
			}
			if kind == "branch" {
				require.NotNil(t, versions.PickItem("v2.1"))
			}
		})
	}
}

func TestLatestPrefersStableCandidates(t *testing.T) {
	versions := latestVersions(t, "prerelease")
	versions.AddBranch("v2.1", "origin")
	require.True(t, versions.Add("lua-5.4.8.tar.gz", "5.4.8", "", ".tar.gz"))
	require.True(t, versions.Add("lua-5.4.9.tar.gz", "5.4.9", "", ".tar.gz"))
	require.Equal(t, "5.4.9", versions.PickItem("latest").Ver)
	require.Equal(t, "5.4.9", versions.PickItem("5.4.").Ver)
	require.Equal(t, "5.4.8", versions.PickItem("5.4.8").Ver)
}
