package install

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/testutil"
	version "github.com/mah0x211/lenv/internal/versions"
	"github.com/stretchr/testify/require"
)

func addIfMissingVersionLists(t *testing.T) {
	t.Helper()
	lua, err := version.NewVersionsFromFile(config.LuaCfg.VersionFile)
	require.NoError(t, err)
	lua.Add("lua-5.5.1.tar.gz", "5.5.1", "sha256:unused", ".tar.gz")
	require.NoError(t, lua.WriteFile(config.LuaCfg.VersionFile))
	rocks := version.NewVersions()
	rocks.Add("luarocks-3.13.0.tar.gz", "3.13.0", "pgp:unused", ".tar.gz")
	require.NoError(t, rocks.WriteFile(config.LuaRocksCfg.VersionFile))
	jit := version.NewVersions()
	jit.AddBranch("v2.1", "origin")
	require.NoError(t, jit.WriteFile(config.LuaJitCfg.VersionFile))
}

func installedTestDirectory(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(path, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(path, "marker"), []byte("keep"), 0600))
}

func runInstallForTest(args ...string) bool {
	cli.ExitCode = 0
	done := make(chan struct{})
	completed := false
	go func() { defer close(done); CmdInstall(args); completed = true }()
	<-done
	return completed
}

func newLocalInstallRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	testutil.PreserveWorkingDirectory(t)
	oldRoot, oldSrc := config.LenvDir, config.SrcDir
	oldCurrent, oldUsed := config.CurrentDir, config.CurrentUsed
	oldRocks, oldExit := config.LuaRocksCfg, cli.ExitCode
	t.Cleanup(func() {
		config.LenvDir, config.SrcDir = oldRoot, oldSrc
		config.CurrentDir, config.CurrentUsed = oldCurrent, oldUsed
		config.LuaRocksCfg, cli.ExitCode = oldRocks, oldExit
	})
	config.LenvDir, config.SrcDir = root, filepath.Join(root, "src")
	config.CurrentDir, config.CurrentUsed = filepath.Join(root, "current"), ""
	config.LuaRocksCfg, cli.ExitCode = &config.TargetConfig{Name: "luarocks"}, 0
	return root
}

func TestIfMissingReportsInaccessibleInstallation(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := newLocalInstallRoot(t)
	jit := filepath.Join(root, "luajit")
	require.NoError(t, os.MkdirAll(filepath.Join(jit, "v2.1"), 0755))
	require.NoError(t, os.Chmod(jit, 0))
	t.Cleanup(func() { require.NoError(t, os.Chmod(jit, 0755)) })

	cfg := &config.TargetConfig{Name: "luajit", RootDir: jit}
	item := &version.VerItem{Name: "v2.1", Ver: "v2.1"}
	_, stderr, completed := testutil.CaptureCommandOutput(t, func() {
		doInstall(cfg, item, &options{IfMissing: true})
	})
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "failed to check installation directory")
}

func TestLatestSelectionFailureKeepsInstallation(t *testing.T) {
	root := newLocalInstallRoot(t)
	oldLua := config.LuaCfg
	t.Cleanup(func() { config.LuaCfg = oldLua })
	config.LuaCfg = &config.TargetConfig{Name: "lua", RootDir: filepath.Join(root, "lua"), VersionFile: filepath.Join(root, "lua_vers.txt")}
	runtime := filepath.Join(config.LuaCfg.RootDir, "5.4.8")
	installedTestDirectory(t, filepath.Join(runtime, "luarocks"))
	require.NoError(t, os.WriteFile(filepath.Join(runtime, "marker"), []byte("installed"), 0600))
	require.NoError(t, os.Symlink("lua/5.4.8", config.CurrentDir))
	require.NoError(t, version.NewVersions().WriteFile(config.LuaCfg.VersionFile))

	_, stderr, completed := testutil.CaptureCommandOutput(t, func() { CmdInstall([]string{"latest"}) })
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "version \"latest\" does not defined")
	testutil.AssertFileContent(t, filepath.Join(runtime, "marker"), "installed")
	link, err := os.Readlink(config.CurrentDir)
	require.NoError(t, err)
	require.Equal(t, "lua/5.4.8", link)
	_, err = os.Lstat(filepath.Join(runtime, "lua_modules"))
	require.True(t, os.IsNotExist(err))
}

