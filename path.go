package main

import (
	"path/filepath"
	"strings"
)

func get_luacpath() string {
	return filepath.Clean(CurrentDir+"/lua_modules/luaclib") + "/?.so;;"
}

func get_luapath() string {
	prefix := filepath.Clean(CurrentDir + "/lua_modules/lualib")
	paths := []string{
		prefix + "/?.lua",
		prefix + "/?/init.lua",
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
