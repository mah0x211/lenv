package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListCommandInstalledSelections(t *testing.T) {
	project, root := newCLIProject(t)
	code, out, stderr := runCLICommand(t, project, "-p", "ls")
	require.Equal(t, 0, code, stderr)
	require.NotContains(t, stderr, "failed")
	require.Contains(t, out, "5.4.8 ("+filepath.Join(root, "lua/5.4.8")+") (used)")
	require.Contains(t, out, "5.5.1 ("+filepath.Join(root, "lua/5.5.1")+")")
	require.Contains(t, out, "v2.1 ("+filepath.Join(root, "luajit/v2.1")+")")
	require.Contains(t, out, "3.12.2 (used)")
	require.Contains(t, out, "3.13.0")
	require.Less(t, strings.Index(out, "3.13.0"), strings.Index(out, "3.12.2"), out)
	require.Equal(t, 2, strings.Count(out, "(used)"), out)
}

func TestListCommandSkipsUnknownAndNonDirectoryEntries(t *testing.T) {
	project, root := newCLIProject(t)
	require.NoError(t, os.Mkdir(filepath.Join(root, "lua/manual"), 0755))
	versions, err := NewVersionsFromFile(filepath.Join(root, "lua_vers.txt"))
	require.NoError(t, err)
	versions.Add("lua-5.3.6.tar.gz", "5.3.6", "", ".tar.gz")
	require.NoError(t, versions.WriteFile(filepath.Join(root, "lua_vers.txt")))
	require.NoError(t, os.WriteFile(filepath.Join(root, "lua/5.3.6"), []byte("keep"), 0600))
	code, out, stderr := runCLICommand(t, project, "-p", "ls")
	require.Equal(t, 0, code, stderr)
	require.Contains(t, stderr, "ignore unknown version: manual")
	require.Contains(t, stderr, "found lua 5.3.6")
	require.Contains(t, stderr, "not a directory")
	require.NotContains(t, out, "manual")
	require.NotContains(t, out, "5.3.6")
	require.Contains(t, out, "v2.1")
	assertArchiveFile(t, filepath.Join(root, "lua/5.3.6"), "keep")
}

func TestListCommandReportsVersionFileFailureAndContinues(t *testing.T) {
	for _, malformed := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "invalid"}[malformed], func(t *testing.T) {
			project, root := newCLIProject(t)
			file := filepath.Join(root, "lua_vers.txt")
			if malformed {
				require.NoError(t, os.WriteFile(file, []byte("invalid JSON"), 0600))
			} else {
				require.NoError(t, os.Remove(file))
			}
			_, out, stderr := runCLICommand(t, project, "-p", "ls")
			require.Contains(t, stderr, "failed to read version file")
			require.Contains(t, out, "v2.1")
			require.NotContains(t, out, "5.4.8")
		})
	}
}

func TestListCommandReportsInvalidModulesSelection(t *testing.T) {
	for _, tc := range []struct{ name, target, message string }{
		{"regular file", "", "is not symlink"},
		{"absolute link", "absolute", "is not relative symlink"},
		{"outside link", "elsewhere", "is not relative symlink"},
		{"missing directory", "luarocks/missing/lua_modules", "is not symlink to directory"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			project, root := newCLIProject(t)
			modules := filepath.Join(root, "lua/5.4.8/lua_modules")
			require.NoError(t, os.Remove(modules))
			switch tc.target {
			case "":
				require.NoError(t, os.WriteFile(modules, []byte("keep"), 0600))
			case "absolute":
				require.NoError(t, os.Symlink(filepath.Join(root, "lua/5.4.8/luarocks/3.12.2/lua_modules"), modules))
			default:
				require.NoError(t, os.Symlink(tc.target, modules))
			}
			_, out, stderr := runCLICommand(t, project, "-p", "ls")
			require.Contains(t, stderr, tc.message)
			require.Contains(t, out, "v2.1")
			_, err := os.Lstat(modules)
			require.NoError(t, err)
		})
	}
}

func TestListCommandReportsUnreadableRocksDirectories(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	for _, selected := range []bool{false, true} {
		t.Run(map[bool]string{false: "read directory", true: "check selection"}[selected], func(t *testing.T) {
			project, root := newCLIProject(t)
			rocks := filepath.Join(root, "lua/5.4.8/luarocks")
			if !selected {
				require.NoError(t, os.Remove(filepath.Join(root, "lua/5.4.8/lua_modules")))
			}
			require.NoError(t, os.Chmod(rocks, 0))
			t.Cleanup(func() { require.NoError(t, os.Chmod(rocks, 0755)) })
			_, out, stderr := runCLICommand(t, project, "-p", "ls")
			message := "failed to readdir"
			if selected {
				message = "failed to check"
			}
			require.Contains(t, stderr, message)
			require.Contains(t, out, "v2.1")
		})
	}
}
