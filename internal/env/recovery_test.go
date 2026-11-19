package env

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mah0x211/lenv/internal/config"
	"github.com/stretchr/testify/require"
)

func TestRecoverySetupUsesMissingSnapshotSafely(t *testing.T) {
	project := t.TempDir()
	recoveryDiagnosisProject(t, project)
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
