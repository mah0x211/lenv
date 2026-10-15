package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCurrentLinkInspectionPermissionFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can bypass directory permissions")
	}
	project, _ := recoveryProject(t, "healthy")
	root := filepath.Join(project, ".lenv")
	locked := filepath.Join(root, "locked")
	require.NoError(t, os.MkdirAll(filepath.Join(locked, "lua"), 0755))
	current := filepath.Join(root, "current")
	require.NoError(t, os.Symlink("locked/lua", current))
	require.NoError(t, os.Chmod(locked, 0000))
	defer os.Chmod(locked, 0755)
	result := recoveryProcess(t, project, []string{"yes"}, "", "-p", "setup", "--recover")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "cannot resolve current selection")
	require.NotContains(t, result.Output, "Delete and recover?")
	info, err := os.Lstat(current)
	require.NoError(t, err)
	require.NotZero(t, info.Mode()&os.ModeSymlink)
}

func TestRecoveryEscapesPromptPaths(t *testing.T) {
	project, _ := recoveryProject(t, "current-file")
	odd := filepath.Join(project, "odd-\x1b[2K\rFAKE")
	require.NoError(t, os.Mkdir(odd, 0755))
	require.NoError(t, os.Rename(filepath.Join(project, ".lenv"), filepath.Join(odd, ".lenv")))
	current := filepath.Join(odd, ".lenv", "current")
	require.NoError(t, os.Remove(current))
	target := "missing-\x1b[2K\rFAKE"
	require.NoError(t, os.Symlink(target, current))
	result := recoveryProcess(t, odd, []string{"n"}, "", "-p", "setup", "--recover")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, fmt.Sprintf("Path: %q", current))
	require.Contains(t, result.Output, fmt.Sprintf("Link target: %q", target))
	require.NotContains(t, result.Output, "\x1b")
	require.NotContains(t, result.Output, "\rFAKE")
	link, err := os.Readlink(current)
	require.NoError(t, err)
	require.Equal(t, target, link)
}

func TestRecoveryHelpWithInspectionError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can bypass directory permissions")
	}
	project, _ := recoveryProject(t, "healthy")
	locked := filepath.Join(project, ".lenv/locked")
	require.NoError(t, os.MkdirAll(filepath.Join(locked, "lua"), 0755))
	require.NoError(t, os.Symlink("locked/lua", filepath.Join(project, ".lenv/current")))
	require.NoError(t, os.Chmod(locked, 0000))
	defer os.Chmod(locked, 0755)
	result := recoveryProcess(t, project, nil, "", "-p", "help")
	require.Equal(t, 0, result.Exit, result.Output)
	require.Contains(t, result.Output, "cannot resolve current selection")
	require.Contains(t, result.Output, "Usage:")
	require.NotContains(t, result.Output, "Delete and recover?")
}
