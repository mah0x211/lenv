package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLuaJitBuildFailureDoesNotInstall(t *testing.T) {
	preserveWorkingDirectory(t)
	source := t.TempDir()
	prefix := filepath.Join(t.TempDir(), "installed")
	require.NoError(t, os.Chdir(source))
	// An incomplete source tree makes the real build command fail.
	require.Error(t, installLuaJit(prefix, nil))
	_, err := os.Lstat(prefix)
	require.True(t, os.IsNotExist(err), "build failure created an installation: %v", err)
}

func TestLuaJitInstallFailureIsReported(t *testing.T) {
	preserveWorkingDirectory(t)
	source := t.TempDir()
	prefix := filepath.Join(t.TempDir(), "installed")
	require.NoError(t, os.WriteFile(filepath.Join(source, "Makefile"), []byte("all: built\nbuilt:\n\tprintf built > built\n"), 0600))
	require.NoError(t, os.Chdir(source))
	// The source builds, but has no install target.
	require.Error(t, installLuaJit(prefix, []string{"MACOSX_DEPLOYMENT_TARGET=10.13"}))
	assertArchiveFile(t, filepath.Join(source, "built"), "built")
	_, err := os.Lstat(prefix)
	require.True(t, os.IsNotExist(err), "failed install created an installation: %v", err)
}
