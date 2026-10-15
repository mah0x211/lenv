package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupOutputForMissingDirectories(t *testing.T) {
	project := t.TempDir()
	result := recoveryProcess(t, project, nil, "", "-p", "setup")
	require.Equal(t, 0, result.Exit, result.Output)
	check := strings.Index(result.Output, "Checking required directories and current selection...")
	creating := strings.Index(result.Output, "Creating required directory")
	created := strings.Index(result.Output, "Created required directory")
	complete := strings.Index(result.Output, "The required directories are ready.")
	require.True(t, check >= 0 && creating > check && created > creating && complete > created, result.Output)
	require.NotContains(t, result.Output, "Error:")
}

func TestSetupOutputForConflictWithoutRecovery(t *testing.T) {
	project, _ := recoveryProject(t, "current-file")
	result := recoveryProcess(t, project, nil, "", "-p", "setup")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "Checking required directories and current selection...")
	require.Contains(t, result.Output, "Error: Cannot initialize lenv:")
	require.Contains(t, result.Output, "Nothing was deleted at this path.")
	require.Contains(t, result.Output, "Run lenv -p setup --recover from an interactive terminal")
	require.Contains(t, result.Output, "move the conflicting entry aside and run lenv -p setup again")
	require.NotContains(t, result.Output, "Delete and recover?")
	require.NotContains(t, result.Output, "Removing regular file")
	require.NotContains(t, result.Output, "The required directories are ready.")
	data, err := os.ReadFile(filepath.Join(project, ".lenv/current"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
}

func TestSetupOutputForConfirmedRecovery(t *testing.T) {
	project, _ := recoveryProject(t, "current-file")
	result := recoveryProcess(t, project, []string{"yes"}, "", "-p", "setup", "--recover")
	require.Equal(t, 0, result.Exit, result.Output)
	prompt := strings.Index(result.Output, "Delete and recover? [y/N]")
	removing := strings.Index(result.Output, "Removing regular file")
	removed := strings.Index(result.Output, "Removed regular file")
	complete := strings.Index(result.Output, "The required directories are ready.")
	require.True(t, prompt >= 0 && removing > prompt && removed > removing && complete > removed, result.Output)
}

func TestSetupOutputForCreationFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can bypass directory permissions")
	}
	project, _ := recoveryProject(t, "healthy")
	root := filepath.Join(project, ".lenv")
	require.NoError(t, os.Remove(filepath.Join(root, "src")))
	require.NoError(t, os.Chmod(root, 0555))
	defer os.Chmod(root, 0755)
	result := recoveryProcess(t, project, nil, "", "-p", "setup")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "Creating required directory")
	require.Contains(t, result.Output, "Error: cannot create required directory")
	require.NotContains(t, result.Output, "Created required directory")
	require.NotContains(t, result.Output, "The required directories are ready.")
}
