package main

import (
	"path/filepath"
	"strconv"
	"strings"
)

func CmdPath() {
	printf(`
#
# please add the following lenv settings to your environment
#
`)

	for _, name := range []string{"PATH", "LUA_PATH", "LUA_CPATH"} {
		var (
			format string
			value  string
		)

		switch name {
		case "PATH":
			format = `export %s="%s"`
			value = strings.Join([]string{
				filepath.Clean(CurrentDir + "/bin"),
				filepath.Clean(CurrentDir + "/lua_modules/bin"),
				"$PATH",
			}, ":")

		case "LUA_PATH":
			format = `export %s="%s;"`
			prefix := filepath.Clean(CurrentDir+"/lua_modules/lualib") + "/"
			paths := []string{}
			for i := 0; i <= 10; i++ {
				dir := prefix + strconv.Itoa(i)
				paths = append(paths, dir+"/?.lua", dir+"/?/init.lua")
			}
			value = strings.Join(paths, ";") + ";"

		case "LUA_CPATH":
			format = `export %s="%s;"`
			prefix := filepath.Clean(CurrentDir+"/lua_modules/luaclib") + "/"
			paths := []string{}
			for i := 0; i <= 10; i++ {
				dir := prefix + strconv.Itoa(i)
				paths = append(paths, dir+"/?.so")
			}
			value = strings.Join(paths, ";") + ";"
		}
		printf(format, name, value)
	}
	printf("")
}
