package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandsRequireVersion(t *testing.T) {
	for _, command := range []string{"use", "uninstall"} {
		for _, selected := range []bool{false, true} {
			state := "unselected"
			if selected {
				state = "selected"
			}
			for _, empty := range []bool{false, true} {
				argument := "missing"
				if empty {
					argument = "empty"
				}
				t.Run(command+"/"+state+"/"+argument, func(t *testing.T) {
					project := t.TempDir()
					root := filepath.Join(project, ".lenv")
					version := filepath.Join(root, "lua", "5.5.1")
					for _, path := range []string{filepath.Join(root, "src"), filepath.Join(root, "luajit"), filepath.Join(version, "luarocks")} {
						require.NoError(t, os.MkdirAll(path, 0755))
					}
					marker := filepath.Join(version, "keep")
					require.NoError(t, os.WriteFile(marker, []byte("installed environment"), 0600))
					current := filepath.Join(root, "current")
					if selected {
						require.NoError(t, os.Symlink("lua/5.5.1", current))
					}
					before, err := os.ReadDir(root)
					require.NoError(t, err)
					args := []string{"-p", command}
					if empty {
						args = append(args, "")
					}
					code, output := startupDiagnosisProcess(t, project, args...)
					require.Equal(t, 1, code, output)
					require.Contains(t, output, "no version specified")
					require.Contains(t, output, "Usage:")
					require.NotContains(t, output, "panic:")
					data, err := os.ReadFile(marker)
					require.NoError(t, err)
					require.Equal(t, "installed environment", string(data))
					after, err := os.ReadDir(root)
					require.NoError(t, err)
					beforeNames, afterNames := []string{}, []string{}
					for _, entry := range before {
						beforeNames = append(beforeNames, entry.Name())
					}
					for _, entry := range after {
						afterNames = append(afterNames, entry.Name())
					}
					require.Equal(t, beforeNames, afterNames)
					if selected {
						target, err := os.Readlink(current)
						require.NoError(t, err)
						require.Equal(t, "lua/5.5.1", target)
					} else {
						_, err := os.Lstat(current)
						require.True(t, os.IsNotExist(err))
					}
				})
			}
		}
	}
}
