package main

import (
	"path/filepath"
	"strconv"
	"strings"
)

func get_luacpath() string {
	prefix := filepath.Clean(CurrentDir+"/lua_modules/luaclib") + "/"
	paths := []string{}
	for i := 0; i <= 10; i++ {
		dir := prefix + strconv.Itoa(i)
		paths = append(paths, dir+"/?.so")
	}
	return strings.Join(paths, ";") + ";;"
}

func get_luapath() string {
	prefix := filepath.Clean(CurrentDir+"/lua_modules/lualib") + "/"
	paths := []string{}
	for i := 0; i <= 10; i++ {
		dir := prefix + strconv.Itoa(i)
		paths = append(paths, dir+"/?.lua", dir+"/?/init.lua")
	}
	return strings.Join(paths, ";") + ";;"
}

func get_path() string {
	return strings.Join([]string{
		filepath.Clean(CurrentDir + "/bin"),
		filepath.Clean(CurrentDir + "/lua_modules/bin"),
	}, ":")
}

func printAll() {
	printf(`
#
# please add the following lenv settings to your environment
#
`)

	for _, name := range []string{"PATH", "LUA_PATH", "LUA_CPATH"} {
		var value string
		switch name {
		case "PATH":
			value = get_path() + ":$PATH"

		case "LUA_PATH":
			value = get_luapath()

		case "LUA_CPATH":
			value = get_luacpath()
		}
		printf(`export %s="%s"`, name, value)
	}
	printf("")
}

func CmdPath(opts []string) {
	if len(opts) > 0 {
		switch opts[0] {
		case "bin":
			printf(get_path())
			return

		case "lualib":
			printf(get_luapath())
			return

		case "luaclib":
			printf(get_luacpath())
			return
		}
	}
	printAll()
}
