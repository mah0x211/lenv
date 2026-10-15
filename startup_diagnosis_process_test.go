package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStartupDiagnosisSubprocess(t *testing.T) {
	if os.Getenv("LENV_DIAG_PROCESS") != "1" {
		return
	}
	var args []string
	require.NoError(t, json.Unmarshal([]byte(os.Getenv("LENV_DIAG_ARGS")), &args))
	os.Args = append([]string{"lenv"}, args...)
	main()
}

func startupDiagnosisProcess(t *testing.T, project string, args ...string) (int, string) {
	t.Helper()
	data, err := json.Marshal(args)
	require.NoError(t, err)
	cmd := exec.Command(os.Args[0], "-test.run=^TestStartupDiagnosisSubprocess$")
	cmd.Dir = project
	cmd.Env = append(os.Environ(), "LENV_DIAG_PROCESS=1", "LENV_DIAG_ARGS="+string(data))
	output, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		var exit *exec.ExitError
		require.ErrorAs(t, err, &exit, "%s", output)
		code = exit.ExitCode()
	}
	return code, string(output)
}

func TestStartupDiagnosisHelpAvailableForInvalidCurrent(t *testing.T) {
	for _, kind := range []string{"file", "directory", "dangling", "loop"} {
		t.Run(kind, func(t *testing.T) {
			project := t.TempDir()
			root := filepath.Join(project, ".lenv")
			for _, name := range []string{"src", "lua", "luajit"} {
				require.NoError(t, os.MkdirAll(filepath.Join(root, name), 0755))
			}
			current := filepath.Join(root, "current")
			switch kind {
			case "file":
				require.NoError(t, os.WriteFile(current, []byte("keep"), 0600))
			case "directory":
				require.NoError(t, os.Mkdir(current, 0755))
			case "dangling":
				require.NoError(t, os.Symlink("missing", current))
			case "loop":
				require.NoError(t, os.Symlink("current", current))
			}
			before, err := os.Lstat(current)
			require.NoError(t, err)
			code, output := startupDiagnosisProcess(t, project, "-p", "help")
			require.Equal(t, 0, code, output)
			require.Contains(t, output, "Usage:")
			require.Contains(t, output, "Cannot initialize lenv")
			require.NotContains(t, output, "Delete and recover?")
			after, err := os.Lstat(current)
			require.NoError(t, err)
			require.True(t, os.SameFile(before, after))
		})
	}
}

func TestStartupDiagnosisSetupReachableForInvalidCurrent(t *testing.T) {
	project := t.TempDir()
	root := filepath.Join(project, ".lenv")
	require.NoError(t, os.Mkdir(root, 0755))
	current := filepath.Join(root, "current")
	require.NoError(t, os.WriteFile(current, []byte("keep"), 0600))
	// Setup must reach directory creation despite the invalid selection. It may
	// stop later if repairing the selection requires explicit confirmation.
	_, output := startupDiagnosisProcess(t, project, "-p", "setup")
	for _, name := range []string{"src", "lua", "luajit"} {
		info, err := os.Lstat(filepath.Join(root, name))
		require.NoError(t, err, "%s", output)
		require.True(t, info.IsDir())
	}
	data, err := os.ReadFile(current)
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
	require.NotContains(t, output, "Delete and recover?")
}

func TestStartupDiagnosisBlocksCommandsWithoutInitializing(t *testing.T) {
	project := t.TempDir()
	code, output := startupDiagnosisProcess(t, project, "-p", "current")
	require.Equal(t, 1, code, output)
	require.Contains(t, output, "Run lenv -p setup")
	_, err := os.Lstat(filepath.Join(project, ".lenv"))
	require.True(t, os.IsNotExist(err))
	code, output = startupDiagnosisProcess(t, project, "-p", "help")
	require.Equal(t, 0, code, output)
	require.Contains(t, output, "Usage:")
	_, err = os.Lstat(filepath.Join(project, ".lenv"))
	require.True(t, os.IsNotExist(err))
}

func TestStartupOutputForAllCommands(t *testing.T) {
	for _, command := range []string{"help", "current", "setup"} {
		for _, healthy := range []bool{false, true} {
			name := command + "/missing"
			if healthy {
				name = command + "/healthy"
			}
			t.Run(name, func(t *testing.T) {
				project := t.TempDir()
				if healthy {
					for _, dir := range []string{"src", "lua", "luajit"} {
						require.NoError(t, os.MkdirAll(filepath.Join(project, ".lenv", dir), 0755))
					}
				}
				code, output := startupDiagnosisProcess(t, project, "-p", command)
				message := "Checking required directories and current selection..."
				require.Equal(t, 1, strings.Count(output, message), output)
				if command == "current" && !healthy {
					require.Equal(t, 1, code, output)
					require.Less(t, strings.Index(output, message), strings.Index(output, "Error:"), output)
				} else {
					require.Equal(t, 0, code, output)
				}
				if command == "help" {
					require.Less(t, strings.Index(output, message), strings.Index(output, "Usage:"), output)
				}
			})
		}
	}
}
