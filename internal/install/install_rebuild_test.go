package install

import (
	"archive/tar"
	"fmt"
	"net/http"
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
	installedTestDirectory(t, other)
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
			installedTestDirectory(t, filepath.Join(config.LuaCfg.RootDir, "5.5.1"))
			installedTestDirectory(t, filepath.Join(config.LuaCfg.RootDir, "5.5.1/luarocks/3.13.0/lua_modules"))
			luaSource := filepath.Join(config.LenvDir, "build/lua/5.5.1")
			rocksSource := filepath.Join(config.LenvDir, "build/luarocks/lua/5.5.1/3.13.0")
			installedTestDirectory(t, luaSource)
			installedTestDirectory(t, rocksSource)
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
	defer server.Close()
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