func TestIfMissingUsesExistingLuaAndLuaJitDirectories(t *testing.T) {
	for _, tc := range []struct{ name, version string }{
		{"lua", "5.5.1"}, {"luajit", "lj-v2.1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newInstallFetchFixture(t, "")
			addIfMissingVersionLists(t)
			dir := filepath.Join(config.LuaCfg.RootDir, "5.5.1")
			if tc.name == "luajit" {
				dir = filepath.Join(config.LuaJitCfg.RootDir, "v2.1")
			}
			installedTestDirectory(t, dir)
			// There is deliberately no binary: the contract matches use's directory check.
			require.True(t, runInstallForTest("--if-missing", tc.version, "linux"))
			require.Zero(t, cli.ExitCode)
			resolved, err := filepath.EvalSymlinks(config.CurrentDir)
			require.NoError(t, err)
			require.Equal(t, dir, resolved)
			data, err := os.ReadFile(filepath.Join(dir, "marker"))
			require.NoError(t, err)
			require.Equal(t, "keep", string(data))
			require.Zero(t, atomic.LoadInt32(&fixture.requests))
			_, err = os.Stat(filepath.Join(config.SrcDir, "luajit"))
			require.True(t, os.IsNotExist(err))
		})
	}
}

func TestIfMissingSelectsLuaRocksForRequestedLua(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t)
	selectedLua := filepath.Join(config.LuaCfg.RootDir, "5.5.1")
	selectedRocks := filepath.Join(selectedLua, "luarocks", "3.13.0")
	oldRocks := filepath.Join(config.CurrentUsed, "luarocks", "3.13.0")
	installedTestDirectory(t, selectedLua)
	installedTestDirectory(t, selectedRocks)
	installedTestDirectory(t, oldRocks)
	require.NoError(t, os.Mkdir(filepath.Join(selectedRocks, "lua_modules"), 0755))
	require.True(t, runInstallForTest("--if-missing", "5.5.1:latest", "linux"))
	require.Equal(t, filepath.Join(selectedLua, "luarocks"), config.LuaRocksCfg.RootDir)
	modules, err := filepath.EvalSymlinks(filepath.Join(selectedLua, "lua_modules"))
	require.NoError(t, err)
	require.Equal(t, filepath.Join(selectedRocks, "lua_modules"), modules)
	data, err := os.ReadFile(filepath.Join(oldRocks, "marker"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
	require.Zero(t, atomic.LoadInt32(&fixture.requests))
}

func TestIfMissingLuaRocksOnlyUsesCurrentLua(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t)
	lua := config.CurrentUsed
	rocks := filepath.Join(lua, "luarocks", "3.13.0")
	installedTestDirectory(t, rocks)
	require.NoError(t, os.Mkdir(filepath.Join(rocks, "lua_modules"), 0755))
	before, err := os.Readlink(config.CurrentDir)
	require.NoError(t, err)
	require.True(t, runInstallForTest("--if-missing", ":latest"))
	after, err := os.Readlink(config.CurrentDir)
	require.NoError(t, err)
	require.Equal(t, before, after)
	modules, err := filepath.EvalSymlinks(filepath.Join(lua, "lua_modules"))
	require.NoError(t, err)
	require.Equal(t, filepath.Join(rocks, "lua_modules"), modules)
	require.Zero(t, atomic.LoadInt32(&fixture.requests))
}

