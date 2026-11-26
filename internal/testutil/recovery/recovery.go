package recovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/stretchr/testify/require"
)

func NewProject(t testing.TB, kind string) (string, string) {
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

func ConfigureProject(t testing.TB, project string) {
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
