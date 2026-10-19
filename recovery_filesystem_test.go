package main

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecoveryDiagnosisReportsSpecialFile(t *testing.T) {
	project, root := newCLIProject(t)
	current := filepath.Join(root, "current")
	require.NoError(t, os.Remove(current))
	require.NoError(t, syscall.Mkfifo(current, 0600))
	code, _, stderr := runCLICommand(t, project, "-p", "current")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, stderr, "Type: special file")
	info, err := os.Lstat(current)
	require.NoError(t, err)
	require.NotZero(t, info.Mode()&os.ModeNamedPipe)
}

func TestRecoveryDiagnosisReportsInaccessibleEnvironment(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	project, root := newCLIProject(t)
	recoveryDiagnosisProject(t, project)
	require.NoError(t, os.Chmod(root, 0))
	t.Cleanup(func() { require.NoError(t, os.Chmod(root, 0755)) })
	diagnosis := diagnoseStartup()
	require.Error(t, diagnosis.err)
	require.Contains(t, diagnosis.err.Error(), "cannot inspect required directory")
	require.False(t, diagnosis.ready())
	diagnosis = diagnoseCurrent()
	require.Error(t, diagnosis.err)
	require.Contains(t, diagnosis.err.Error(), "cannot inspect current selection")
	require.False(t, diagnosis.ready())
	err := resolveCurrentDir()
	require.Error(t, err)
	require.Empty(t, CurrentUsed)
	require.Empty(t, LuaRocksCfg.RootDir)
}
