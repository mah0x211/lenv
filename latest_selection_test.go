package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func latestTestVersions(t *testing.T, kind string) *Versions {
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
			versions := latestTestVersions(t, kind)
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
	versions := latestTestVersions(t, "prerelease")
	versions.AddBranch("v2.1", "origin")
	require.True(t, versions.Add("lua-5.4.8.tar.gz", "5.4.8", "", ".tar.gz"))
	require.True(t, versions.Add("lua-5.4.9.tar.gz", "5.4.9", "", ".tar.gz"))
	require.Equal(t, "5.4.9", versions.PickItem("latest").Ver)
	require.Equal(t, "5.4.9", versions.PickItem("5.4.").Ver)
	require.Equal(t, "5.4.8", versions.PickItem("5.4.8").Ver)
}

func TestLatestSelectionCLIError(t *testing.T) {
	for _, target := range []struct{ kind, version, file string }{
		{"empty", "latest", "lua_vers.txt"},
		{"prerelease", "latest", "lua_vers.txt"},
		{"branch", "lj-latest", "luajit_vers.txt"},
		{"prerelease", ":latest", "luarocks_vers.txt"},
	} {
		for _, command := range []string{"install", "use"} {
			t.Run(command+"/"+target.file+"/"+target.kind, func(t *testing.T) {
				project := t.TempDir()
				root := filepath.Join(project, ".lenv")
				runtime := filepath.Join(root, "lua", "5.4.8")
				for _, dir := range []string{filepath.Join(root, "src"), filepath.Join(root, "luajit"), filepath.Join(runtime, "luarocks")} {
					require.NoError(t, os.MkdirAll(dir, 0755))
				}
				marker := filepath.Join(runtime, "marker")
				require.NoError(t, os.WriteFile(marker, []byte("installed"), 0600))
				current := filepath.Join(root, "current")
				require.NoError(t, os.Symlink("lua/5.4.8", current))
				require.NoError(t, latestTestVersions(t, target.kind).WriteFile(filepath.Join(root, target.file)))
				code, output := startupDiagnosisProcess(t, project, "--project", command, target.version)
				require.Equal(t, 1, code, output)
				require.Contains(t, output, "version \"latest\" does not defined")
				require.Contains(t, output, target.file)
				require.NotContains(t, output, "panic:")
				data, err := os.ReadFile(marker)
				require.NoError(t, err)
				require.Equal(t, "installed", string(data))
				selection, err := os.Readlink(current)
				require.NoError(t, err)
				require.Equal(t, "lua/5.4.8", selection)
				_, err = os.Stat(filepath.Join(runtime, "lua_modules"))
				require.True(t, os.IsNotExist(err))
			})
		}
	}
}
