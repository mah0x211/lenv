package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func luaRocksSourceTree(t *testing.T) (string, string, string) {
	t.Helper()
	preserveWorkingDirectory(t)
	lua, err := exec.LookPath("lua")
	require.NoError(t, err, "a real Lua installation is required, as in CI")
	lua, err = filepath.EvalSymlinks(lua)
	require.NoError(t, err)
	luaRoot := filepath.Dir(filepath.Dir(lua))
	data, _ := officialLuaRocksArchive(t)
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
	require.NoError(t, installRocks(prefix, &TargetConfig{RootDir: filepath.Join(luaRoot, "luarocks")}))
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
		assertArchiveFile(t, filepath.Join(alias, "marker"), "keep")
		require.NoError(t, os.RemoveAll(alias))
		require.NoError(t, os.Symlink(target, alias))
	}

	// The module-alias parent itself may also be occupied by a file.
	modules := filepath.Join(prefix, "lua_modules")
	require.NoError(t, os.Rename(modules, modules+".saved"))
	require.NoError(t, os.WriteFile(modules, []byte("keep"), 0600))
	require.Error(t, postInstallLuaRocks(prefix))
	assertArchiveFile(t, modules, "keep")
	require.NoError(t, os.Remove(modules))
	require.NoError(t, os.Rename(modules+".saved", modules))
	// A file occupying the C-module directory must also remain intact.
	cmodules, err := filepath.EvalSymlinks(filepath.Join(prefix, "lua_modules/luaclib"))
	require.NoError(t, err)
	require.NoError(t, os.Remove(cmodules))
	require.NoError(t, os.WriteFile(cmodules, []byte("keep"), 0600))
	require.Error(t, postInstallLuaRocks(prefix))
	assertArchiveFile(t, cmodules, "keep")
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
			require.Error(t, installRocks(prefix, &TargetConfig{RootDir: filepath.Join(luaRoot, "luarocks")}))
			_, err := os.Lstat(prefix)
			require.True(t, os.IsNotExist(err), "failed %s created an installation: %v", stage, err)
			require.FileExists(t, filepath.Join(source, "configure"))
			if stage != "configure" {
				require.FileExists(t, filepath.Join(source, "config.unix"), "build/install failures must happen after successful configuration")
			}
			if stage == "install" {
				assertArchiveFile(t, filepath.Join(source, "built"), "built")
			}
		})
	}
}
