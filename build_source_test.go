package main

import (
	"archive/tar"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRebuildOptions(t *testing.T) {
	for _, flag := range []string{"--rebuild", "--no-rebuild"} {
		options, err := parseInstallOptions([]string{"--fetch", flag, "--if-missing", "--allow-missing-gpg", "5.5.1", "linux"})
		require.NoError(t, err)
		require.Equal(t, flag == "--rebuild", options.rebuild)
		require.Equal(t, flag == "--no-rebuild", options.noRebuild)
		require.True(t, options.fetch && options.ifMissing && options.allowMissingGPG)
		require.Equal(t, []string{"linux"}, options.build)
		options, err = parseInstallOptions([]string{"5.5.1", flag})
		require.NoError(t, err)
		require.False(t, options.rebuild || options.noRebuild)
		require.Equal(t, []string{flag}, options.build)
		_, err = parseInstallOptions([]string{flag})
		require.Error(t, err)
	}
	for _, flags := range [][]string{{"--rebuild", "--no-rebuild"}, {"--no-rebuild", "--rebuild"}} {
		fixture := newInstallFetchFixture(t, "")
		args := append([]string{"--fetch"}, flags...)
		completed, message := captureInstallError(t, append(args, "5.5.1")...)
		require.False(t, completed)
		require.Contains(t, message, "cannot be combined")
		require.Zero(t, atomic.LoadInt32(&fixture.requests))
	}
}

func TestBuildSourcePaths(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	for _, target := range []string{"lua", "luajit"} {
		path, err := buildSourcePath(&TargetConfig{Name: target}, &VerItem{Ver: "5.5.1"})
		require.NoError(t, err)
		require.Equal(t, filepath.Join(fixture.dir, "build", target, "5.5.1"), path)
	}
	for _, runtime := range []string{"lua/5.5.1", "luajit/v2.1"} {
		dir := filepath.Join(fixture.dir, filepath.FromSlash(runtime))
		require.NoError(t, os.MkdirAll(dir, 0755))
		cfg := &TargetConfig{Name: "luarocks", RootDir: filepath.Join(dir, "luarocks")}
		path, err := buildSourcePath(cfg, &VerItem{Ver: "3.13.0"})
		require.NoError(t, err)
		require.Equal(t, filepath.Join(fixture.dir, "build", "luarocks", filepath.FromSlash(runtime), "3.13.0"), path)
	}
	external := t.TempDir()
	cfg := &TargetConfig{Name: "luarocks", RootDir: filepath.Join(external, "luarocks")}
	first, err := buildSourcePath(cfg, &VerItem{Ver: "3.13.0"})
	require.NoError(t, err)
	require.Contains(t, first, filepath.Join("build", "luarocks", "external"))
	cfg.RootDir = filepath.Join(t.TempDir(), "luarocks")
	second, err := buildSourcePath(cfg, &VerItem{Ver: "3.13.0"})
	require.NoError(t, err)
	require.NotEqual(t, first, second)
	for _, version := range []string{"", ".", "../outside", "/outside", "a/b", "a\\b"} {
		_, err := buildSourcePath(LuaCfg, &VerItem{Ver: version})
		require.Error(t, err)
	}
	_, err = buildSourcePath(&TargetConfig{Name: "unknown"}, &VerItem{Ver: "1.0"})
	require.Error(t, err)
}

func TestBuildSourceRejectsConflicts(t *testing.T) {
	for _, conflict := range []string{"parent link", "parent file", "source link", "source file"} {
		t.Run(conflict, func(t *testing.T) {
			newInstallFetchFixture(t, "")
			item := &VerItem{Ver: "5.5.1"}
			path, err := buildSourcePath(LuaCfg, item)
			require.NoError(t, err)
			outside := t.TempDir()
			marker := filepath.Join(outside, "marker")
			require.NoError(t, os.WriteFile(marker, []byte("keep"), 0600))
			at := filepath.Join(LenvDir, "build")
			if strings.HasPrefix(conflict, "source") {
				at = path
				require.NoError(t, os.MkdirAll(filepath.Dir(at), 0755))
			}
			if strings.HasSuffix(conflict, "link") {
				require.NoError(t, os.Symlink(outside, at))
			} else {
				require.NoError(t, os.WriteFile(at, []byte("keep"), 0600))
			}
			for _, opts := range []*installOptions{{rebuild: true}, {noRebuild: true}} {
				_, err := prepareBuildSource(LuaCfg, item, opts)
				require.Error(t, err)
			}
			assertArchiveFile(t, marker, "keep")
			info, err := os.Lstat(at)
			require.NoError(t, err)
			if strings.HasSuffix(conflict, "link") {
				require.NotZero(t, info.Mode()&os.ModeSymlink)
			} else {
				assertArchiveFile(t, at, "keep")
			}
		})
	}
}

func TestBuildSourceRocksReuseWithoutVerifier(t *testing.T) {
	newInstallFetchFixture(t, "")
	cfg := &TargetConfig{Name: "luarocks", RootDir: filepath.Join(CurrentUsed, "luarocks")}
	item := &VerItem{Ver: "3.13.0"}
	path, err := buildSourcePath(cfg, item)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(path, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(path, "edited"), []byte("debug"), 0600))
	t.Setenv("PATH", t.TempDir())
	got, err := prepareBuildSource(cfg, item, &installOptions{rebuild: true})
	require.NoError(t, err)
	require.Equal(t, path, got)
	assertArchiveFile(t, filepath.Join(path, "edited"), "debug")
}

