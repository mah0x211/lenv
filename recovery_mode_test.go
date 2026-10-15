package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecoveryDisabledByDefault(t *testing.T) {
	for _, kind := range []string{"current-directory", "current-file", "current-dangling", "current-loop", "current-to-file", "current-intermediate-file", "rocks-file", "rocks-link", "rocks-dangling", "required-file", "required-link", "root-file"} {
		t.Run(kind, func(t *testing.T) {
			project, _ := recoveryProject(t, kind)
			path := filepath.Join(project, ".lenv", "current")
			switch kind {
			case "root-file":
				path = filepath.Join(project, ".lenv")
			case "rocks-file", "rocks-link", "rocks-dangling":
				path = filepath.Join(project, ".lenv/lua/5.5.1/luarocks")
			case "required-file", "required-link":
				path = filepath.Join(project, ".lenv/src")
			}
			before, err := os.Lstat(path)
			require.NoError(t, err)
			result := recoveryProcess(t, project, []string{"yes"}, "", "-p", "current")
			require.Equal(t, 1, result.Exit, result.Output)
			require.Contains(t, result.Output, "Run lenv -p setup --recover")
			require.Contains(t, result.Output, fmt.Sprintf("Path: %q", path))
			require.NotContains(t, result.Output, "Delete and recover?")
			after, err := os.Lstat(path)
			require.NoError(t, err)
			require.True(t, os.SameFile(before, after))
		})
	}
}

func TestSetupRecoveryFlagWithProjectOption(t *testing.T) {
	for _, option := range []string{"-p", "--project"} {
		project, _ := recoveryProject(t, "current-file")
		result := recoveryProcess(t, project, []string{"yes"}, "", option, "setup", "--recover")
		require.Equal(t, 0, result.Exit, result.Output)
		require.Contains(t, result.Output, "Delete and recover?")
		_, err := os.Lstat(filepath.Join(project, ".lenv/current"))
		require.True(t, os.IsNotExist(err))
	}
}

func TestSetupWithoutRecoveryFlagDoesNotPrompt(t *testing.T) {
	project, _ := recoveryProject(t, "current-file")
	result := recoveryProcess(t, project, []string{"yes"}, "", "-p", "setup")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "Recovery is disabled")
	require.Contains(t, result.Output, "lenv -p setup --recover")
	require.NotContains(t, result.Output, "Delete and recover?")
	_, err := os.Lstat(filepath.Join(project, ".lenv/current"))
	require.NoError(t, err)
}

func TestRecoveryFlagIsSetupOnly(t *testing.T) {
	for _, args := range [][]string{{"-p", "--recover", "setup"}, {"--recover", "setup"}} {
		project := t.TempDir()
		// Automatic discovery uses this empty real directory for the second case.
		require.NoError(t, os.Mkdir(filepath.Join(project, ".lenv"), 0755))
		result := recoveryProcess(t, project, []string{"yes"}, "", args...)
		require.Equal(t, 1, result.Exit, result.Output)
		require.NotContains(t, result.Output, "Delete and recover?")
		_, err := os.Lstat(filepath.Join(project, ".lenv/src"))
		require.True(t, os.IsNotExist(err))
	}
}

func TestRecoveryDirectoryOptionConflict(t *testing.T) {
	for _, args := range [][]string{{"--global", "--project", "setup", "--recover"}, {"-g", "-p", "setup"}} {
		project := t.TempDir()
		result := recoveryProcess(t, project, nil, "", args...)
		require.Equal(t, 1, result.Exit, result.Output)
		require.Contains(t, result.Output, "specify only one installation directory option")
		require.NotContains(t, result.Output, "Delete and recover?")
	}
}

func TestRecoveryHelpDoesNotChangeState(t *testing.T) {
	for _, kind := range []string{"current-file", "root-file"} {
		project, _ := recoveryProject(t, kind)
		result := recoveryProcess(t, project, []string{"yes"}, "", "-p", "help")
		require.Equal(t, 0, result.Exit, result.Output)
		require.Contains(t, result.Output, "Usage:")
		require.Contains(t, result.Output, "Run lenv -p setup --recover")
		require.NotContains(t, result.Output, "Delete and recover?")
	}
	project := t.TempDir()
	result := recoveryProcess(t, project, nil, "", "-p", "help")
	require.Equal(t, 0, result.Exit, result.Output)
	require.Contains(t, result.Output, "Run lenv -p setup to")
	_, err := os.Lstat(filepath.Join(project, ".lenv"))
	require.True(t, os.IsNotExist(err))
}

func TestRecoveryBlocksOtherCommandsWithoutChanges(t *testing.T) {
	for _, command := range []string{"path", "fetch", "vers", "ls", "install", "uninstall", "use", "current"} {
		t.Run(command, func(t *testing.T) {
			project := t.TempDir()
			result := recoveryProcess(t, project, nil, "", "-p", command)
			require.Equal(t, 1, result.Exit, result.Output)
			require.Contains(t, result.Output, "Run lenv -p setup to")
			require.NotContains(t, result.Output, "Delete and recover?")
			_, err := os.Lstat(filepath.Join(project, ".lenv"))
			require.True(t, os.IsNotExist(err))
		})
	}
}
