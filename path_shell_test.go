package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func pathSettingsOutput(t *testing.T, args ...string) string {
	t.Helper()
	stdout := NewFakeStdout(t)
	defer stdout.CloseAll()
	CmdPath(args)
	data, err := os.ReadFile(stdout.Close().Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func sourcePathSettings(t *testing.T, shell, file, inherited string) []string {
	t.Helper()
	cmd := exec.Command(shell, "-c", `. "$1"; printf '%s\000%s\000%s\000' "$PATH" "$LUA_PATH" "$LUA_CPATH"`, "path-settings", file)
	// The inherited values deliberately differ from the generating process.
	// Quoted expansions must preserve them without interpreting their contents.
	cmd.Env = []string{"PATH=" + inherited, "LUA_PATH=previous", "LUA_CPATH=previous"}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	data, err := cmd.Output()
	if err != nil {
		t.Fatalf("source settings with %s: %v\n%s", shell, err, stderr.String())
	}
	values := strings.Split(string(data), "\x00")
	if len(values) != 4 || values[3] != "" {
		t.Fatalf("invalid environment output: %q", data)
	}
	return values[:3]
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
	old := CurrentDir
	t.Cleanup(func() { CurrentDir = old })
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
			CurrentDir = filepath.Join(base, tc.path, ".lenv", "current")
			settings := pathSettingsOutput(t)
			rc := filepath.Join(t.TempDir(), "lenvrc")
			if err := os.WriteFile(rc, []byte(settings), 0600); err != nil {
				t.Fatal(err)
			}
			for _, shell := range []string{"/bin/sh", "/bin/bash"} {
				for _, inherited := range []string{"", `/inherited path:$HOME:$(printf changed):` + "`printf changed`"} {
					values := sourcePathSettings(t, shell, rc, inherited)
					want := filepath.Join(CurrentDir, "bin") + ":" + filepath.Join(CurrentDir, "lua_modules/bin") + ":" + inherited
					if values[0] != want {
						t.Errorf("%s PATH = %q; want %q", shell, values[0], want)
					}
					assertLuaSettings(t, values[1], filepath.Join(CurrentDir, "lua_modules/lualib"), patterns["LUA_PATH"])
					assertLuaSettings(t, values[2], filepath.Join(CurrentDir, "lua_modules/luaclib"), patterns["LUA_CPATH"])
				}
			}
			// Target-specific commands return plain paths for non-shell consumers.
			if got := pathSettingsOutput(t, "bin"); got != getPath()+"\n" {
				t.Fatalf("raw bin path = %q; want %q", got, getPath()+"\n")
			}
			assertLuaSettings(t, strings.TrimSuffix(pathSettingsOutput(t, "lualib"), "\n"), filepath.Join(CurrentDir, "lua_modules/lualib"), patterns["LUA_PATH"])
			assertLuaSettings(t, strings.TrimSuffix(pathSettingsOutput(t, "luaclib"), "\n"), filepath.Join(CurrentDir, "lua_modules/luaclib"), patterns["LUA_CPATH"])
		})
	}
}

func TestPathShellCommandSubprocess(t *testing.T) {
	if os.Getenv("LENV_PATH_SHELL_PROCESS") != "1" {
		return
	}
	os.Args = []string{"lenv", "-p", "path"}
	main()
}

func TestPathShellSettingsFromProjectCommand(t *testing.T) {
	base := t.TempDir()
	project := filepath.Join(base, "project ' \"$HOME $(printf changed) `printf changed`")
	for _, name := range []string{"src", "lua", "luajit"} {
		if err := os.MkdirAll(filepath.Join(project, ".lenv", name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	// main discovers the physical working directory, including on macOS where
	// temporary directory names may start with the /var alias.
	project, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatal(err)
	}
	cmd := commandForLenvTest("TestPathShellCommandSubprocess", "-p", "path")
	cmd.Dir = project
	cmd.Env = append(os.Environ(), "LENV_PATH_SHELL_PROCESS=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	settings, err := cmd.Output()
	if err != nil {
		t.Fatalf("lenv -p path: %v\n%s", err, stderr.String())
	}
	rc := filepath.Join(base, "lenvrc")
	if err := os.WriteFile(rc, settings, 0600); err != nil {
		t.Fatal(err)
	}
	for _, shell := range []string{"/bin/sh", "/bin/bash"} {
		values := sourcePathSettings(t, shell, rc, "/inherited/bin")
		current := filepath.Join(project, ".lenv", "current")
		want := filepath.Join(current, "bin") + ":" + filepath.Join(current, "lua_modules/bin") + ":/inherited/bin"
		if values[0] != want {
			t.Fatalf("%s project PATH = %q; want %q", shell, values[0], want)
		}
		for i, suffix := range []string{"lua_modules/lualib/", "lua_modules/luaclib/"} {
			if !strings.HasPrefix(values[i+1], filepath.Join(current, suffix)+string(filepath.Separator)) {
				t.Fatalf("project Lua setting %q lacks expected directory", values[i+1])
			}
			if !strings.HasSuffix(values[i+1], ";;") {
				t.Fatalf("project Lua setting %q lacks default search path", values[i+1])
			}
		}
	}
}