func TestBuildSourceRejectsUnsafeExtraction(t *testing.T) {
	newInstallFetchFixture(t, "")
	data := archiveData(t, archiveEntry{name: "../outside", kind: tar.TypeReg, body: "rejected"})
	item := &VerItem{Name: "lua.tar.gz", Ver: "5.5.1", Sum: archiveSHA256(data)}
	require.NoError(t, os.WriteFile(filepath.Join(SrcDir, item.Name), data, 0600))
	LuaCfg.DownloadURL = "http://example.invalid/"
	path, err := buildSourcePath(LuaCfg, item)
	require.NoError(t, err)
	_, err = prepareBuildSource(LuaCfg, item, &installOptions{rebuild: true})
	require.Error(t, err)
	_, err = os.Lstat(path)
	require.True(t, os.IsNotExist(err))
	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestVerifiedBuildSourceLuaRocks(t *testing.T) {
	newInstallFetchFixture(t, "")
	data, signature := officialLuaRocksArchive(t)
	cfg := &TargetConfig{Name: "luarocks", RootDir: filepath.Join(CurrentUsed, "luarocks"), SigningKeyURL: publisherKeyURL(t)}
	item := &VerItem{Name: "luarocks-3.13.0.tar.gz", Ver: "3.13.0", Sum: "pgp:luarocks-3.13.0.tar.gz.asc"}
	server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".asc") {
			w.Write(signature)
		} else {
			w.Write(data)
		}
	}))
	cfg.DownloadURL = server.URL + "/"
	path, err := prepareBuildSource(cfg, item, &installOptions{rebuild: true})
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(path, "configure"))
	require.NoError(t, err)
	server.Close()
	// Reuse does not fetch the signing key or archive after sources are edited.
	cfg.SigningKeyURL = "http://example.invalid/"
	got, err := prepareBuildSource(cfg, item, &installOptions{rebuild: true})
	require.NoError(t, err)
	require.Equal(t, path, got)
}

func TestBuildSourceRejectsInvalidEnvironmentPaths(t *testing.T) {
	newInstallFetchFixture(t, "")
	_, err := prepareBuildSource(LuaCfg, &VerItem{Ver: "../outside"}, &installOptions{rebuild: true})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid build source version")
	err = ensureBuildParent(filepath.Dir(LenvDir))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid build directory")
	cfg := &TargetConfig{Name: "luarocks", RootDir: filepath.Join(LenvDir, "lua/missing/luarocks")}
	_, err = prepareBuildSource(cfg, &VerItem{Ver: "3.13.0"}, &installOptions{rebuild: true})
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot resolve Lua environment")
}

func TestBuildSourcePublicationKeepsNewlyAppearedEntry(t *testing.T) {
	newInstallFetchFixture(t, "")
	data := rebuildLuaArchive(t, false)
	item := &VerItem{Name: "lua-5.5.1.tar.gz", Ver: "5.5.1", Sum: archiveSHA256(data)}
	path, err := buildSourcePath(LuaCfg, item)
	require.NoError(t, err)
	server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Another operation creates the destination while this acquisition is running.
		if err := os.Mkdir(path, 0755); err != nil {
			t.Error(err)
			w.WriteHeader(500)
			return
		}
		if err := os.WriteFile(filepath.Join(path, "marker"), []byte("new owner"), 0600); err != nil {
			t.Error(err)
			w.WriteHeader(500)
			return
		}
		w.Write(data)
	}))
	defer server.Close()
	LuaCfg.DownloadURL = server.URL + "/"
	_, err = prepareBuildSource(LuaCfg, item, &installOptions{rebuild: true})
	require.Error(t, err)
	require.Contains(t, err.Error(), "appeared during extraction")
	assertArchiveFile(t, filepath.Join(path, "marker"), "new owner")
	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)
	require.Len(t, entries, 1)
}
