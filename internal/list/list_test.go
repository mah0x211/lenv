package list

import (
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

func newListProject(t *testing.T) string {
	t.Helper()
	testutil.PreserveWorkingDirectory(t)
	root := filepath.Join(t.TempDir(), ".lenv")
	for _, name := range []string{"lua/5.4.8", "lua/5.5.1", "luajit/v2.1"} {
		require.NoError(t, os.MkdirAll(filepath.Join(root, name), 0755))
	}
	lua, jit := version.NewVersions(), version.NewVersions()
	for _, item := range []string{"5.4.8", "5.5.1"} {
		lua.Add("lua-"+item+".tar.gz", item, "", ".tar.gz")
		require.NoError(t, os.WriteFile(filepath.Join(root, "lua", item, "marker"), []byte(item), 0600))
	}
	jit.AddBranch("v2.1", "origin")
	for _, item := range []string{"3.12.2", "3.13.0"} {
		dir := filepath.Join(root, "lua/5.4.8/luarocks", item, "lua_modules")
		require.NoError(t, os.MkdirAll(dir, 0755))
	}
	require.NoError(t, lua.WriteFile(filepath.Join(root, "lua_vers.txt")))
	require.NoError(t, jit.WriteFile(filepath.Join(root, "luajit_vers.txt")))
	require.NoError(t, os.Symlink("luarocks/3.12.2/lua_modules", filepath.Join(root, "lua/5.4.8/lua_modules")))

	oldLua, oldJit, oldUsed, oldExit := config.LuaCfg, config.LuaJitCfg, config.CurrentUsed, cli.ExitCode
	t.Cleanup(func() {
		config.LuaCfg, config.LuaJitCfg, config.CurrentUsed, cli.ExitCode = oldLua, oldJit, oldUsed, oldExit
	})
	config.LuaCfg = &config.TargetConfig{Name: "lua", RootDir: filepath.Join(root, "lua"), VersionFile: filepath.Join(root, "lua_vers.txt")}
	config.LuaJitCfg = &config.TargetConfig{Name: "luajit", RootDir: filepath.Join(root, "luajit"), VersionFile: filepath.Join(root, "luajit_vers.txt")}
	config.CurrentUsed = filepath.Join(root, "lua/5.4.8")
	cli.ExitCode = 0
	return root
}

func runList(t *testing.T) (string, string) {
	t.Helper()
	out, stderr, completed := testutil.CaptureCommandOutput(t, CmdList)
	require.True(t, completed)
	require.Equal(t, 0, cli.ExitCode, stderr)
	return out, stderr
}

func TestListCommandInstalledSelections(t *testing.T) {
	root := newListProject(t)
	out, stderr := runList(t)
	require.NotContains(t, stderr, "failed")
	require.Contains(t, out, "5.4.8 ("+filepath.Join(root, "lua/5.4.8")+") (used)")
	require.Contains(t, out, "5.5.1 ("+filepath.Join(root, "lua/5.5.1")+")")
	require.Contains(t, out, "v2.1 ("+filepath.Join(root, "luajit/v2.1")+")")
	require.Contains(t, out, "3.12.2 (used)")
	require.Contains(t, out, "3.13.0")
	require.Less(t, strings.Index(out, "3.13.0"), strings.Index(out, "3.12.2"), out)
	require.Equal(t, 2, strings.Count(out, "(used)"), out)
}

func TestListCommandSkipsUnknownAndNonDirectoryEntries(t *testing.T) {
	root := newListProject(t)
	require.NoError(t, os.Mkdir(filepath.Join(root, "lua/manual"), 0755))
	versions, err := version.NewVersionsFromFile(filepath.Join(root, "lua_vers.txt"))
	require.NoError(t, err)
	versions.Add("lua-5.3.6.tar.gz", "5.3.6", "", ".tar.gz")
	require.NoError(t, versions.WriteFile(filepath.Join(root, "lua_vers.txt")))
	require.NoError(t, os.WriteFile(filepath.Join(root, "lua/5.3.6"), []byte("keep"), 0600))
	out, stderr := runList(t)
	require.Contains(t, stderr, "ignore unknown version: manual")
	require.Contains(t, stderr, "found lua 5.3.6")
	require.Contains(t, stderr, "not a directory")
	require.NotContains(t, out, "manual")
	require.NotContains(t, out, "5.3.6")
	require.Contains(t, out, "v2.1")
	testutil.AssertFileContent(t, filepath.Join(root, "lua/5.3.6"), "keep")
}

func TestListCommandReportsVersionFileFailureAndContinues(t *testing.T) {
	for _, malformed := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "invalid"}[malformed], func(t *testing.T) {
			root := newListProject(t)
			file := filepath.Join(root, "lua_vers.txt")
			if malformed {
				require.NoError(t, os.WriteFile(file, []byte("invalid JSON"), 0600))
			} else {
				require.NoError(t, os.Remove(file))
			}
			out, stderr := runList(t)
			require.Contains(t, stderr, "failed to read version file")
			require.Contains(t, out, "v2.1")
			require.NotContains(t, out, "5.4.8")
		})
	}
}

