package install

import (
	"archive/tar"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/env"
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
	testutil.CreateDirectoryWithMarker(t, filepath.Join(runtime, "luarocks"))
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
			testutil.CreateDirectoryWithMarker(t, dir)
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
	testutil.CreateDirectoryWithMarker(t, selectedLua)
	testutil.CreateDirectoryWithMarker(t, selectedRocks)
	testutil.CreateDirectoryWithMarker(t, oldRocks)
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
	testutil.CreateDirectoryWithMarker(t, rocks)
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
	testutil.CreateDirectoryWithMarker(t, oldRocks)
	selected := filepath.Join(config.LuaCfg.RootDir, "5.5.1")
	testutil.CreateDirectoryWithMarker(t, selected)
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
	testutil.CreateDirectoryWithMarker(t, filepath.Join(config.LuaCfg.RootDir, "5.5.1"))
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
	testutil.CreateDirectoryWithMarker(t, lua)
	testutil.CreateDirectoryWithMarker(t, rocks)
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
	testutil.CreateDirectoryWithMarker(t, filepath.Join(config.LuaCfg.RootDir, "5.5.1"))
	before, err := os.Readlink(config.CurrentDir)
	require.NoError(t, err)
	require.False(t, runInstallForTest("--fetch", "--if-missing", "5.5.1"))
	require.Equal(t, 1, cli.ExitCode)
	require.Zero(t, atomic.LoadInt32(&fixture.downloads))
	after, err := os.Readlink(config.CurrentDir)
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func pathWithoutGit(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"make", "sh", "cat", "cp", "mkdir"} {
		binary, err := exec.LookPath(name)
		require.NoError(t, err)
		binary, err = filepath.Abs(binary)
		require.NoError(t, err)
		require.NoError(t, os.Symlink(binary, filepath.Join(dir, name)))
	}
	t.Setenv("PATH", dir)
	_, err := exec.LookPath("git")
	require.Error(t, err, "Git must not be available in this test")
}

func luaJitBuildArchive(t *testing.T, version string) []byte {
	t.Helper()
	// This small project exercises the real make/install and post-install flow.
	// Like LuaJIT's archive build, it reads its release from .relver without Git.
	makefile := "all:\n\ttest ! -d .git\n\tcat .relver > built\n" +
		"install:\n\tmkdir -p \"$(PREFIX)/bin\" \"$(PREFIX)/include/luajit-2.1\" \"$(PREFIX)/lib\"\n" +
		"\tcp built \"$(PREFIX)/bin/luajit-2.1\"\n" +
		"\tprintf header > \"$(PREFIX)/include/luajit-2.1/lua.h\"\n" +
		"\tprintf library > \"$(PREFIX)/lib/libluajit.so\"\n"
	return testutil.ArchiveData(t,
		testutil.ArchiveEntry{Name: "LuaJIT-2.1/", Kind: tar.TypeDir},
		testutil.ArchiveEntry{Name: "LuaJIT-2.1/.relver", Kind: tar.TypeReg, Body: version},
		testutil.ArchiveEntry{Name: "LuaJIT-2.1/Makefile", Kind: tar.TypeReg, Body: makefile})
}

func TestLuaJitArchiveInstallWithoutGit(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t) // Legacy Name/Remote/Ver branch-list format.
	legacy := filepath.Join(config.SrcDir, "luajit", ".git")
	require.NoError(t, os.MkdirAll(legacy, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(legacy, "keep"), []byte("old clone"), 0600))
	pathWithoutGit(t)
	buildTmp := t.TempDir()
	t.Setenv("TMPDIR", buildTmp)
	var downloads int32
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2.1.tar.gz" {
			t.Errorf("unexpected archive request %q", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		version := strconv.Itoa(1000 + int(atomic.AddInt32(&downloads, 1)))
		_, _ = w.Write(luaJitBuildArchive(t, version))
	}))
	config.LuaJitCfg.DownloadURL = server.URL + "/"
	// A mutable branch must not reuse an older cached archive.
	require.NoError(t, os.WriteFile(filepath.Join(config.SrcDir, "v2.1.tar.gz"), luaJitBuildArchive(t, "cached"), 0600))
	for i := 1; i <= 2; i++ {
		args := []string{"lj-v2.1"}
		if i == 2 {
			args = append([]string{"--fetch"}, args...)
		}
		require.True(t, runInstallForTest(args...))
		require.Zero(t, cli.ExitCode)
		require.Equal(t, int32(i), atomic.LoadInt32(&downloads))
		testutil.AssertFileContent(t, filepath.Join(config.LuaJitCfg.RootDir, "v2.1/bin/luajit-2.1"), strconv.Itoa(1000+i))
		testutil.AssertFileContent(t, filepath.Join(legacy, "keep"), "old clone")
		testutil.AssertDirectoryEmpty(t, buildTmp)
		binary, err := filepath.EvalSymlinks(filepath.Join(config.CurrentDir, "bin/lua"))
		require.NoError(t, err)
		expected, err := filepath.EvalSymlinks(filepath.Join(config.LuaJitCfg.RootDir, "v2.1/bin/luajit-2.1"))
		require.NoError(t, err)
		require.Equal(t, expected, binary)
	}
	require.True(t, runInstallForTest("--if-missing", "lj-v2.1"))
	require.Equal(t, int32(2), atomic.LoadInt32(&downloads))
	require.Equal(t, int32(3), atomic.LoadInt32(&fixture.requests))
	testutil.AssertDirectoryEmpty(t, buildTmp)
}

