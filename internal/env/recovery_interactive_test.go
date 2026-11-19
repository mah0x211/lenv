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

func newInteractiveRecoveryProject(t *testing.T, kind string) (string, string) {
	t.Helper()
	project, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	root := filepath.Join(project, ".lenv")
	target := filepath.Join(project, "kept")
	require.NoError(t, os.Mkdir(target, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "marker"), []byte("keep"), 0600))
	if kind == "root-file" {
		require.NoError(t, os.WriteFile(root, nil, 0600))
		return project, target
	}
	for _, dir := range []string{"src", "lua/5.5.1", "luajit"} {
		require.NoError(t, os.MkdirAll(filepath.Join(root, dir), 0755))
	}
	current := filepath.Join(root, "current")
	switch kind {
	case "current-directory":
		require.NoError(t, os.Mkdir(current, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(current, "marker"), []byte("keep"), 0600))
	case "current-file":
		require.NoError(t, os.WriteFile(current, []byte("keep"), 0600))
	case "current-dangling":
		require.NoError(t, os.Symlink("missing", current))
	case "current-loop":
		require.NoError(t, os.Symlink("current", current))
	case "current-intermediate-file":
		target = filepath.Join(root, "lua", "entry")
		require.NoError(t, os.WriteFile(target, []byte("keep"), 0600))
		require.NoError(t, os.Symlink("lua/entry/subdir", current))
	case "current-to-file":
		target = filepath.Join(project, "kept-file")
		require.NoError(t, os.WriteFile(target, []byte("keep"), 0600))
		require.NoError(t, os.Symlink(target, current))
	case "rocks-file", "rocks-link", "rocks-dangling":
		lua := filepath.Join(root, "lua/5.5.1")
		require.NoError(t, os.Symlink("lua/5.5.1", current))
		rocks := filepath.Join(lua, "luarocks")
		switch kind {
		case "rocks-file":
			require.NoError(t, os.WriteFile(rocks, []byte("keep"), 0600))
		case "rocks-link":
			require.NoError(t, os.Symlink(target, rocks))
		case "rocks-dangling":
			require.NoError(t, os.Symlink("missing", rocks))
		}
	case "required-file", "required-link", "multiple":
		src := filepath.Join(root, "src")
		require.NoError(t, os.Remove(src))
		if kind == "required-link" {
			require.NoError(t, os.Symlink(target, src))
		} else {
			require.NoError(t, os.WriteFile(src, []byte("keep"), 0600))
		}
		if kind == "multiple" {
			jit := filepath.Join(root, "luajit")
			require.NoError(t, os.Remove(jit))
			require.NoError(t, os.WriteFile(jit, nil, 0600))
		}
	}
	return project, target
}

func TestInteractiveRecoveryCancellation(t *testing.T) {
	for _, answer := range []string{"n", "", "EOF", "not-yes"} {
		t.Run(answer, func(t *testing.T) {
			project, _ := newInteractiveRecoveryProject(t, "current-directory")
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
	project, _ := newInteractiveRecoveryProject(t, "current-file")
	result := recoveryProcess(t, project, []string{"yes"}, "replace", "--recover")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "changed while awaiting confirmation")
	data, err := os.ReadFile(filepath.Join(project, ".lenv/current"))
	require.NoError(t, err)
	require.Equal(t, "replacement", string(data))
	project, _ = newInteractiveRecoveryProject(t, "current-directory")
	result = recoveryProcess(t, project, []string{"yes"}, "contents", "--recover")
	require.Equal(t, 0, result.Exit, result.Output)
}

func TestNoninteractiveRecoveryRejectsPipedApproval(t *testing.T) {
	project, _ := newInteractiveRecoveryProject(t, "current-file")
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
	project, _ := newInteractiveRecoveryProject(t, "current-file")
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
	project, _ := newInteractiveRecoveryProject(t, "current-file")
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
			project, _ := newInteractiveRecoveryProject(t, kind)
			recoveryDiagnosisProject(t, project)
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
