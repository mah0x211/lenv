package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandsReportInaccessibleVersionDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	for _, tc := range []struct{ command, message string }{
		{"ls", "failed to readdir"}, {"use", "failed to readdir"}, {"uninstall", "failed to stat"},
	} {
		t.Run(tc.command, func(t *testing.T) {
			project, root := newCLIProject(t)
			blocked := filepath.Join(root, "luajit")
			require.NoError(t, os.Chmod(blocked, 0))
			t.Cleanup(func() { require.NoError(t, os.Chmod(blocked, 0755)) })
			args := []string{"-p", tc.command}
			if tc.command != "ls" {
				args = append(args, "lj-v2.1")
			}
			code, _, stderr := runCLICommand(t, project, args...)
			require.Contains(t, stderr, tc.message)
			if tc.command != "ls" {
				require.Equal(t, 1, code, stderr)
			}
			current, err := os.Readlink(filepath.Join(root, "current"))
			require.NoError(t, err)
			require.Equal(t, "lua/5.4.8", current)
			require.NoError(t, os.Chmod(blocked, 0755))
			require.DirExists(t, filepath.Join(blocked, "v2.1"))
		})
	}
}

func TestUninstallCommandReportsRemovalFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	project, root := newCLIProject(t)
	lua := filepath.Join(root, "lua")
	version := filepath.Join(lua, "5.5.1")
	require.NoError(t, os.Chmod(version, 0555))
	require.NoError(t, os.Chmod(lua, 0555))
	t.Cleanup(func() {
		require.NoError(t, os.Chmod(lua, 0755))
		require.NoError(t, os.Chmod(version, 0755))
	})
	code, _, stderr := runCLICommand(t, project, "-p", "uninstall", "5.5.1")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, stderr, "failed to uninstall version")
	assertArchiveFile(t, filepath.Join(version, "marker"), "5.5.1")
	assertArchiveFile(t, filepath.Join(lua, "5.4.8/marker"), "5.4.8")
}

func TestListCommandReportsInaccessibleRuntime(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	project, root := newCLIProject(t)
	runtime := filepath.Join(root, "lua/5.5.1")
	require.NoError(t, os.Chmod(runtime, 0))
	t.Cleanup(func() { require.NoError(t, os.Chmod(runtime, 0755)) })
	_, out, stderr := runCLICommand(t, project, "-p", "ls")
	require.Contains(t, stderr, "permission denied")
	require.Contains(t, out, "v2.1")
}

func TestUseCommandReportsUnwritableEnvironment(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	project, root := newCLIProject(t)
	require.NoError(t, os.Chmod(root, 0555))
	t.Cleanup(func() { require.NoError(t, os.Chmod(root, 0755)) })
	code, _, stderr := runCLICommand(t, project, "-p", "use", "5.5.1")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, stderr, "failed to tempfile")
	link, err := os.Readlink(filepath.Join(root, "current"))
	require.NoError(t, err)
	require.Equal(t, "lua/5.4.8", link)
}

func TestIfMissingReportsInaccessibleInstallation(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	project, root := newCLIProject(t)
	jit := filepath.Join(root, "luajit")
	require.NoError(t, os.Chmod(jit, 0))
	t.Cleanup(func() { require.NoError(t, os.Chmod(jit, 0755)) })
	code, _, stderr := runCLICommand(t, project, "-p", "install", "--if-missing", "lj-v2.1")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, stderr, "failed to check installation directory")
}

func TestInstallReportsUnremovablePreviousInstallation(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	project, root := newCLIProject(t)
	source := filepath.Join(root, "build/lua/5.5.1")
	require.NoError(t, os.MkdirAll(source, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(source, "marker"), []byte("source"), 0600))
	lua := filepath.Join(root, "lua")
	installed := filepath.Join(lua, "5.5.1")
	require.NoError(t, os.Chmod(installed, 0555))
	require.NoError(t, os.Chmod(lua, 0555))
	t.Cleanup(func() { require.NoError(t, os.Chmod(lua, 0755)); require.NoError(t, os.Chmod(installed, 0755)) })
	code, _, stderr := runCLICommand(t, project, "-p", "install", "--rebuild", "5.5.1")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, stderr, "failed to os.RemoveAll")
	assertArchiveFile(t, filepath.Join(installed, "marker"), "5.5.1")
	assertArchiveFile(t, filepath.Join(source, "marker"), "source")
}
