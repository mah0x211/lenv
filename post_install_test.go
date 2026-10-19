package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func preserveWorkingDirectory(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.Chdir(wd)) })
}

func newLuaJitPostInstallTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"bin/luajit-2.1", "include/luajit-2.1/lua.h", "lib/libluajit.a", "lib/libluajit.so", "lib/README"} {
		file := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(file), 0755))
		require.NoError(t, os.WriteFile(file, []byte(name), 0600))
	}
	return dir
}

func TestPostInstallLuaJitCreatesLayoutAliases(t *testing.T) {
	preserveWorkingDirectory(t)
	dir := newLuaJitPostInstallTree(t)
	require.NoError(t, postInstallLuaJit(dir))
	for link, target := range map[string]string{
		"bin/lua": "luajit-2.1", "include/lua.h": "luajit-2.1/lua.h", "lib/liblua.a": "libluajit.a", "lib/liblua.so": "libluajit.so",
	} {
		got, err := os.Readlink(filepath.Join(dir, link))
		require.NoError(t, err)
		require.Equal(t, target, got)
		assertArchiveFile(t, filepath.Join(dir, link), filepath.ToSlash(filepath.Join(filepath.Dir(link), target)))
	}
	assertArchiveFile(t, filepath.Join(dir, "lib/README"), "lib/README")
	_, err := os.Lstat(filepath.Join(dir, "lib/liblua"))
	require.True(t, os.IsNotExist(err), "%v", err)
}

func TestPostInstallLuaJitReportsLayoutConflicts(t *testing.T) {
	for _, entry := range []string{"bin", "include", "lib", "bin/lua", "include/lua.h", "lib/liblua.a"} {
		t.Run(entry, func(t *testing.T) {
			preserveWorkingDirectory(t)
			dir := newLuaJitPostInstallTree(t)
			conflict := filepath.Join(dir, entry)
			if filepath.Dir(entry) == "." {
				require.NoError(t, os.RemoveAll(conflict))
				require.NoError(t, os.WriteFile(conflict, []byte("keep"), 0600))
			} else {
				require.NoError(t, os.Mkdir(conflict, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(conflict, "marker"), []byte("keep"), 0600))
			}
			require.Error(t, postInstallLuaJit(dir))
			if filepath.Dir(entry) == "." {
				assertArchiveFile(t, conflict, "keep")
			} else {
				assertArchiveFile(t, filepath.Join(conflict, "marker"), "keep")
			}
		})
	}
}

func TestPostInstallLuaRocksReportsMissingInstallationAndCommand(t *testing.T) {
	preserveWorkingDirectory(t)
	dir := filepath.Join(t.TempDir(), "missing")
	require.Error(t, postInstallLuaRocks(dir))
	require.NoError(t, os.Mkdir(dir, 0755))
	require.Error(t, postInstallLuaRocks(dir))
	_, err := os.Lstat(filepath.Join(dir, "lua_modules"))
	require.True(t, os.IsNotExist(err), "%v", err)
}

func TestPostInstallLuaJitReportsUnreadableLayout(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	for _, name := range []string{"bin", "include/luajit-2.1", "lib"} {
		t.Run(name, func(t *testing.T) {
			preserveWorkingDirectory(t)
			dir := newLuaJitPostInstallTree(t)
			blocked := filepath.Join(dir, name)
			require.NoError(t, os.Chmod(blocked, 0111))
			t.Cleanup(func() { require.NoError(t, os.Chmod(blocked, 0755)) })
			require.Error(t, postInstallLuaJit(dir))
			require.NoError(t, os.Chmod(blocked, 0755))
			require.FileExists(t, filepath.Join(dir, "include/luajit-2.1/lua.h"))
		})
	}
}
