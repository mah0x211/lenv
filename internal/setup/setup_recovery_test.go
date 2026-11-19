package setup

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/env"
	"github.com/mah0x211/lenv/internal/testutil"
	"github.com/stretchr/testify/require"
)

func recoveryDiagnosisProject(t *testing.T, project string) {
	t.Helper()
	oldRoot, oldSrc, oldCurrent, oldUsed := config.LenvDir, config.SrcDir, config.CurrentDir, config.CurrentUsed
	oldLua, oldJit, oldRocks, oldExit := config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg, cli.ExitCode
	t.Cleanup(func() {
		config.LenvDir, config.SrcDir, config.CurrentDir, config.CurrentUsed = oldRoot, oldSrc, oldCurrent, oldUsed
		config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg, cli.ExitCode = oldLua, oldJit, oldRocks, oldExit
	})
	config.LenvDir = filepath.Join(project, ".lenv")
	config.SrcDir, config.CurrentDir = filepath.Join(config.LenvDir, "src"), filepath.Join(config.LenvDir, "current")
	config.LuaCfg = &config.TargetConfig{RootDir: filepath.Join(config.LenvDir, "lua")}
	config.LuaJitCfg = &config.TargetConfig{RootDir: filepath.Join(config.LenvDir, "luajit")}
	config.LuaRocksCfg = &config.TargetConfig{}
}

func createRecoveryProjectDirectories(t *testing.T) {
	t.Helper()
	for _, path := range []string{config.SrcDir, filepath.Join(config.LuaCfg.RootDir, "5.5.1"), config.LuaJitCfg.RootDir} {
		require.NoError(t, os.MkdirAll(path, 0755))
	}
}

func runSetup(t *testing.T, opts ...string) (string, bool) {
	t.Helper()
	input, err := os.Open(os.DevNull)
	require.NoError(t, err)
	defer input.Close()
	oldRecovery := env.StartupRecovery
	defer func() { env.StartupRecovery = oldRecovery }()
	cli.ExitCode = 0
	stdout, stderr, completed := testutil.CaptureCommandOutput(t, func() {
		env.StartupRecovery = env.NewRecoverySession(input, os.Stderr)
		env.StartupRecovery.Command = env.SetupCommand("project", true)
		CmdSetup(env.DiagnoseStartup(), opts)
	})
	return stdout + stderr, completed
}

func TestRecoverySetupUsesStartupDiagnosis(t *testing.T) {
	oldRecovery := env.StartupRecovery
	t.Cleanup(func() { env.StartupRecovery = oldRecovery })
	project, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	recoveryDiagnosisProject(t, project)
	createRecoveryProjectDirectories(t)
	require.NoError(t, os.WriteFile(config.CurrentDir, []byte("keep"), 0600))
	diagnosis := env.DiagnoseStartup()
	require.NoError(t, diagnosis.Err)
	require.Len(t, diagnosis.Steps, 1)
	require.True(t, diagnosis.Steps[0].Info.Mode().IsRegular())
	require.NoError(t, os.Rename(config.CurrentDir, config.CurrentDir+".saved"))
	require.NoError(t, os.Mkdir(config.CurrentDir, 0755))
	marker := filepath.Join(config.CurrentDir, "keep")
	require.NoError(t, os.WriteFile(marker, []byte("keep"), 0600))
	input, err := os.Open(os.DevNull)
	require.NoError(t, err)
	defer input.Close()
	var output bytes.Buffer
	env.StartupRecovery = env.NewRecoverySession(input, &output)
	cli.ExitCode = 0
	done := make(chan struct{})
	go func() {
		defer close(done)
		CmdSetup(diagnosis, nil)
	}()
	<-done
	require.Equal(t, 1, cli.ExitCode)
	// A new diagnosis would report a directory; setup must retain the snapshot.
	require.Contains(t, output.String(), "Type: regular file")
	require.NotContains(t, output.String(), "Type: directory")
	require.FileExists(t, marker)
}

func TestRecoveryCreatesMissingDirectories(t *testing.T) {
	project := t.TempDir()
	recoveryDiagnosisProject(t, project)
	output, completed := runSetup(t)
	require.True(t, completed, output)
	require.Equal(t, 0, cli.ExitCode, output)
	require.NotContains(t, output, "Delete and recover?")
	for _, dir := range []string{".lenv", ".lenv/src", ".lenv/lua", ".lenv/luajit"} {
		info, err := os.Lstat(filepath.Join(project, dir))
		require.NoError(t, err)
		require.True(t, info.IsDir())
	}
}

func TestSetupWithoutRecoveryFlagDoesNotPrompt(t *testing.T) {
	recoveryDiagnosisProject(t, t.TempDir())
	createRecoveryProjectDirectories(t)
	require.NoError(t, os.WriteFile(config.CurrentDir, []byte("keep"), 0600))
	output, completed := runSetup(t)
	require.False(t, completed, output)
	require.Equal(t, 1, cli.ExitCode, output)
	require.Contains(t, output, "Recovery is disabled")
	require.Contains(t, output, "lenv -p setup --recover")
	require.NotContains(t, output, "Delete and recover?")
	_, err := os.Lstat(config.CurrentDir)
	require.NoError(t, err)
}

func TestSetupOutputForCreationFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can bypass directory permissions")
	}
	recoveryDiagnosisProject(t, t.TempDir())
	createRecoveryProjectDirectories(t)
	root := config.LenvDir
	require.NoError(t, os.Remove(config.SrcDir))
	require.NoError(t, os.Chmod(root, 0555))
	t.Cleanup(func() { require.NoError(t, os.Chmod(root, 0755)) })

	output, completed := runSetup(t)
	require.False(t, completed, output)
	require.Equal(t, 1, cli.ExitCode, output)
	require.Contains(t, output, "Creating required directory")
	require.Contains(t, output, "Error: cannot create required directory")
	require.NotContains(t, output, "Created required directory")
	require.NotContains(t, output, "The required directories are ready.")
}

func TestRecoverySetupRejectsUnknownOptionWithoutChanges(t *testing.T) {
	recoveryDiagnosisProject(t, t.TempDir())
	oldExit := cli.OsExit
	t.Cleanup(func() { cli.OsExit = oldExit })
	exitCode := 0
	cli.OsExit = func(code int) {
		exitCode = code
		runtime.Goexit()
	}

	output, completed := runSetup(t, "--unknown")
	require.False(t, completed, output)
	require.Equal(t, 1, exitCode, output)
	require.Contains(t, output, "unknown setup option")
	_, err := os.Lstat(config.LenvDir)
	require.True(t, os.IsNotExist(err))
}

func TestCurrentLinkInspectionPermissionFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can bypass directory permissions")
	}
	recoveryDiagnosisProject(t, t.TempDir())
	createRecoveryProjectDirectories(t)
	locked := filepath.Join(config.LenvDir, "locked")
	require.NoError(t, os.MkdirAll(filepath.Join(locked, "lua"), 0755))
	require.NoError(t, os.Symlink("locked/lua", config.CurrentDir))
	require.NoError(t, os.Chmod(locked, 0000))
	t.Cleanup(func() { require.NoError(t, os.Chmod(locked, 0755)) })

	output, completed := runSetup(t, "--recover")
	require.False(t, completed, output)
	require.Equal(t, 1, cli.ExitCode, output)
	require.Contains(t, output, "cannot resolve current selection")
	require.NotContains(t, output, "Delete and recover?")
	info, err := os.Lstat(config.CurrentDir)
	require.NoError(t, err)
	require.NotZero(t, info.Mode()&os.ModeSymlink)
}