func TestLuaJitArchivePreparationFailureKeepsInstallation(t *testing.T) {
	for _, failure := range []string{"http", "gzip"} {
		t.Run(failure, func(t *testing.T) {
			newInstallFetchFixture(t, "")
			addIfMissingVersionLists(t)
			installed := filepath.Join(config.LuaJitCfg.RootDir, "v2.1")
			require.NoError(t, os.MkdirAll(installed, 0755))
			marker := filepath.Join(installed, "keep")
			require.NoError(t, os.WriteFile(marker, []byte("installed"), 0600))
			require.NoError(t, os.Remove(config.CurrentDir))
			require.NoError(t, os.Symlink(installed, config.CurrentDir))
			require.NoError(t, env.ResolveCurrentDir())
			pathWithoutGit(t)
			buildTmp := t.TempDir()
			t.Setenv("TMPDIR", buildTmp)
			var downloads int32
			server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&downloads, 1)
				if failure == "http" {
					w.WriteHeader(503)
				} else {
					fmt.Fprint(w, "not gzip")
				}
			}))
			config.LuaJitCfg.DownloadURL = server.URL + "/"
			require.NoError(t, os.WriteFile(filepath.Join(config.SrcDir, "v2.1.tar.gz"), luaJitBuildArchive(t, "stale"), 0600))
			require.False(t, runInstallForTest("lj-v2.1"))
			require.Equal(t, 1, cli.ExitCode)
			require.Equal(t, int32(1), atomic.LoadInt32(&downloads))
			testutil.AssertFileContent(t, marker, "installed")
			target, err := filepath.EvalSymlinks(config.CurrentDir)
			require.NoError(t, err)
			require.Equal(t, installed, target)
			testutil.AssertDirectoryEmpty(t, buildTmp)
		})
	}
}

const rebuildMakefile = "all:\n\tcp input built\ninstall:\n\tmkdir -p \"$(INSTALL_TOP)\"\n\tcp built \"$(INSTALL_TOP)/marker\"\nclean:\n\tfalse\n"

func rebuildLuaArchive(t *testing.T, fail bool) []byte {
	t.Helper()
	makefile := rebuildMakefile
	if fail {
		makefile = "all:\n\tfalse\n"
	}
	return testutil.ArchiveData(t, testutil.ArchiveEntry{Name: "source/", Kind: tar.TypeDir},
		testutil.ArchiveEntry{Name: "source/Makefile", Kind: tar.TypeReg, Body: makefile},
		testutil.ArchiveEntry{Name: "source/input", Kind: tar.TypeReg, Body: "upstream"})
}

func cacheRebuildLua(t *testing.T, fail bool) string {
	t.Helper()
	data := rebuildLuaArchive(t, fail)
	versions := version.NewVersions()
	versions.Add("lua-5.5.1.tar.gz", "5.5.1", testutil.ArchiveSHA256(data), ".tar.gz")
	require.NoError(t, versions.WriteFile(config.LuaCfg.VersionFile))
	require.NoError(t, os.WriteFile(filepath.Join(config.SrcDir, "lua-5.5.1.tar.gz"), data, 0600))
	path, err := buildSourcePath(config.LuaCfg, &version.VerItem{Ver: "5.5.1"})
	require.NoError(t, err)
	return path
}

