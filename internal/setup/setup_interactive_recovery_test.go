package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/env"
	"github.com/mah0x211/lenv/internal/testutil"
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

func recoveryProject(t *testing.T, kind string) (string, string) {
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

func TestRecoverySetupDiagnosesPreviouslyBlockedSelection(t *testing.T) {
	project, _ := recoveryProject(t, "multiple")
	current := filepath.Join(project, ".lenv/current")
	require.NoError(t, os.WriteFile(current, []byte("old current"), 0600))
	result := recoveryProcess(t, project, []string{"yes", "yes", "yes"}, "--recover")
	require.Equal(t, 0, result.Exit, result.Output)
	require.Contains(t, result.Output, "The required directories are ready")
	_, err := os.Lstat(current)
	require.True(t, os.IsNotExist(err))
}

func TestRecoverySetupStopsWhenDeferredRecoveryDeclined(t *testing.T) {
	project, _ := recoveryProject(t, "required-file")
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
	project, _ := recoveryProject(t, "required-file")
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
			project, kept := recoveryProject(t, kind)
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
	project, _ := recoveryProject(t, "current-file")
	result := recoveryProcess(t, project, []string{"yes"}, "--recover")
	require.Equal(t, 0, result.Exit, result.Output)
	prompt := strings.Index(result.Output, "Delete and recover? [y/N]")
	removing := strings.Index(result.Output, "Removing regular file")
	removed := strings.Index(result.Output, "Removed regular file")
	complete := strings.Index(result.Output, "The required directories are ready.")
	require.True(t, prompt >= 0 && removing > prompt && removed > removing && complete > removed, result.Output)
}