func TestIfMissingFallsBackToInstallation(t *testing.T) {
	for _, kind := range []string{"missing", "file", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			fixture := newInstallFetchFixture(t, "")
			addIfMissingVersionLists(t)
			target := filepath.Join(config.LuaCfg.RootDir, "5.5.1")
			switch kind {
			case "file":
				require.NoError(t, os.WriteFile(target, []byte("keep"), 0600))
			case "symlink":
				require.NoError(t, os.Symlink(config.CurrentUsed, target))
			}
			before, err := os.Readlink(config.CurrentDir)
			require.NoError(t, err)
			// The local server rejects downloads. A download request proves that install
			// was attempted instead of treating a missing/non-directory target as installed.
			require.False(t, runInstallForTest("--if-missing", "5.5.1", "linux"))
			require.Equal(t, 1, cli.ExitCode)
			require.Equal(t, int32(1), atomic.LoadInt32(&fixture.downloads))
			after, err := os.Readlink(config.CurrentDir)
			require.NoError(t, err)
			require.Equal(t, before, after)
		})
	}
}

func TestIfMissingInstallsMissingLuaRocksAfterSelectingLua(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t)
	oldRocks := filepath.Join(config.CurrentUsed, "luarocks", "3.13.0")
	installedTestDirectory(t, oldRocks)
	selected := filepath.Join(config.LuaCfg.RootDir, "5.5.1")
	installedTestDirectory(t, selected)
	require.False(t, runInstallForTest("--if-missing", "5.5.1:latest", "linux"))
	require.Equal(t, 1, cli.ExitCode)
	require.Equal(t, int32(1), atomic.LoadInt32(&fixture.downloads))
	require.Equal(t, selected, config.CurrentUsed)
	require.Equal(t, filepath.Join(selected, "luarocks"), config.LuaRocksCfg.RootDir)
	data, err := os.ReadFile(filepath.Join(oldRocks, "marker"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
}

func TestInstallWithoutIfMissingStillAttemptsReinstallation(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	require.False(t, runInstallForTest("5.4.8", "linux"))
	require.Equal(t, 1, cli.ExitCode)
	require.Equal(t, int32(1), atomic.LoadInt32(&fixture.downloads))
}

func TestIfMissingUsesResolvedPatchVersion(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t)
	installedTestDirectory(t, filepath.Join(config.LuaCfg.RootDir, "5.5.1"))
	lua, err := version.NewVersionsFromFile(config.LuaCfg.VersionFile)
	require.NoError(t, err)
	lua.Add("lua-5.5.2.tar.gz", "5.5.2", "sha256:unused", ".tar.gz")
	require.NoError(t, lua.WriteFile(config.LuaCfg.VersionFile))
	require.False(t, runInstallForTest("--if-missing", "5.5.", "linux"))
	require.Equal(t, 1, cli.ExitCode)
	require.Equal(t, int32(1), atomic.LoadInt32(&fixture.downloads))
}

func TestFetchAndIfMissingTogether(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	lua := filepath.Join(config.LuaCfg.RootDir, "5.5.1")
	rocks := filepath.Join(lua, "luarocks", "3.13.0")
	installedTestDirectory(t, lua)
	installedTestDirectory(t, rocks)
	require.NoError(t, os.Mkdir(filepath.Join(rocks, "lua_modules"), 0755))
	require.True(t, runInstallForTest("--fetch", "--if-missing", "5.5.:latest", "linux"))
	require.Zero(t, cli.ExitCode)
	require.Equal(t, int32(3), atomic.LoadInt32(&fixture.requests))
	require.Zero(t, atomic.LoadInt32(&fixture.downloads))
	require.Equal(t, lua, config.CurrentUsed)
}

func TestFetchFailureDoesNotUseExistingDirectory(t *testing.T) {
	fixture := newInstallFetchFixture(t, "http")
	addIfMissingVersionLists(t)
	installedTestDirectory(t, filepath.Join(config.LuaCfg.RootDir, "5.5.1"))
	before, err := os.Readlink(config.CurrentDir)
	require.NoError(t, err)
	require.False(t, runInstallForTest("--fetch", "--if-missing", "5.5.1"))
	require.Equal(t, 1, cli.ExitCode)
	require.Zero(t, atomic.LoadInt32(&fixture.downloads))
	after, err := os.Readlink(config.CurrentDir)
	require.NoError(t, err)
	require.Equal(t, before, after)
}
