package env

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/testutil"
	"github.com/mah0x211/lenv/internal/testutil/recovery"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentRecoveryChild(t *testing.T) {
	opts, ok := testutil.TestChildArgs(t)
	if !ok {
		return
	}
	config.InitGlobalVars("project")
	StartupRecovery.Command = SetupCommand("project", true)
	done := make(chan struct{})
	go func() {
		defer close(done)
		Setup(DiagnoseStartup(), opts)
	}()
	<-done
	os.Exit(cli.ExitCode)
}

func recoveryProcess(t *testing.T, project string, answers []string, mutation string, opts ...string) testutil.PTYResult {
	t.Helper()
	cmd := testutil.TestChild(t, "TestEnvironmentRecoveryChild", project, opts...)
	return testutil.RunPTY(t, cmd, "Delete and recover? [y/N] ", answers, filepath.Join(project, ".lenv", "current"), mutation)
}

func TestInteractiveRecoveryCancellation(t *testing.T) {
	for _, answer := range []string{"n", "", "EOF", "not-yes"} {
		t.Run(answer, func(t *testing.T) {
			project, _ := recovery.NewProject(t, "current-directory")
			result := recoveryProcess(t, project, []string{answer}, "", "--recover")
			require.Equal(t, 1, result.Exit, result.Output)
			require.Contains(t, result.Output, "recovery cancelled")
			require.Contains(t, result.Output, "removed or moved aside")
			data, err := os.ReadFile(filepath.Join(project, ".lenv/current/marker"))
			require.NoError(t, err)
			require.Equal(t, "keep", string(data))
		})
	}
}

func TestInteractiveRecoveryRechecksTarget(t *testing.T) {
	project, _ := recovery.NewProject(t, "current-file")
	result := recoveryProcess(t, project, []string{"yes"}, "replace", "--recover")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "changed while awaiting confirmation")
	data, err := os.ReadFile(filepath.Join(project, ".lenv/current"))
	require.NoError(t, err)
	require.Equal(t, "replacement", string(data))
	project, _ = recovery.NewProject(t, "current-directory")
	result = recoveryProcess(t, project, []string{"yes"}, "contents", "--recover")
	require.Equal(t, 0, result.Exit, result.Output)
}

func TestNoninteractiveRecoveryRejectsPipedApproval(t *testing.T) {
	project, _ := recovery.NewProject(t, "current-file")
	cmd := testutil.TestChild(t, "TestEnvironmentRecoveryChild", project, "--recover")
	cmd.Stdin = strings.NewReader("yes\n")
	out, err := cmd.CombinedOutput()
	require.Error(t, err)
	require.Contains(t, string(out), "interactive terminal")
	require.Contains(t, string(out), "recovery cancelled")
	data, err := os.ReadFile(filepath.Join(project, ".lenv/current"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
}

func TestInteractiveRecoveryRemovalFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can bypass directory write permissions")
	}
	project, _ := recovery.NewProject(t, "current-file")
	root := filepath.Join(project, ".lenv")
	require.NoError(t, os.Chmod(root, 0555))
	defer os.Chmod(root, 0755)
	result := recoveryProcess(t, project, []string{"yes"}, "", "--recover")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "cannot remove regular file")
	require.Contains(t, result.Output, filepath.Join(root, "current"))
	data, err := os.ReadFile(filepath.Join(root, "current"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
}

func TestRecoveryEscapesPromptPaths(t *testing.T) {
	project, _ := recovery.NewProject(t, "current-file")
	odd := filepath.Join(project, "odd-\x1b[2K\rFAKE")
	require.NoError(t, os.Mkdir(odd, 0755))
	require.NoError(t, os.Rename(filepath.Join(project, ".lenv"), filepath.Join(odd, ".lenv")))
	current := filepath.Join(odd, ".lenv", "current")
	require.NoError(t, os.Remove(current))
	target := "missing-\x1b[2K\rFAKE"
	require.NoError(t, os.Symlink(target, current))
	result := recoveryProcess(t, odd, []string{"n"}, "", "--recover")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, fmt.Sprintf("Path: %q", current))
	require.Contains(t, result.Output, fmt.Sprintf("Link target: %q", target))
	require.NotContains(t, result.Output, "\x1b")
	require.NotContains(t, result.Output, "\rFAKE")
	link, err := os.Readlink(current)
	require.NoError(t, err)
	require.Equal(t, target, link)
}

