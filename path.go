package main

import (
	"path/filepath"
	"strings"
)

// The final ;; lets the consuming Lua add its own default search paths.
func getLuaCPath() string {
	return filepath.Join(CurrentDir, "lua_modules/luaclib/?.so") + ";;"
}

func getLuaPath() string {
	prefix := filepath.Join(CurrentDir, "lua_modules/lualib")
	return strings.Join([]string{
		filepath.Join(prefix, "?.lua"),
		filepath.Join(prefix, "?/init.lua"),
	}, ";") + ";;"
}

func getPath() string {
	return strings.Join([]string{
		filepath.Join(CurrentDir, "bin"),
		filepath.Join(CurrentDir, "lua_modules/bin"),
	}, ":")
}

// Single quotes suppress shell expansion; emit embedded quotes separately.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func printAll() {
	println(`
#
# please add the following lenv settings to your environment
#`)

	for _, name := range []string{"PATH", "LUA_PATH", "LUA_CPATH"} {
		var value string
		switch name {
		case "PATH":
			value = shellQuote(getPath()) + `:"$PATH"`

		case "LUA_PATH":
			value = shellQuote(getLuaPath())

		case "LUA_CPATH":
			value = shellQuote(getLuaCPath())
		}
		printf(`export %s=%s`, name, value)
	}
	println("")
}

func CmdPath(opts []string) {
	if len(opts) > 0 {
		switch opts[0] {
		case "bin":
			println(getPath())
			return

		case "lualib":
			println(getLuaPath())
			return

		case "luaclib":
			println(getLuaCPath())
			return
		}
	}
	printAll()
}
