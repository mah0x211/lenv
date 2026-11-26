package path

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/testutil"
	"github.com/stretchr/testify/require"
)

func pathSettingsOutput(t *testing.T, args ...string) string {
	t.Helper()
	stdout := testutil.NewFakeStdout(t)
	defer stdout.CloseAll()
	CmdPath(config.CurrentDir, args)
	data, err := os.ReadFile(stdout.Close().Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func assertLuaSettings(t *testing.T, got, prefix string, suffixes []string) {
	t.Helper()
	unique := make(map[string]bool)
	want := []string{"", ""}
	for _, suffix := range suffixes {
		if !unique[suffix] {
			want = append(want, "\x00/"+suffix)
			unique[suffix] = true
		}
	}
	actual := strings.Split(strings.ReplaceAll(got, prefix+"/", "\x00/"), ";")
	sort.Strings(want)
	sort.Strings(actual)
	if strings.Join(actual, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("Lua path = %q; want entries %q", got, want)
	}
}

func TestPathShellSettingsPreserveLiteralValues(t *testing.T) {
	old := config.CurrentDir
	t.Cleanup(func() { config.CurrentDir = old })
	patterns := map[string][]string{
		"LUA_PATH":  {"?.lua", "?/init.lua"},
		"LUA_CPATH": {"?.so"},
	}
	base := t.TempDir()
	for _, tc := range []struct{ name, path string }{
		{"ordinary", "project"},
		{"space", "project with spaces"},
		{"single quote", "project's"},
		{"double quote", `project"quoted"`},
		{"variable", "project-$HOME"},
		{"command substitution", "project-$(printf changed)"},
		{"backticks", "project-`printf changed`"},
		{"backslash", `project\name`},
		{"newline", "project\nname"},
		{"punctuation", "project;&|()#*?~[]"},
		{"unicode", "日本語"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config.CurrentDir = filepath.Join(base, tc.path, ".lenv", "current")
			settings := pathSettingsOutput(t)
			rc := filepath.Join(t.TempDir(), "lenvrc")
			if err := os.WriteFile(rc, []byte(settings), 0600); err != nil {
				t.Fatal(err)
			}
			for _, shell := range []string{"/bin/sh", "/bin/bash"} {
				for _, inherited := range []string{"", `/inherited path:$HOME:$(printf changed):` + "`printf changed`"} {
					values := testutil.SourcePathSettings(t, shell, rc, inherited)
					want := filepath.Join(config.CurrentDir, "bin") + ":" + filepath.Join(config.CurrentDir, "lua_modules/bin") + ":" + inherited
					if values[0] != want {
						t.Errorf("%s PATH = %q; want %q", shell, values[0], want)
					}
					assertLuaSettings(t, values[1], filepath.Join(config.CurrentDir, "lua_modules/lualib"), patterns["LUA_PATH"])
					assertLuaSettings(t, values[2], filepath.Join(config.CurrentDir, "lua_modules/luaclib"), patterns["LUA_CPATH"])
				}
			}
			// Target-specific commands return plain paths for non-shell consumers.
			want := filepath.Join(config.CurrentDir, "bin") + ":" + filepath.Join(config.CurrentDir, "lua_modules/bin") + "\n"
			if got := pathSettingsOutput(t, "bin"); got != want {
				t.Fatalf("raw bin path = %q; want %q", got, want)
			}
			assertLuaSettings(t, strings.TrimSuffix(pathSettingsOutput(t, "lualib"), "\n"), filepath.Join(config.CurrentDir, "lua_modules/lualib"), patterns["LUA_PATH"])
			assertLuaSettings(t, strings.TrimSuffix(pathSettingsOutput(t, "luaclib"), "\n"), filepath.Join(config.CurrentDir, "lua_modules/luaclib"), patterns["LUA_CPATH"])
		})
	}
}

func TestPathOutputWithoutLua(t *testing.T) {
	project := t.TempDir()
	for _, name := range []string{"src", "lua", "luajit"} {
		require.NoError(t, os.MkdirAll(filepath.Join(project, ".lenv", name), 0755))
	}
	oldCurrent := config.CurrentDir
	t.Cleanup(func() { config.CurrentDir = oldCurrent })
	current := filepath.Join(project, ".lenv", "current")
	config.CurrentDir = current
	t.Setenv("PATH", t.TempDir())
	t.Setenv("LUA_INIT", "error('must not run')")
	for _, key := range []string{"LUA_PATH", "LUA_CPATH", "LUA_PATH_5_5", "LUA_CPATH_5_5"} {
		t.Setenv(key, "unexpected")
	}
	expected := map[string]string{
		"bin":     filepath.Join(current, "bin") + ":" + filepath.Join(current, "lua_modules/bin"),
		"lualib":  filepath.Join(current, "lua_modules/lualib/?.lua") + ";" + filepath.Join(current, "lua_modules/lualib/?/init.lua") + ";;",
		"luaclib": filepath.Join(current, "lua_modules/luaclib/?.so") + ";;",
	}
	for _, target := range []string{"bin", "lualib", "luaclib"} {
		require.Equal(t, expected[target]+"\n", pathSettingsOutput(t, target))
	}
	rc := filepath.Join(t.TempDir(), "lenvrc")
	require.NoError(t, os.WriteFile(rc, []byte(pathSettingsOutput(t)), 0600))
	values := testutil.SourcePathSettings(t, "/bin/sh", rc, "/inherited/bin")
	require.Equal(t, []string{expected["bin"] + ":/inherited/bin", expected["lualib"], expected["luaclib"]}, values)
	require.NotContains(t, strings.Join(values, "\n"), "unexpected")
}
