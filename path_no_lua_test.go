package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathWithoutLuaSubprocess(t *testing.T) {
	if os.Getenv("LENV_PATH_WITHOUT_LUA_PROCESS") != "1" {
		return
	}
	os.Args = []string{"lenv", "-p", "path"}
	if target := os.Getenv("LENV_PATH_WITHOUT_LUA_TARGET"); target != "" {
		os.Args = append(os.Args, target)
	}
	main()
}

func TestPathWithoutLua(t *testing.T) {
	project, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	for _, name := range []string{"src", "lua", "luajit"} {
		require.NoError(t, os.MkdirAll(filepath.Join(project, ".lenv", name), 0755))
	}
	current := filepath.Join(project, ".lenv/current")
	expected := map[string]string{
		"bin":     filepath.Join(current, "bin") + ":" + filepath.Join(current, "lua_modules/bin"),
		"lualib":  filepath.Join(current, "lua_modules/lualib/?.lua") + ";" + filepath.Join(current, "lua_modules/lualib/?/init.lua") + ";;",
		"luaclib": filepath.Join(current, "lua_modules/luaclib/?.so") + ";;",
	}
	emptyPath := t.TempDir()
	for _, target := range []string{"", "bin", "lualib", "luaclib"} {
		t.Run(target, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestPathWithoutLuaSubprocess$")
			cmd.Dir = project
			cmd.Env = append(os.Environ(), "PATH="+emptyPath, "LENV_PATH_WITHOUT_LUA_PROCESS=1", "LENV_PATH_WITHOUT_LUA_TARGET="+target,
				"LUA_INIT=error('must not run')", "LUA_PATH=unexpected", "LUA_CPATH=unexpected", "LUA_PATH_5_5=unexpected", "LUA_CPATH_5_5=unexpected")
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			output, err := cmd.Output()
			require.NoError(t, err, "%s", stderr.String())
			if target != "" {
				require.Equal(t, expected[target]+"\n", string(output))
				return
			}
			rc := filepath.Join(t.TempDir(), "lenvrc")
			require.NoError(t, os.WriteFile(rc, output, 0600))
			values := sourcePathSettings(t, "/bin/sh", rc, "/inherited/bin")
			require.Equal(t, []string{expected["bin"] + ":/inherited/bin", expected["lualib"], expected["luaclib"]}, values)
			require.NotContains(t, strings.Join(values, "\n"), "unexpected")
		})
	}
	// Generating settings must not install or select a runtime.
	_, err = os.Lstat(current)
	require.True(t, os.IsNotExist(err))
	for _, name := range []string{"lua", "luajit"} {
		entries, err := os.ReadDir(filepath.Join(project, ".lenv", name))
		require.NoError(t, err)
		require.Empty(t, entries)
	}
}
