package main

import (
	"archive/tar"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

const rebuildMakefile = "all:\n\tcp input built\ninstall:\n\tmkdir -p \"$(INSTALL_TOP)\"\n\tcp built \"$(INSTALL_TOP)/marker\"\nclean:\n\tfalse\n"

func rebuildLuaArchive(t *testing.T, fail bool) []byte {
	t.Helper()
	makefile := rebuildMakefile
	if fail {
		makefile = "all:\n\tfalse\n"
	}
	return archiveData(t, archiveEntry{name: "source/", kind: tar.TypeDir},
		archiveEntry{name: "source/Makefile", kind: tar.TypeReg, body: makefile},
		archiveEntry{name: "source/input", kind: tar.TypeReg, body: "upstream"})
}

func cacheRebuildLua(t *testing.T, fail bool) string {
	t.Helper()
	data := rebuildLuaArchive(t, fail)
	versions := NewVersions()
	versions.Add("lua-5.5.1.tar.gz", "5.5.1", archiveSHA256(data), ".tar.gz")
	require.NoError(t, versions.WriteFile(LuaCfg.VersionFile))
	require.NoError(t, os.WriteFile(filepath.Join(SrcDir, "lua-5.5.1.tar.gz"), data, 0600))
	path, err := buildSourcePath(LuaCfg, &VerItem{Ver: "5.5.1"})
	require.NoError(t, err)
	return path
}

func TestRebuildLuaRetainsEditsAndNoRebuildReplaces(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	path := cacheRebuildLua(t, false)
	require.True(t, runInstallForTest("--rebuild", "5.5.1"))
	assertArchiveFile(t, filepath.Join(path, "built"), "upstream")
	require.NoError(t, os.WriteFile(filepath.Join(path, "input"), []byte("edited"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(path, "debug-output"), []byte("keep"), 0600))
	// Reusing edited sources does not require usable download or verification metadata.
	LuaCfg.DownloadURL = "http://example.invalid/"
	versions := NewVersions()
	versions.Add("lua-5.5.1.tar.gz", "5.5.1", "sha256:invalid", ".tar.gz")
	require.NoError(t, versions.WriteFile(LuaCfg.VersionFile))
	require.True(t, runInstallForTest("--rebuild", "5.5.1"))
	assertArchiveFile(t, filepath.Join(LuaCfg.RootDir, "5.5.1/marker"), "edited")
	assertArchiveFile(t, filepath.Join(path, "debug-output"), "keep")
	cacheRebuildLua(t, false)
	other := filepath.Join(LenvDir, "build/lua/5.4.8")
	installedTestDirectory(t, other)
	require.True(t, runInstallForTest("--no-rebuild", "5.5.1"))
	assertArchiveFile(t, filepath.Join(LuaCfg.RootDir, "5.5.1/marker"), "upstream")
	_, err := os.Lstat(path)
	require.True(t, os.IsNotExist(err))
	assertArchiveFile(t, filepath.Join(other, "marker"), "keep")
	_, err = os.Stat(filepath.Join(SrcDir, "lua-5.5.1.tar.gz"))
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
			require.Equal(t, 1, ExitCode)
			_, err := os.Stat(filepath.Join(LuaCfg.RootDir, "5.5.1"))
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

func TestRebuildIfMissingLeavesSourcesUntouched(t *testing.T) {
	for _, flag := range []string{"--rebuild", "--no-rebuild"} {
		t.Run(flag, func(t *testing.T) {
			fixture := newInstallFetchFixture(t, "")
			addIfMissingVersionLists(t)
			installedTestDirectory(t, filepath.Join(LuaCfg.RootDir, "5.5.1"))
			installedTestDirectory(t, filepath.Join(LuaCfg.RootDir, "5.5.1/luarocks/3.13.0/lua_modules"))
			luaSource := filepath.Join(LenvDir, "build/lua/5.5.1")
			rocksSource := filepath.Join(LenvDir, "build/luarocks/lua/5.5.1/3.13.0")
			installedTestDirectory(t, luaSource)
			installedTestDirectory(t, rocksSource)
			require.True(t, runInstallForTest("--fetch", "--if-missing", flag, "5.5.1:latest"))
			assertArchiveFile(t, filepath.Join(luaSource, "marker"), "keep")
			assertArchiveFile(t, filepath.Join(rocksSource, "marker"), "keep")
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
	server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := fmt.Sprint(atomic.AddInt32(&downloads, 1))
		w.Write(luaJitBuildArchive(t, version))
	}))
	defer server.Close()
	LuaJitCfg.DownloadURL = server.URL + "/"
	source := filepath.Join(LenvDir, "build/luajit/v2.1")
	require.True(t, runInstallForTest("--rebuild", "lj-v2.1"))
	require.NoError(t, os.WriteFile(filepath.Join(source, ".relver"), []byte("edited"), 0600))
	require.True(t, runInstallForTest("--fetch", "--rebuild", "lj-v2.1"))
	require.Equal(t, int32(1), atomic.LoadInt32(&downloads))
	assertArchiveFile(t, filepath.Join(LuaJitCfg.RootDir, "v2.1/bin/luajit-2.1"), "edited")
	require.True(t, runInstallForTest("--no-rebuild", "lj-v2.1"))
	require.Equal(t, int32(2), atomic.LoadInt32(&downloads))
	assertArchiveFile(t, filepath.Join(LuaJitCfg.RootDir, "v2.1/bin/luajit-2.1"), "2")
	_, err := os.Stat(source)
	require.True(t, os.IsNotExist(err))
}

func TestRebuildLuaRocksUsesEditedSourcesWithoutVerifier(t *testing.T) {
	newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t)
	LuaRocksCfg.RootDir = filepath.Join(CurrentUsed, "luarocks")
	source, err := buildSourcePath(LuaRocksCfg, &VerItem{Ver: "3.13.0"})
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(source, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(source, "configure"), []byte("#!/bin/sh\nprintf edited > configure-ran\nexit 1\n"), 0700))
	pathWithoutGit(t) // No gpg, gpg2 or gpgv, but a real shell for configure.
	require.False(t, runInstallForTest("--rebuild", ":3.13.0"))
	assertArchiveFile(t, filepath.Join(source, "configure-ran"), "edited")
	require.Equal(t, 1, ExitCode)
}
