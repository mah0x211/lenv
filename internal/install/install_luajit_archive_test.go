package install

import (
	"archive/tar"
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
	"github.com/stretchr/testify/require"
)

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
	defer server.Close()
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
			defer server.Close()
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