func TestStartupDiagnosisDoesNotRecoverConflicts(t *testing.T) {
	for _, kind := range []string{"current-directory", "current-file", "current-dangling", "current-loop", "current-to-file", "current-intermediate-file", "rocks-file", "rocks-link", "rocks-dangling", "required-file", "required-link", "root-file"} {
		t.Run(kind, func(t *testing.T) {
			project, _ := recovery.NewProject(t, kind)
			recovery.ConfigureProject(t, project)
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
			diagnosis := DiagnoseStartup()
			require.NoError(t, diagnosis.Err)
			require.False(t, diagnosis.Ready())
			var report bytes.Buffer
			diagnosis.Report(&report, "project")
			require.Contains(t, report.String(), "Run lenv -p setup --recover")
			require.Contains(t, report.String(), fmt.Sprintf("Path: %q", path))
			require.NotContains(t, report.String(), "Delete and recover?")
			after, err := os.Lstat(path)
			require.NoError(t, err)
			require.True(t, os.SameFile(before, after))
		})
	}
}

func TestRecoverySetupUsesMissingSnapshotSafely(t *testing.T) {
	project := t.TempDir()
	recovery.ConfigureProject(t, project)
	diagnosis := DiagnoseStartup()
	require.NoError(t, os.WriteFile(config.LenvDir, []byte("replacement"), 0600))
	input, err := os.Open(os.DevNull)
	require.NoError(t, err)
	defer input.Close()
	var output bytes.Buffer
	session := NewRecoverySession(input, &output)
	err = session.apply(diagnosis.Steps[0])
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot create required directory")
	data, err := os.ReadFile(config.LenvDir)
	require.NoError(t, err)
	require.Equal(t, "replacement", string(data))
}

func TestRecoveryDirectory(t *testing.T) {
	for _, kind := range []string{"missing", "directory", "file", "link", "dangling"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "required")
			target := filepath.Join(root, "target")
			require.NoError(t, os.Mkdir(target, 0755))
			require.NoError(t, os.WriteFile(filepath.Join(target, "keep"), []byte("keep"), 0600))
			switch kind {
			case "directory":
				require.NoError(t, os.Mkdir(path, 0755))
			case "file":
				require.NoError(t, os.WriteFile(path, []byte("keep"), 0600))
			case "link":
				require.NoError(t, os.Symlink(target, path))
			case "dangling":
				require.NoError(t, os.Symlink("missing", path))
			}
			input, err := os.CreateTemp(root, "input")
			require.NoError(t, err)
			defer input.Close()
			_, err = input.WriteString("yes\n")
			require.NoError(t, err)
			_, err = input.Seek(0, 0)
			require.NoError(t, err)
			var output bytes.Buffer
			session := NewRecoverySession(input, &output)
			session.Enabled = true
			step, inspectErr := diagnoseDirectory(path)
			err = inspectErr
			if err == nil && step != nil {
				err = session.apply(*step)
			}
			if kind == "missing" || kind == "directory" {
				require.NoError(t, err)
				info, err := os.Lstat(path)
				require.NoError(t, err)
				require.True(t, info.IsDir())
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), "recovery cancelled")
				require.Contains(t, output.String(), "interactive terminal")
				require.Contains(t, output.String(), path)
				info, statErr := os.Lstat(path)
				require.NoError(t, statErr)
				require.False(t, info.IsDir())
			}
			data, err := os.ReadFile(filepath.Join(target, "keep"))
			require.NoError(t, err)
			require.Equal(t, "keep", string(data))
		})
	}
}

func TestRecoveryTargetIdentity(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "entry")
	require.NoError(t, os.WriteFile(path, []byte("before"), 0600))
	before, err := os.Lstat(path)
	require.NoError(t, err)
	require.NoError(t, checkRecoveryTarget(path, before))
	// Contents are outside the identity check.
	require.NoError(t, os.WriteFile(path, []byte("after"), 0600))
	require.NoError(t, checkRecoveryTarget(path, before))
	require.NoError(t, os.Rename(path, path+".saved"))
	require.NoError(t, os.WriteFile(path, []byte("replacement"), 0600))
	err = checkRecoveryTarget(path, before)
	require.Error(t, err)
	require.Contains(t, err.Error(), "changed while awaiting confirmation")
	require.NoError(t, os.Remove(path))
	require.Error(t, checkRecoveryTarget(path, before))
}

