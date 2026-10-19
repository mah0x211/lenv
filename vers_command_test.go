package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionsCommandListsAvailableRatherThanInstalledVersions(t *testing.T) {
	project, root := newCLIProject(t)
	require.NoError(t, os.RemoveAll(filepath.Join(root, "lua/5.5.1")))
	versions, err := NewVersionsFromFile(filepath.Join(root, "lua_vers.txt"))
	require.NoError(t, err)
	for i := 0; i < 12; i++ {
		version := fmt.Sprintf("5.3.%d", i+1)
		versions.Add("lua-"+version+".tar.gz", version, "", ".tar.gz")
	}
	require.NoError(t, versions.WriteFile(filepath.Join(root, "lua_vers.txt")))
	code, out, stderr := runCLICommand(t, project, "-p", "vers")
	require.Equal(t, 0, code, stderr)
	require.Contains(t, out, `list available "lua" versions:`)
	require.Contains(t, out, `list available "luajit" versions:`)
	require.Contains(t, out, `list available "luarocks" versions:`)
	for _, version := range []string{"5.5.1", "5.4.8", "v2.1", "3.13.0", "3.12.2"} {
		require.Contains(t, out, version)
	}
	require.Less(t, strings.Index(out, "5.5.1"), strings.Index(out, "5.4.8"))
	require.NotContains(t, out, "failed")
	require.NotContains(t, stderr, "Error:")
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "5.") {
			require.Less(t, len(line), 80, line)
		}
	}
}

func TestVersionsCommandReportsUnreadableMetadataAndContinues(t *testing.T) {
	for _, state := range []string{"missing", "invalid", "empty"} {
		t.Run(state, func(t *testing.T) {
			project, root := newCLIProject(t)
			file := filepath.Join(root, "lua_vers.txt")
			switch state {
			case "missing":
				require.NoError(t, os.Remove(file))
			case "invalid":
				require.NoError(t, os.WriteFile(file, []byte("{"), 0600))
			case "empty":
				require.NoError(t, NewVersions().WriteFile(file))
			}
			code, out, stderr := runCLICommand(t, project, "-p", "vers")
			require.Equal(t, 0, code, stderr)
			if state != "empty" {
				require.Contains(t, out, "failed to read version file")
			} else {
				require.NotContains(t, out, "failed")
			}
			require.NotContains(t, out, "5.4.8")
			require.Contains(t, out, "v2.1")
			require.Contains(t, out, "3.13.0")
		})
	}
}

func TestVersionsCommandReadsMetadataWithoutParsedSemVer(t *testing.T) {
	project, root := newCLIProject(t)
	require.NoError(t, os.WriteFile(filepath.Join(root, "lua_vers.txt"), []byte(`{
 "5.3.6":{"Ver":"5.3.6"},"5.4.8":{"Ver":"5.4.8"},"5.5.1":{"Ver":"5.5.1"}
 }`), 0600))
	code, out, stderr := runCLICommand(t, project, "-p", "vers")
	require.Equal(t, 0, code, stderr)
	require.Less(t, strings.Index(out, "5.5.1"), strings.Index(out, "5.4.8"))
	require.Less(t, strings.Index(out, "5.4.8"), strings.Index(out, "5.3.6"))
	for _, version := range []string{"5.5.1", "5.4.8", "5.3.6"} {
		require.Contains(t, out, version)
	}
}

func TestVersionsCommandReportsVersionNumberOverflow(t *testing.T) {
	project, root := newCLIProject(t)
	require.NoError(t, os.WriteFile(filepath.Join(root, "lua_vers.txt"), []byte(`{
 "5.3.6":{"Ver":"5.3.6"},"overflow":{"Ver":"5.99999999999999999999999999999.1"}
 }`), 0600))
	code, _, stderr := runCLICommand(t, project, "-p", "vers")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, stderr, "failed to strconv.Atoi")
	assertArchiveFile(t, filepath.Join(root, "lua/5.4.8/marker"), "5.4.8")
}
