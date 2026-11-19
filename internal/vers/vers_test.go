package vers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/testutil"
	version "github.com/mah0x211/lenv/internal/versions"
	"github.com/stretchr/testify/require"
)

func newVersionsProject(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), ".lenv")
	for _, name := range []string{"lua/5.4.8", "lua/5.5.1", "luajit/v2.1"} {
		require.NoError(t, os.MkdirAll(filepath.Join(root, name), 0755))
	}
	lua, jit, rocks := version.NewVersions(), version.NewVersions(), version.NewVersions()
	for _, item := range []string{"5.4.8", "5.5.1"} {
		lua.Add("lua-"+item+".tar.gz", item, "", ".tar.gz")
		require.NoError(t, os.WriteFile(filepath.Join(root, "lua", item, "marker"), []byte(item), 0600))
	}
	jit.AddBranch("v2.1", "origin")
	for _, item := range []string{"3.12.2", "3.13.0"} {
		rocks.Add("luarocks-"+item+".tar.gz", item, "", ".tar.gz")
	}
	for name, versions := range map[string]*version.Versions{"lua": lua, "luajit": jit, "luarocks": rocks} {
		require.NoError(t, versions.WriteFile(filepath.Join(root, name+"_vers.txt")))
	}

	oldLua, oldJit, oldRocks, oldExit := config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg, cli.ExitCode
	t.Cleanup(func() {
		config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg, cli.ExitCode = oldLua, oldJit, oldRocks, oldExit
	})
	config.LuaCfg = &config.TargetConfig{Name: "lua", RootDir: filepath.Join(root, "lua"), VersionFile: filepath.Join(root, "lua_vers.txt")}
	config.LuaJitCfg = &config.TargetConfig{Name: "luajit", RootDir: filepath.Join(root, "luajit"), VersionFile: filepath.Join(root, "luajit_vers.txt")}
	config.LuaRocksCfg = &config.TargetConfig{Name: "luarocks", VersionFile: filepath.Join(root, "luarocks_vers.txt")}
	cli.ExitCode = 0
	return root
}

func runVersions(t *testing.T) (string, string, bool) {
	t.Helper()
	return testutil.CaptureCommandOutput(t, CmdVers)
}

func TestVersionsCommandListsAvailableRatherThanInstalledVersions(t *testing.T) {
	root := newVersionsProject(t)
	require.NoError(t, os.RemoveAll(filepath.Join(root, "lua/5.5.1")))
	versions, err := version.NewVersionsFromFile(filepath.Join(root, "lua_vers.txt"))
	require.NoError(t, err)
	for i := 0; i < 12; i++ {
		item := fmt.Sprintf("5.3.%d", i+1)
		versions.Add("lua-"+item+".tar.gz", item, "", ".tar.gz")
	}
	require.NoError(t, versions.WriteFile(filepath.Join(root, "lua_vers.txt")))
	out, stderr, completed := runVersions(t)
	require.True(t, completed)
	require.Equal(t, 0, cli.ExitCode, stderr)
	require.Contains(t, out, `list available "lua" versions:`)
	require.Contains(t, out, `list available "luajit" versions:`)
	require.Contains(t, out, `list available "luarocks" versions:`)
	for _, item := range []string{"5.5.1", "5.4.8", "v2.1", "3.13.0", "3.12.2"} {
		require.Contains(t, out, item)
	}
	require.Less(t, strings.Index(out, "5.5.1"), strings.Index(out, "5.4.8"))
	require.NotContains(t, out, "failed")
	require.NotContains(t, stderr, "Error:")
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "5.") {
			require.Less(t, len(line), 80, line)
		}
	}
}

func TestVersionsCommandReportsUnreadableMetadataAndContinues(t *testing.T) {
	for _, state := range []string{"missing", "invalid", "empty"} {
		t.Run(state, func(t *testing.T) {
			root := newVersionsProject(t)
			file := filepath.Join(root, "lua_vers.txt")
			switch state {
			case "missing":
				require.NoError(t, os.Remove(file))
			case "invalid":
				require.NoError(t, os.WriteFile(file, []byte("{"), 0600))
			case "empty":
				require.NoError(t, version.NewVersions().WriteFile(file))
			}
			out, stderr, completed := runVersions(t)
			require.True(t, completed)
			require.Equal(t, 0, cli.ExitCode, stderr)
			if state != "empty" {
				require.Contains(t, out, "failed to read version file")
			} else {
				require.NotContains(t, out, "failed")
			}
			require.NotContains(t, out, "5.4.8")
			require.Contains(t, out, "v2.1")
			require.Contains(t, out, "3.13.0")
		})
	}
}

func TestVersionsCommandReadsMetadataWithoutParsedSemVer(t *testing.T) {
	root := newVersionsProject(t)
	require.NoError(t, os.WriteFile(filepath.Join(root, "lua_vers.txt"), []byte(`{
 "5.3.6":{"Ver":"5.3.6"},"5.4.8":{"Ver":"5.4.8"},"5.5.1":{"Ver":"5.5.1"}
 }`), 0600))
	out, stderr, completed := runVersions(t)
	require.True(t, completed)
	require.Equal(t, 0, cli.ExitCode, stderr)
	require.Less(t, strings.Index(out, "5.5.1"), strings.Index(out, "5.4.8"))
	require.Less(t, strings.Index(out, "5.4.8"), strings.Index(out, "5.3.6"))
	for _, item := range []string{"5.5.1", "5.4.8", "5.3.6"} {
		require.Contains(t, out, item)
	}
}

func TestVersionsCommandReportsVersionNumberOverflow(t *testing.T) {
	root := newVersionsProject(t)
	require.NoError(t, os.WriteFile(filepath.Join(root, "lua_vers.txt"), []byte(`{
 "5.3.6":{"Ver":"5.3.6"},"overflow":{"Ver":"5.99999999999999999999999999999.1"}
 }`), 0600))
	_, stderr, completed := runVersions(t)
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "failed to strconv.Atoi")
	testutil.AssertFileContent(t, filepath.Join(root, "lua/5.4.8/marker"), "5.4.8")
}