func TestListCommandReportsInvalidModulesSelection(t *testing.T) {
	for _, tc := range []struct{ name, target, message string }{
		{"regular file", "", "is not symlink"},
		{"absolute link", "absolute", "is not relative symlink"},
		{"outside link", "elsewhere", "is not relative symlink"},
		{"missing directory", "luarocks/missing/lua_modules", "is not symlink to directory"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := newListProject(t)
			modules := filepath.Join(root, "lua/5.4.8/lua_modules")
			require.NoError(t, os.Remove(modules))
			switch tc.target {
			case "":
				require.NoError(t, os.WriteFile(modules, []byte("keep"), 0600))
			case "absolute":
				require.NoError(t, os.Symlink(filepath.Join(root, "lua/5.4.8/luarocks/3.12.2/lua_modules"), modules))
			default:
				require.NoError(t, os.Symlink(tc.target, modules))
			}
			out, stderr := runList(t)
			require.Contains(t, stderr, tc.message)
			require.Contains(t, out, "v2.1")
			_, err := os.Lstat(modules)
			require.NoError(t, err)
		})
	}
}

func TestListCommandReportsInaccessibleRuntime(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := newListProject(t)
	runtime := filepath.Join(root, "lua/5.5.1")
	require.NoError(t, os.Chmod(runtime, 0))
	t.Cleanup(func() { require.NoError(t, os.Chmod(runtime, 0755)) })
	out, stderr := runList(t)
	require.Contains(t, stderr, "permission denied")
	require.Contains(t, out, "v2.1")
}

func TestListCommandReportsUnreadableRocksDirectories(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	for _, selected := range []bool{false, true} {
		t.Run(map[bool]string{false: "read directory", true: "check selection"}[selected], func(t *testing.T) {
			root := newListProject(t)
			rocks := filepath.Join(root, "lua/5.4.8/luarocks")
			if !selected {
				require.NoError(t, os.Remove(filepath.Join(root, "lua/5.4.8/lua_modules")))
			}
			require.NoError(t, os.Chmod(rocks, 0))
			t.Cleanup(func() { require.NoError(t, os.Chmod(rocks, 0755)) })
			out, stderr := runList(t)
			message := "failed to readdir"
			if selected {
				message = "failed to check"
			}
			require.Contains(t, stderr, message)
			require.Contains(t, out, "v2.1")
		})
	}
}

func TestListCommandReportsInaccessibleVersionDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := newListProject(t)
	require.NoError(t, os.Symlink("lua/5.4.8", filepath.Join(root, "current")))
	blocked := filepath.Join(root, "luajit")
	require.NoError(t, os.Chmod(blocked, 0))
	t.Cleanup(func() { require.NoError(t, os.Chmod(blocked, 0755)) })
	_, stderr := runList(t)
	require.Contains(t, stderr, "failed to readdir")
	current, err := os.Readlink(filepath.Join(root, "current"))
	require.NoError(t, err)
	require.Equal(t, "lua/5.4.8", current)
	require.NoError(t, os.Chmod(blocked, 0755))
	require.DirExists(t, filepath.Join(blocked, "v2.1"))
}
