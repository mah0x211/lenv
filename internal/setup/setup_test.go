package setup

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/env"
	"github.com/mah0x211/lenv/internal/testutil"
	"github.com/mah0x211/lenv/internal/testutil/recovery"
	"github.com/stretchr/testify/require"
)

func TestSetupRecoveryChild(t *testing.T) {
	opts, ok := testutil.TestChildArgs(t)
	if !ok {
		return
	}
	config.InitGlobalVars("project")
	env.StartupRecovery.Command = env.SetupCommand("project", true)
	done := make(chan struct{})
	go func() {
		defer close(done)
		CmdSetup(env.DiagnoseStartup(), opts)
	}()
	<-done
	os.Exit(cli.ExitCode)
}

func recoveryProcess(t *testing.T, project string, answers []string, opts ...string) testutil.PTYResult {
	t.Helper()
	cmd := testutil.TestChild(t, "TestSetupRecoveryChild", project, opts...)
	return testutil.RunPTY(t, cmd, "Delete and recover? [y/N] ", answers, "", "")
}

func TestRecoverySetupDiagnosesPreviouslyBlockedSelection(t *testing.T) {
	project, _ := recovery.NewProject(t, "multiple")
	current := filepath.Join(project, ".lenv/current")
	require.NoError(t, os.WriteFile(current, []byte("old current"), 0600))
	result := recoveryProcess(t, project, []string{"yes", "yes", "yes"}, "--recover")
	require.Equal(t, 0, result.Exit, result.Output)
	require.Contains(t, result.Output, "The required directories are ready")
	_, err := os.Lstat(current)
	require.True(t, os.IsNotExist(err))
}

func TestRecoverySetupStopsWhenDeferredRecoveryDeclined(t *testing.T) {
	project, _ := recovery.NewProject(t, "required-file")
	current := filepath.Join(project, ".lenv/current")
	require.NoError(t, os.WriteFile(current, []byte("keep"), 0600))
	result := recoveryProcess(t, project, []string{"yes", "n"}, "--recover")
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
	project, _ := recovery.NewProject(t, "required-file")
	locked := filepath.Join(project, ".lenv/locked")
	require.NoError(t, os.MkdirAll(filepath.Join(locked, "lua"), 0755))
	current := filepath.Join(project, ".lenv/current")
	require.NoError(t, os.Symlink("locked/lua", current))
	require.NoError(t, os.Chmod(locked, 0000))
	defer os.Chmod(locked, 0755)
	result := recoveryProcess(t, project, []string{"yes"}, "--recover")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "cannot resolve current selection")
	link, err := os.Readlink(current)
	require.NoError(t, err)
	require.Equal(t, "locked/lua", link)
}

func TestInteractiveRecoveryApproval(t *testing.T) {
	kinds := []string{"current-directory", "current-file", "current-dangling", "current-loop", "current-to-file", "current-intermediate-file",
		"rocks-file", "rocks-link", "rocks-dangling", "required-file", "required-link", "root-file", "multiple"}
	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			project, kept := recovery.NewProject(t, kind)
			root := filepath.Join(project, ".lenv")
			path := filepath.Join(root, "current")
			wantType, wantAfter := "symbolic link", "leave Lua and LuaJIT unselected"
			switch {
			case strings.HasPrefix(kind, "rocks-"):
				path = filepath.Join(root, "lua/5.5.1/luarocks")
				wantAfter = "create an actual directory here"
			case kind == "root-file":
				path = root
				wantAfter = "create an actual directory here"
			case strings.HasPrefix(kind, "required-") || kind == "multiple":
				path = filepath.Join(root, "src")
				wantAfter = "create an actual directory here"
			}
			switch kind {
			case "current-directory":
				wantType = "directory"
			case "current-file", "rocks-file", "required-file", "root-file", "multiple":
				wantType = "regular file"
			}
			var linkTarget string
			if wantType == "symbolic link" {
				var err error
				linkTarget, err = os.Readlink(path)
				require.NoError(t, err)
			}
			result := recoveryProcess(t, project, []string{"yes", "y"}, "--recover")
			require.Equal(t, 0, result.Exit, result.Output)
			output := strings.ReplaceAll(result.Output, "\r\n", "\n")
			require.Contains(t, output, fmt.Sprintf("\nPath: %q\nType: %s\n", path, wantType))
			require.Contains(t, output, "\nAfter recovery: "+wantAfter+"\n")
			switch wantType {
			case "symbolic link":
				require.Contains(t, output, fmt.Sprintf("\nLink target: %q\n", linkTarget))
				require.Contains(t, output, "Delete only this link; its target will be kept.\n")
			case "directory":
				require.Contains(t, output, "Delete this directory and ALL of its contents.\n")
			case "regular file":
				require.Contains(t, output, "Delete this file and its contents.\n")
			}
			for _, name := range []string{"", "src", "lua", "luajit"} {
				info, err := os.Lstat(filepath.Join(root, name))
				require.NoError(t, err)
				require.True(t, info.IsDir())
			}
			if strings.HasPrefix(kind, "current-") {
				_, err := os.Lstat(filepath.Join(root, "current"))
				require.True(t, os.IsNotExist(err))
			}
			if strings.HasPrefix(kind, "rocks-") {
				info, err := os.Lstat(filepath.Join(root, "lua/5.5.1/luarocks"))
				require.NoError(t, err)
				require.True(t, info.IsDir())
			}
			info, err := os.Stat(kept)
			require.NoError(t, err)
			if info.IsDir() {
				kept = filepath.Join(kept, "marker")
			}
			data, err := os.ReadFile(kept)
			require.NoError(t, err)
			require.Equal(t, "keep", string(data))
		})
	}
}

func TestSetupOutputForConfirmedRecovery(t *testing.T) {
	project, _ := recovery.NewProject(t, "current-file")
	result := recoveryProcess(t, project, []string{"yes"}, "--recover")
	require.Equal(t, 0, result.Exit, result.Output)
	prompt := strings.Index(result.Output, "Delete and recover? [y/N]")
	removing := strings.Index(result.Output, "Removing regular file")
	removed := strings.Index(result.Output, "Removed regular file")
	complete := strings.Index(result.Output, "The required directories are ready.")
	require.True(t, prompt >= 0 && removing > prompt && removed > removing && complete > removed, result.Output)
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
	recovery.ConfigureProject(t, project)
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
	recovery.ConfigureProject(t, project)
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
	recovery.ConfigureProject(t, t.TempDir())
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
	recovery.ConfigureProject(t, t.TempDir())
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
	recovery.ConfigureProject(t, t.TempDir())
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
	recovery.ConfigureProject(t, t.TempDir())
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