func TestRebuildLuaRetainsEditsAndNoRebuildReplaces(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	path := cacheRebuildLua(t, false)
	require.True(t, runInstallForTest("--rebuild", "5.5.1"))
	testutil.AssertFileContent(t, filepath.Join(path, "built"), "upstream")
	require.NoError(t, os.WriteFile(filepath.Join(path, "input"), []byte("edited"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(path, "debug-output"), []byte("keep"), 0600))
	// Reusing edited sources does not require usable download or verification metadata.
	config.LuaCfg.DownloadURL = "http://example.invalid/"
	versions := version.NewVersions()
	versions.Add("lua-5.5.1.tar.gz", "5.5.1", "sha256:invalid", ".tar.gz")
	require.NoError(t, versions.WriteFile(config.LuaCfg.VersionFile))
	require.True(t, runInstallForTest("--rebuild", "5.5.1"))
	testutil.AssertFileContent(t, filepath.Join(config.LuaCfg.RootDir, "5.5.1/marker"), "edited")
	testutil.AssertFileContent(t, filepath.Join(path, "debug-output"), "keep")
	cacheRebuildLua(t, false)
	other := filepath.Join(config.LenvDir, "build/lua/5.4.8")
	testutil.CreateDirectoryWithMarker(t, other)
	require.True(t, runInstallForTest("--no-rebuild", "5.5.1"))
	testutil.AssertFileContent(t, filepath.Join(config.LuaCfg.RootDir, "5.5.1/marker"), "upstream")
	_, err := os.Lstat(path)
	require.True(t, os.IsNotExist(err))
	testutil.AssertFileContent(t, filepath.Join(other, "marker"), "keep")
	_, err = os.Stat(filepath.Join(config.SrcDir, "lua-5.5.1.tar.gz"))
	require.NoError(t, err)
	require.Zero(t, atomic.LoadInt32(&fixture.downloads))
}

func TestRebuildSourceLifetimeOnBuildFailure(t *testing.T) {
	for _, flag := range []string{"", "--rebuild", "--no-rebuild"} {
		t.Run(flag, func(t *testing.T) {
			newInstallFetchFixture(t, "")
			path := cacheRebuildLua(t, true)
			args := []string{"5.5.1"}
			if flag != "" {
				args = append([]string{flag}, args...)
			}
			require.False(t, runInstallForTest(args...))
			require.Equal(t, 1, cli.ExitCode)
			_, err := os.Stat(filepath.Join(config.LuaCfg.RootDir, "5.5.1"))
			require.True(t, os.IsNotExist(err), "make install must not run after failed build")
			_, err = os.Stat(filepath.Join(path, "input"))
			if flag == "--rebuild" {
				require.NoError(t, err)
			} else {
				require.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestInstallReportsUnremovablePreviousInstallation(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := newLocalInstallRoot(t)
	source := filepath.Join(root, "build/lua/5.5.1")
	require.NoError(t, os.MkdirAll(source, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(source, "marker"), []byte("source"), 0600))
	lua := filepath.Join(root, "lua")
	installed := filepath.Join(lua, "5.5.1")
	require.NoError(t, os.MkdirAll(installed, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(installed, "marker"), []byte("5.5.1"), 0600))
	require.NoError(t, os.Chmod(installed, 0555))
	require.NoError(t, os.Chmod(lua, 0555))
	t.Cleanup(func() {
		require.NoError(t, os.Chmod(lua, 0755))
		require.NoError(t, os.Chmod(installed, 0755))
	})

	cfg := &config.TargetConfig{Name: "lua", RootDir: lua}
	item := &version.VerItem{Name: "lua-5.5.1.tar.gz", Ver: "5.5.1"}
	_, stderr, completed := testutil.CaptureCommandOutput(t, func() {
		doInstall(cfg, item, &options{Rebuild: true})
	})
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "failed to os.RemoveAll")
	testutil.AssertFileContent(t, filepath.Join(installed, "marker"), "5.5.1")
	testutil.AssertFileContent(t, filepath.Join(source, "marker"), "source")
}

func TestRebuildIfMissingLeavesSourcesUntouched(t *testing.T) {
	for _, flag := range []string{"--rebuild", "--no-rebuild"} {
		t.Run(flag, func(t *testing.T) {
			fixture := newInstallFetchFixture(t, "")
			addIfMissingVersionLists(t)
			testutil.CreateDirectoryWithMarker(t, filepath.Join(config.LuaCfg.RootDir, "5.5.1"))
			testutil.CreateDirectoryWithMarker(t, filepath.Join(config.LuaCfg.RootDir, "5.5.1/luarocks/3.13.0/lua_modules"))
			luaSource := filepath.Join(config.LenvDir, "build/lua/5.5.1")
			rocksSource := filepath.Join(config.LenvDir, "build/luarocks/lua/5.5.1/3.13.0")
			testutil.CreateDirectoryWithMarker(t, luaSource)
			testutil.CreateDirectoryWithMarker(t, rocksSource)
			require.True(t, runInstallForTest("--fetch", "--if-missing", flag, "5.5.1:latest"))
			testutil.AssertFileContent(t, filepath.Join(luaSource, "marker"), "keep")
			testutil.AssertFileContent(t, filepath.Join(rocksSource, "marker"), "keep")
			require.Zero(t, atomic.LoadInt32(&fixture.downloads))
			require.Equal(t, int32(3), atomic.LoadInt32(&fixture.requests))
		})
	}
}

func TestRebuildLuaJitRetainsEditsUntilFreshReplacement(t *testing.T) {
	newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t)
	pathWithoutGit(t)
	var downloads int32
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := fmt.Sprint(atomic.AddInt32(&downloads, 1))
		w.Write(luaJitBuildArchive(t, version))
	}))
	config.LuaJitCfg.DownloadURL = server.URL + "/"
	source := filepath.Join(config.LenvDir, "build/luajit/v2.1")
	require.True(t, runInstallForTest("--rebuild", "lj-v2.1"))
	require.NoError(t, os.WriteFile(filepath.Join(source, ".relver"), []byte("edited"), 0600))
	require.True(t, runInstallForTest("--fetch", "--rebuild", "lj-v2.1"))
	require.Equal(t, int32(1), atomic.LoadInt32(&downloads))
	testutil.AssertFileContent(t, filepath.Join(config.LuaJitCfg.RootDir, "v2.1/bin/luajit-2.1"), "edited")
	require.True(t, runInstallForTest("--no-rebuild", "lj-v2.1"))
	require.Equal(t, int32(2), atomic.LoadInt32(&downloads))
	testutil.AssertFileContent(t, filepath.Join(config.LuaJitCfg.RootDir, "v2.1/bin/luajit-2.1"), "2")
	_, err := os.Stat(source)
	require.True(t, os.IsNotExist(err))
}

func TestRebuildLuaRocksUsesEditedSourcesWithoutVerifier(t *testing.T) {
	newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t)
	config.LuaRocksCfg.RootDir = filepath.Join(config.CurrentUsed, "luarocks")
	source, err := buildSourcePath(config.LuaRocksCfg, &version.VerItem{Ver: "3.13.0"})
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(source, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(source, "configure"), []byte("#!/bin/sh\nprintf edited > configure-ran\nexit 1\n"), 0700))
	pathWithoutGit(t) // No gpg, gpg2 or gpgv, but a real shell for configure.
	require.False(t, runInstallForTest("--rebuild", ":3.13.0"))
	testutil.AssertFileContent(t, filepath.Join(source, "configure-ran"), "edited")
	require.Equal(t, 1, cli.ExitCode)
}

