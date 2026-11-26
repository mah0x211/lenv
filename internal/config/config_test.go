package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mah0x211/lenv/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestInitGlobalVarsRejectsUnknownDirectoryMode(t *testing.T) {
	oldCWD := CWD
	t.Cleanup(func() { CWD = oldCWD })
	require.PanicsWithValue(t, "unknown basedir: unknown", func() { InitGlobalVars("unknown") })
}

func TestRecoveryPreservesAutomaticDiscovery(t *testing.T) {
	for _, kind := range []string{"file", "link", "dangling"} {
		t.Run(kind, func(t *testing.T) {
			parent, err := filepath.EvalSymlinks(t.TempDir())
			require.NoError(t, err)
			root := filepath.Join(parent, ".lenv")
			require.NoError(t, os.Mkdir(root, 0755))
			child := filepath.Join(parent, "child")
			require.NoError(t, os.Mkdir(child, 0755))
			candidate := filepath.Join(child, ".lenv")
			if kind == "file" {
				require.NoError(t, os.WriteFile(candidate, []byte("keep"), 0600))
			} else {
				target := root
				if kind == "dangling" {
					target = "missing"
				}
				require.NoError(t, os.Symlink(target, candidate))
			}
			before, err := os.Lstat(candidate)
			require.NoError(t, err)

			testutil.PreserveWorkingDirectory(t)
			oldCWD, oldRoot, oldSrc, oldCurrent, oldUsed := CWD, LenvDir, SrcDir, CurrentDir, CurrentUsed
			oldLua, oldJit, oldRocks := LuaCfg, LuaJitCfg, LuaRocksCfg
			t.Cleanup(func() {
				CWD, LenvDir, SrcDir, CurrentDir, CurrentUsed = oldCWD, oldRoot, oldSrc, oldCurrent, oldUsed
				LuaCfg, LuaJitCfg, LuaRocksCfg = oldLua, oldJit, oldRocks
			})
			LuaCfg, LuaJitCfg, LuaRocksCfg = &TargetConfig{}, &TargetConfig{}, &TargetConfig{}
			require.NoError(t, os.Chdir(child))
			InitGlobalVars("")
			require.Equal(t, root, LenvDir)

			after, err := os.Lstat(candidate)
			require.NoError(t, err)
			require.True(t, os.SameFile(before, after))
			_, err = os.Lstat(filepath.Join(root, "src"))
			require.True(t, os.IsNotExist(err))
		})
	}
}
