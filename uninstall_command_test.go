package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUninstallCommandRemovesOnlyRequestedRuntime(t *testing.T) {
	for _, tc := range []struct{ target, dir string }{
		{"5.5.1", "lua/5.5.1"}, {"5.4.8", "lua/5.4.8"}, {"lj-v2.1", "luajit/v2.1"},
	} {
		t.Run(tc.target, func(t *testing.T) {
			project, root := newCLIProject(t)
			code, out, stderr := runCLICommand(t, project, "-p", "uninstall", tc.target)
			require.Equal(t, 0, code, stderr)
			require.Contains(t, out, "has been uninstalled")
			_, err := os.Lstat(filepath.Join(root, tc.dir))
			require.True(t, os.IsNotExist(err), "%v", err)
			for _, other := range []string{"lua/5.4.8", "lua/5.5.1", "luajit/v2.1"} {
				if other != tc.dir {
					require.DirExists(t, filepath.Join(root, other))
				}
			}
			current, err := os.Readlink(filepath.Join(root, "current"))
			require.NoError(t, err)
			require.Equal(t, "lua/5.4.8", current)
			// Removing a selected runtime leaves its selection for setup --recover.
			if tc.target == "5.4.8" {
				code, _, stderr = runCLICommand(t, project, "-p", "current")
				require.Equal(t, 1, code, stderr)
				require.Contains(t, stderr, "setup")
			}
		})
	}
}

func TestUninstallCommandRejectsMissingAndNonDirectoryRuntime(t *testing.T) {
	for _, file := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "file"}[file], func(t *testing.T) {
			project, root := newCLIProject(t)
			target := filepath.Join(root, "lua/5.5.1")
			require.NoError(t, os.RemoveAll(target))
			message := "is not installed"
			if file {
				require.NoError(t, os.WriteFile(target, []byte("keep"), 0600))
				message = "not a directory"
			}
			code, _, stderr := runCLICommand(t, project, "-p", "uninstall", "5.5.1")
			require.Equal(t, 1, code, stderr)
			require.Contains(t, stderr, message)
			assertArchiveFile(t, filepath.Join(root, "lua/5.4.8/marker"), "5.4.8")
			if file {
				assertArchiveFile(t, target, "keep")
			}
		})
	}
}

func TestUninstallCommandRequiresExactVersion(t *testing.T) {
	project, root := newCLIProject(t)
	code, _, stderr := runCLICommand(t, project, "-p", "uninstall", "5.5")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, stderr, `version "5.5" does not defined`)
	assertArchiveFile(t, filepath.Join(root, "lua/5.5.1/marker"), "5.5.1")
}

func TestUninstallCommandRejectsCyclicModulesSelection(t *testing.T) {
	project, root := newCLIProject(t)
	modules := filepath.Join(root, "lua/5.4.8/lua_modules")
	require.NoError(t, os.Remove(modules))
	require.NoError(t, os.Symlink("lua_modules", modules))
	code, _, stderr := runCLICommand(t, project, "-p", "uninstall", ":3.12.2")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, stderr, "failed to EvalSymlink")
	require.DirExists(t, filepath.Join(root, "lua/5.4.8/luarocks/3.12.2"))
	link, err := os.Readlink(modules)
	require.NoError(t, err)
	require.Equal(t, "lua_modules", link)
}