func luaRocksSourceTree(t *testing.T) (string, string, string) {
	t.Helper()
	testutil.PreserveWorkingDirectory(t)
	lua, err := exec.LookPath("lua")
	require.NoError(t, err, "a real Lua installation is required, as in CI")
	lua, err = filepath.EvalSymlinks(lua)
	require.NoError(t, err)
	luaRoot := filepath.Dir(filepath.Dir(lua))
	data, _ := testutil.OfficialLuaRocksArchive(t)
	root, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	source, err := untarGz(root, bytes.NewReader(data))
	require.NoError(t, err)
	prefix := filepath.Join(root, "installed")
	return source, prefix, luaRoot
}

func TestInstallLuaRocksFromOfficialSource(t *testing.T) {
	source, prefix, luaRoot := luaRocksSourceTree(t)
	// The upstream build reads Lua/headers but writes only into this test tree.
	require.NoError(t, os.Chdir(source))
	require.NoError(t, installRocks(prefix, &config.TargetConfig{RootDir: filepath.Join(luaRoot, "luarocks")}))
	for _, name := range []string{"bin/luarocks", "bin/luarocks-admin"} {
		require.FileExists(t, filepath.Join(prefix, name))
	}
	for _, name := range []string{"bin", "lualib", "luaclib"} {
		alias := filepath.Join(prefix, "lua_modules", name)
		info, err := os.Lstat(alias)
		require.NoError(t, err)
		require.NotZero(t, info.Mode()&os.ModeSymlink)
		target, err := filepath.EvalSymlinks(alias)
		require.NoError(t, err)
		relative, err := filepath.Rel(prefix, target)
		require.NoError(t, err)
		require.True(t, filepath.IsLocal(relative), "alias %s escaped installation: %s", name, target)
		require.DirExists(t, target)
	}
	cmd := exec.Command(filepath.Join(prefix, "bin/luarocks"), "--version")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	require.Contains(t, string(output), "3.13.0")
	// A literal Lua search path without a question mark is not a module directory.
	t.Setenv("LUA_PATH", prefix+"/literal.lua;"+os.Getenv("LUA_PATH"))
	// Repeated post-install processing must retain valid aliases.
	require.NoError(t, postInstallLuaRocks(prefix))
	_, err = os.Lstat(filepath.Join(prefix, "literal.lua"))
	require.True(t, os.IsNotExist(err), "literal search path was converted into a directory: %v", err)

	// A directory occupying any alias cannot be silently removed.
	for _, name := range []string{"bin", "lualib", "luaclib"} {
		alias := filepath.Join(prefix, "lua_modules", name)
		target, err := os.Readlink(alias)
		require.NoError(t, err)
		require.NoError(t, os.Remove(alias))
		require.NoError(t, os.Mkdir(alias, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(alias, "marker"), []byte("keep"), 0600))
		require.Error(t, postInstallLuaRocks(prefix))
		testutil.AssertFileContent(t, filepath.Join(alias, "marker"), "keep")
		require.NoError(t, os.RemoveAll(alias))
		require.NoError(t, os.Symlink(target, alias))
	}

	// The module-alias parent itself may also be occupied by a file.
	modules := filepath.Join(prefix, "lua_modules")
	require.NoError(t, os.Rename(modules, modules+".saved"))
	require.NoError(t, os.WriteFile(modules, []byte("keep"), 0600))
	require.Error(t, postInstallLuaRocks(prefix))
	testutil.AssertFileContent(t, modules, "keep")
	require.NoError(t, os.Remove(modules))
	require.NoError(t, os.Rename(modules+".saved", modules))
	// A file occupying the C-module directory must also remain intact.
	cmodules, err := filepath.EvalSymlinks(filepath.Join(prefix, "lua_modules/luaclib"))
	require.NoError(t, err)
	require.NoError(t, os.Remove(cmodules))
	require.NoError(t, os.WriteFile(cmodules, []byte("keep"), 0600))
	require.Error(t, postInstallLuaRocks(prefix))
	testutil.AssertFileContent(t, cmodules, "keep")
}

