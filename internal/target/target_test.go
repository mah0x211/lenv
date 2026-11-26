package target

import (
	"path/filepath"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/testutil"
	version "github.com/mah0x211/lenv/internal/versions"
	"github.com/stretchr/testify/require"
)

func TestLatestSelectionErrors(t *testing.T) {
	for _, tc := range []struct{ kind, query, file string }{
		{"empty", "latest", "lua_vers.txt"},
		{"prerelease", "latest", "lua_vers.txt"},
		{"branch", "lj-latest", "luajit_vers.txt"},
		{"prerelease", ":latest", "luarocks_vers.txt"},
	} {
		t.Run(tc.file+"/"+tc.kind, func(t *testing.T) {
			root := t.TempDir()
			oldLua, oldJit, oldRocks, oldExit := config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg, cli.ExitCode
			t.Cleanup(func() {
				config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg, cli.ExitCode = oldLua, oldJit, oldRocks, oldExit
			})
			config.LuaCfg = &config.TargetConfig{Name: "lua", VersionFile: filepath.Join(root, "lua_vers.txt")}
			config.LuaJitCfg = &config.TargetConfig{Name: "luajit", VersionFile: filepath.Join(root, "luajit_vers.txt")}
			config.LuaRocksCfg = &config.TargetConfig{Name: "luarocks", VersionFile: filepath.Join(root, "luarocks_vers.txt")}
			cli.ExitCode = 0

			versions := version.NewVersions()
			switch tc.kind {
			case "prerelease":
				require.True(t, versions.Add("lua-5.5.0-rc1.tar.gz", "5.5.0-rc1", "", ".tar.gz"))
			case "branch":
				versions.AddBranch("v2.1", "origin")
			}
			require.NoError(t, versions.WriteFile(filepath.Join(root, tc.file)))

			stdout, stderr, completed := testutil.CaptureCommandOutput(t, func() { PickTargetVersion(tc.query, false) })
			require.False(t, completed)
			require.Equal(t, 1, cli.ExitCode, stderr)
			require.Contains(t, stderr, "version \"latest\" does not defined")
			require.Contains(t, stderr, tc.file)
			require.NotContains(t, stdout+stderr, "panic:")
		})
	}
}

func TestLuaJitBranchURLPreservesName(t *testing.T) {
	cfg := &config.TargetConfig{Name: "luajit", DownloadURL: "https://example.invalid/archive/refs/heads/"}
	item := &version.VerItem{Name: "v2.1", Remote: "origin"}
	require.Equal(t, cfg.DownloadURL+"v2.1.tar.gz", SourceURL(cfg, item))
}
