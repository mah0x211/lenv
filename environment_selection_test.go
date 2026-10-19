package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirectoryOptionsAndParentEnvironmentSelection(t *testing.T) {
	project, root := newCLIProject(t)
	child := filepath.Join(project, "child", "nested")
	require.NoError(t, os.MkdirAll(child, 0755))
	code, out, stderr := runCLICommand(t, child, "path", "bin")
	require.Equal(t, 0, code, stderr)
	require.Contains(t, out, filepath.Join(root, "current/bin"))
	code, out, stderr = runCLICommand(t, project, "--global", "help")
	require.Equal(t, 0, code, stderr)
	require.Contains(t, out, "Usage:")
	require.NotContains(t, stderr, root)
}

func TestHelpWithoutDirectoryOptionLeavesProjectUntouched(t *testing.T) {
	project := t.TempDir()
	code, out, stderr := runCLICommand(t, project, "help")
	require.Equal(t, 0, code, stderr)
	require.Contains(t, out, "Usage:")
	entries, err := os.ReadDir(project)
	require.NoError(t, err)
	require.Empty(t, entries, "help must not initialize an environment")
}

func TestUnknownCommandReportsUsageWithoutChangingSelection(t *testing.T) {
	project, root := newCLIProject(t)
	code, out, stderr := runCLICommand(t, project, "-p", "not-a-command")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, out, "unknown command")
	require.Contains(t, out, "Usage:")
	link, err := os.Readlink(filepath.Join(root, "current"))
	require.NoError(t, err)
	require.Equal(t, "lua/5.4.8", link)
}

func TestLuaRocksSelectionRequiresSelectedRuntime(t *testing.T) {
	project, root := newCLIProject(t)
	require.NoError(t, os.Remove(filepath.Join(root, "current")))
	code, _, stderr := runCLICommand(t, project, "-p", "use", ":3.13.0")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, stderr, "please run `lenv use <ver>`")
	require.DirExists(t, filepath.Join(root, "lua/5.4.8/luarocks/3.13.0"))
}