func TestLuaRocksBuildFailuresDoNotPublishInstallation(t *testing.T) {
	for _, stage := range []string{"configure", "build", "install"} {
		t.Run(stage, func(t *testing.T) {
			source, prefix, luaRoot := luaRocksSourceTree(t)
			require.NoError(t, os.Chdir(source))
			if stage == "configure" {
				luaRoot = filepath.Join(t.TempDir(), "missing-lua")
			} else {
				makefile := "all: missing-build-input\n"
				if stage == "install" {
					makefile = "all: built\nbuilt:\n\tprintf built > built\n"
				}
				require.NoError(t, os.WriteFile(filepath.Join(source, "GNUmakefile"), []byte(makefile), 0600))
			}
			require.Error(t, installRocks(prefix, &config.TargetConfig{RootDir: filepath.Join(luaRoot, "luarocks")}))
			_, err := os.Lstat(prefix)
			require.True(t, os.IsNotExist(err), "failed %s created an installation: %v", stage, err)
			require.FileExists(t, filepath.Join(source, "configure"))
			if stage != "configure" {
				require.FileExists(t, filepath.Join(source, "config.unix"), "build/install failures must happen after successful configuration")
			}
			if stage == "install" {
				testutil.AssertFileContent(t, filepath.Join(source, "built"), "built")
			}
		})
	}
}