func TestRecoveryKind(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file")
	require.NoError(t, os.WriteFile(file, nil, 0600))
	link := filepath.Join(root, "link")
	require.NoError(t, os.Symlink(file, link))
	for path, kind := range map[string]string{root: "directory", file: "regular file", link: "symbolic link"} {
		info, err := os.Lstat(path)
		require.NoError(t, err)
		require.Equal(t, kind, recoveryKind(info))
	}
}

func TestResolveCurrentRecoveryStates(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	oldCurrent, oldUsed, oldRocks, oldRecovery := config.CurrentDir, config.CurrentUsed, config.LuaRocksCfg, StartupRecovery
	t.Cleanup(func() {
		config.CurrentDir, config.CurrentUsed, config.LuaRocksCfg, StartupRecovery = oldCurrent, oldUsed, oldRocks, oldRecovery
	})
	input, err := os.Open(os.DevNull)
	require.NoError(t, err)
	defer input.Close()
	var output bytes.Buffer
	StartupRecovery = NewRecoverySession(input, &output)
	StartupRecovery.Enabled = true
	config.LuaRocksCfg = &config.TargetConfig{}
	lua := filepath.Join(root, "lua")
	require.NoError(t, os.Mkdir(lua, 0755))
	for _, kind := range []string{"missing", "healthy", "dangling", "loop", "file", "directory", "link-to-file", "intermediate-file"} {
		t.Run(kind, func(t *testing.T) {
			config.CurrentDir = filepath.Join(root, "current-"+kind)
			switch kind {
			case "healthy":
				require.NoError(t, os.Symlink(lua, config.CurrentDir))
			case "dangling":
				require.NoError(t, os.Symlink("missing", config.CurrentDir))
			case "loop":
				require.NoError(t, os.Symlink(config.CurrentDir, config.CurrentDir))
			case "file":
				require.NoError(t, os.WriteFile(config.CurrentDir, nil, 0600))
			case "directory":
				require.NoError(t, os.Mkdir(config.CurrentDir, 0755))
			case "intermediate-file":
				file := filepath.Join(root, "path-component")
				require.NoError(t, os.WriteFile(file, []byte("keep"), 0600))
				require.NoError(t, os.Symlink(filepath.Join(file, "child"), config.CurrentDir))
			case "link-to-file":
				target := filepath.Join(root, "plain")
				require.NoError(t, os.WriteFile(target, []byte("keep"), 0600))
				require.NoError(t, os.Symlink(target, config.CurrentDir))
			}
			config.CurrentUsed, config.LuaRocksCfg.RootDir = "stale", "stale"
			err := ResolveCurrentDir()
			switch kind {
			case "missing":
				require.NoError(t, err)
				require.Empty(t, config.CurrentUsed)
				require.Empty(t, config.LuaRocksCfg.RootDir)
			case "healthy":
				require.NoError(t, err)
				require.Equal(t, lua, config.CurrentUsed)
				info, err := os.Lstat(filepath.Join(lua, "luarocks"))
				require.NoError(t, err)
				require.True(t, info.IsDir())
			default:
				require.Error(t, err)
				require.True(t, strings.Contains(err.Error(), "cancelled"), err)
				_, err := os.Lstat(config.CurrentDir)
				require.NoError(t, err)
			}
		})
	}
}

func TestRecoveryFilesystemErrors(t *testing.T) {
	root := t.TempDir()
	input, err := os.Open(os.DevNull)
	require.NoError(t, err)
	defer input.Close()
	var output bytes.Buffer
	session := NewRecoverySession(input, &output)
	file := filepath.Join(root, "file")
	require.NoError(t, os.WriteFile(file, nil, 0600))
	_, err = diagnoseDirectory(filepath.Join(file, "child"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot inspect required directory")

	readonly := filepath.Join(root, "readonly")
	require.NoError(t, os.Mkdir(readonly, 0555))
	defer os.Chmod(readonly, 0755)
	step, err := diagnoseDirectory(filepath.Join(readonly, "missing"))
	require.NoError(t, err)
	err = session.apply(*step)
	if os.Geteuid() != 0 {
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot create required directory")
	}
}

func TestRecoveryReportsMissingLinkFromSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "current")
	require.NoError(t, os.Symlink("previous", path))
	info, err := os.Lstat(path)
	require.NoError(t, err)
	require.NoError(t, os.Remove(path))

	var output bytes.Buffer
	session := NewRecoverySession(os.Stdin, &output)
	err = session.remove(path, info, "invalid selection", "leave it unselected")
	require.ErrorIs(t, err, os.ErrNotExist)
	require.Contains(t, err.Error(), "cannot read recovery link")
}
