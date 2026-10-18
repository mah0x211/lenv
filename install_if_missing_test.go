package main

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseIfMissingOptions(t *testing.T) {
	for _, args := range [][]string{
		{"--if-missing", "5.5.1", "linux"},
		{"--fetch", "--if-missing", "5.5.1", "linux"},
		{"--if-missing", "--fetch", "5.5.1", "linux"},
	} {
		options, err := parseInstallOptions(args)
		require.NoError(t, err)
		require.True(t, options.ifMissing)
		require.Equal(t, "5.5.1", options.version)
		require.Equal(t, []string{"linux"}, options.build)
		require.Equal(t, len(args) == 4, options.fetch)
	}
	options, err := parseInstallOptions([]string{"5.5.1", "--if-missing", "linux"})
	require.NoError(t, err)
	require.False(t, options.ifMissing)
	require.Equal(t, []string{"--if-missing", "linux"}, options.build)
	_, err = parseInstallOptions([]string{"--if-missing"})
	require.Error(t, err)
}

func addIfMissingVersionLists(t *testing.T) {
	t.Helper()
	lua, err := NewVersionsFromFile(LuaCfg.VersionFile)
	require.NoError(t, err)
	lua.Add("lua-5.5.1.tar.gz", "5.5.1", "sha256:unused", ".tar.gz")
	require.NoError(t, lua.WriteFile(LuaCfg.VersionFile))
	rocks := NewVersions()
	rocks.Add("luarocks-3.13.0.tar.gz", "3.13.0", "pgp:unused", ".tar.gz")
	require.NoError(t, rocks.WriteFile(LuaRocksCfg.VersionFile))
	jit := NewVersions()
	jit.AddBranch("v2.1", "origin")
	require.NoError(t, jit.WriteFile(LuaJitCfg.VersionFile))
}

func installedTestDirectory(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(path, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(path, "marker"), []byte("keep"), 0600))
}

func runInstallForTest(args ...string) bool {
	ExitCode = 0
	done := make(chan struct{})
	completed := false
	go func() { defer close(done); CmdInstall(args); completed = true }()
	<-done
	return completed
}

func TestIfMissingUsesExistingLuaAndLuaJitDirectories(t *testing.T) {
	for _, tc := range []struct{ name, version string }{
		{"lua", "5.5.1"}, {"luajit", "lj-v2.1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newInstallFetchFixture(t, "")
			addIfMissingVersionLists(t)
			dir := filepath.Join(LuaCfg.RootDir, "5.5.1")
			if tc.name == "luajit" {
				dir = filepath.Join(LuaJitCfg.RootDir, "v2.1")
			}
			installedTestDirectory(t, dir)
			// There is deliberately no binary: the contract matches use's directory check.
			require.True(t, runInstallForTest("--if-missing", tc.version, "linux"))
			require.Zero(t, ExitCode)
			resolved, err := filepath.EvalSymlinks(CurrentDir)
			require.NoError(t, err)
			require.Equal(t, dir, resolved)
			data, err := os.ReadFile(filepath.Join(dir, "marker"))
			require.NoError(t, err)
			require.Equal(t, "keep", string(data))
			require.Zero(t, atomic.LoadInt32(&fixture.requests))
			_, err = os.Stat(filepath.Join(SrcDir, "luajit"))
			require.True(t, os.IsNotExist(err))
		})
	}
}

func TestIfMissingSelectsLuaRocksForRequestedLua(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t)
	selectedLua := filepath.Join(LuaCfg.RootDir, "5.5.1")
	selectedRocks := filepath.Join(selectedLua, "luarocks", "3.13.0")
	oldRocks := filepath.Join(CurrentUsed, "luarocks", "3.13.0")
	installedTestDirectory(t, selectedLua)
	installedTestDirectory(t, selectedRocks)
	installedTestDirectory(t, oldRocks)
	require.NoError(t, os.Mkdir(filepath.Join(selectedRocks, "lua_modules"), 0755))
	require.True(t, runInstallForTest("--if-missing", "5.5.1:latest", "linux"))
	require.Equal(t, filepath.Join(selectedLua, "luarocks"), LuaRocksCfg.RootDir)
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
	lua := CurrentUsed
	rocks := filepath.Join(lua, "luarocks", "3.13.0")
	installedTestDirectory(t, rocks)
	require.NoError(t, os.Mkdir(filepath.Join(rocks, "lua_modules"), 0755))
	before, err := os.Readlink(CurrentDir)
	require.NoError(t, err)
	require.True(t, runInstallForTest("--if-missing", ":latest"))
	after, err := os.Readlink(CurrentDir)
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
			target := filepath.Join(LuaCfg.RootDir, "5.5.1")
			switch kind {
			case "file":
				require.NoError(t, os.WriteFile(target, []byte("keep"), 0600))
			case "symlink":
				require.NoError(t, os.Symlink(CurrentUsed, target))
			}
			before, err := os.Readlink(CurrentDir)
			require.NoError(t, err)
			// The local server rejects downloads. A download request proves that install
			// was attempted instead of treating a missing/non-directory target as installed.
			require.False(t, runInstallForTest("--if-missing", "5.5.1", "linux"))
			require.Equal(t, 1, ExitCode)
			require.Equal(t, int32(1), atomic.LoadInt32(&fixture.downloads))
			after, err := os.Readlink(CurrentDir)
			require.NoError(t, err)
			require.Equal(t, before, after)
		})
	}
}

