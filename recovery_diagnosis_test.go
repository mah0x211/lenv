package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecoverySetupUsesStartupDiagnosis(t *testing.T) {
	oldRecovery := startupRecovery
	t.Cleanup(func() { startupRecovery = oldRecovery })
	project, _ := recoveryProject(t, "current-file")
	recoveryDiagnosisProject(t, project)
	diagnosis := diagnoseStartup()
	require.NoError(t, diagnosis.err)
	require.Len(t, diagnosis.steps, 1)
	require.True(t, diagnosis.steps[0].info.Mode().IsRegular())
	require.NoError(t, os.Rename(CurrentDir, CurrentDir+".saved"))
	require.NoError(t, os.Mkdir(CurrentDir, 0755))
	marker := filepath.Join(CurrentDir, "keep")
	require.NoError(t, os.WriteFile(marker, []byte("keep"), 0600))
	input, err := os.Open(os.DevNull)
	require.NoError(t, err)
	defer input.Close()
	var output bytes.Buffer
	startupRecovery = newRecoverySession(input, &output)
	ExitCode = 0
	done := make(chan struct{})
	go func() {
		defer close(done)
		CmdSetup(diagnosis, nil)
	}()
	<-done
	require.Equal(t, 1, ExitCode)
	// A new diagnosis would report a directory; setup must retain the snapshot.
	require.Contains(t, output.String(), "Type: regular file")
	require.NotContains(t, output.String(), "Type: directory")
	require.FileExists(t, marker)
}

func TestRecoverySetupUsesMissingSnapshotSafely(t *testing.T) {
	project := t.TempDir()
	recoveryDiagnosisProject(t, project)
	diagnosis := diagnoseStartup()
	require.NoError(t, os.WriteFile(LenvDir, []byte("replacement"), 0600))
	input, err := os.Open(os.DevNull)
	require.NoError(t, err)
	defer input.Close()
	var output bytes.Buffer
	session := newRecoverySession(input, &output)
	err = session.apply(diagnosis.steps[0])
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot create required directory")
	data, err := os.ReadFile(LenvDir)
	require.NoError(t, err)
	require.Equal(t, "replacement", string(data))
}

func TestRecoverySetupDiagnosesPreviouslyBlockedSelection(t *testing.T) {
	project, _ := recoveryProject(t, "multiple")
	current := filepath.Join(project, ".lenv/current")
	require.NoError(t, os.WriteFile(current, []byte("old current"), 0600))
	result := recoveryProcess(t, project, []string{"yes", "yes", "yes"}, "", "-p", "setup", "--recover")
	require.Equal(t, 0, result.Exit, result.Output)
	require.Contains(t, result.Output, "The required directories are ready")
	_, err := os.Lstat(current)
	require.True(t, os.IsNotExist(err))
}

func TestRecoverySetupRejectsUnknownOptionWithoutChanges(t *testing.T) {
	project := t.TempDir()
	result := recoveryProcess(t, project, nil, "", "-p", "setup", "--unknown")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "unknown setup option")
	_, err := os.Lstat(filepath.Join(project, ".lenv"))
	require.True(t, os.IsNotExist(err))
}

func TestRecoveryMissingRocksRequiresSetupAtStartup(t *testing.T) {
	project, _ := recoveryProject(t, "rocks-file")
	rocks := filepath.Join(project, ".lenv/lua/5.5.1/luarocks")
	require.NoError(t, os.Remove(rocks))
	result := recoveryProcess(t, project, nil, "", "-p", "current")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "Run lenv -p setup to")
	_, err := os.Lstat(rocks)
	require.True(t, os.IsNotExist(err))
	result = recoveryProcess(t, project, nil, "", "-p", "setup")
	require.Equal(t, 0, result.Exit, result.Output)
	require.NotContains(t, result.Output, "Delete and recover?")
	info, err := os.Lstat(rocks)
	require.NoError(t, err)
	require.True(t, info.IsDir())
	result = recoveryProcess(t, project, nil, "", "-p", "current")
	require.Equal(t, 0, result.Exit, result.Output)
	require.NotContains(t, result.Output, "Run lenv")
}

func TestRecoverySetupStopsWhenDeferredRecoveryDeclined(t *testing.T) {
	project, _ := recoveryProject(t, "required-file")
	current := filepath.Join(project, ".lenv/current")
	require.NoError(t, os.WriteFile(current, []byte("keep"), 0600))
	result := recoveryProcess(t, project, []string{"yes", "n"}, "", "-p", "setup", "--recover")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "recovery cancelled")
	data, err := os.ReadFile(current)
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
	info, err := os.Lstat(filepath.Join(project, ".lenv/src"))
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestRecoverySetupStopsWhenDeferredInspectionFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can bypass directory permissions")
	}
	project, _ := recoveryProject(t, "required-file")
	locked := filepath.Join(project, ".lenv/locked")
	require.NoError(t, os.MkdirAll(filepath.Join(locked, "lua"), 0755))
	current := filepath.Join(project, ".lenv/current")
	require.NoError(t, os.Symlink("locked/lua", current))
	require.NoError(t, os.Chmod(locked, 0000))
	defer os.Chmod(locked, 0755)
	result := recoveryProcess(t, project, []string{"yes"}, "", "-p", "setup", "--recover")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "cannot resolve current selection")
	link, err := os.Readlink(current)
	require.NoError(t, err)
	require.Equal(t, "locked/lua", link)
}