func TestIfMissingInstallsMissingLuaRocksAfterSelectingLua(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t)
	oldRocks := filepath.Join(CurrentUsed, "luarocks", "3.13.0")
	installedTestDirectory(t, oldRocks)
	selected := filepath.Join(LuaCfg.RootDir, "5.5.1")
	installedTestDirectory(t, selected)
	require.False(t, runInstallForTest("--if-missing", "5.5.1:latest", "linux"))
	require.Equal(t, 1, ExitCode)
	require.Equal(t, int32(1), atomic.LoadInt32(&fixture.downloads))
	require.Equal(t, selected, CurrentUsed)
	require.Equal(t, filepath.Join(selected, "luarocks"), LuaRocksCfg.RootDir)
	data, err := os.ReadFile(filepath.Join(oldRocks, "marker"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
}

func TestInstallWithoutIfMissingStillAttemptsReinstallation(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	require.False(t, runInstallForTest("5.4.8", "linux"))
	require.Equal(t, 1, ExitCode)
	require.Equal(t, int32(1), atomic.LoadInt32(&fixture.downloads))
}

func TestIfMissingUsesResolvedPatchVersion(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t)
	installedTestDirectory(t, filepath.Join(LuaCfg.RootDir, "5.5.1"))
	lua, err := NewVersionsFromFile(LuaCfg.VersionFile)
	require.NoError(t, err)
	lua.Add("lua-5.5.2.tar.gz", "5.5.2", "sha256:unused", ".tar.gz")
	require.NoError(t, lua.WriteFile(LuaCfg.VersionFile))
	require.False(t, runInstallForTest("--if-missing", "5.5.", "linux"))
	require.Equal(t, 1, ExitCode)
	require.Equal(t, int32(1), atomic.LoadInt32(&fixture.downloads))
}

func TestFetchAndIfMissingTogether(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	lua := filepath.Join(LuaCfg.RootDir, "5.5.1")
	rocks := filepath.Join(lua, "luarocks", "3.13.0")
	installedTestDirectory(t, lua)
	installedTestDirectory(t, rocks)
	require.NoError(t, os.Mkdir(filepath.Join(rocks, "lua_modules"), 0755))
	require.True(t, runInstallForTest("--fetch", "--if-missing", "5.5.:latest", "linux"))
	require.Zero(t, ExitCode)
	require.Equal(t, int32(3), atomic.LoadInt32(&fixture.requests))
	require.Zero(t, atomic.LoadInt32(&fixture.downloads))
	require.Equal(t, lua, CurrentUsed)
}

func TestFetchFailureDoesNotUseExistingDirectory(t *testing.T) {
	fixture := newInstallFetchFixture(t, "http")
	addIfMissingVersionLists(t)
	installedTestDirectory(t, filepath.Join(LuaCfg.RootDir, "5.5.1"))
	before, err := os.Readlink(CurrentDir)
	require.NoError(t, err)
	require.False(t, runInstallForTest("--fetch", "--if-missing", "5.5.1"))
	require.Equal(t, 1, ExitCode)
	require.Zero(t, atomic.LoadInt32(&fixture.downloads))
	after, err := os.Readlink(CurrentDir)
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func TestProjectOptionWithIfMissing(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	project := filepath.Join(fixture.dir, "project")
	manager := filepath.Join(project, ".lenv")
	selected := filepath.Join(manager, "lua", "5.5.1")
	installedTestDirectory(t, selected)
	require.NoError(t, os.MkdirAll(filepath.Join(manager, "src"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(manager, "luajit"), 0755))
	versions := NewVersions()
	versions.Add("lua-5.5.1.tar.gz", "5.5.1", "sha256:unused", ".tar.gz")
	require.NoError(t, versions.WriteFile(filepath.Join(manager, "lua_vers.txt")))
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"lenv", "--project", "install", "--if-missing", "5.5.1"}
	require.NoError(t, os.Chdir(project))
	start()
	require.Zero(t, ExitCode)
	require.Equal(t, selected, CurrentUsed)
	require.Zero(t, atomic.LoadInt32(&fixture.requests))
}
