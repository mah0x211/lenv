package path

import (
	"path/filepath"
	"strings"

	"github.com/mah0x211/lenv/internal/cli"
)

// The final ;; lets the consuming Lua add its own default search paths.
func getLuaCPath(currentDir string) string {
	return filepath.Join(currentDir, "lua_modules/luaclib/?.so") + ";;"
}

func getLuaPath(currentDir string) string {
	prefix := filepath.Join(currentDir, "lua_modules/lualib")
	return strings.Join([]string{
		filepath.Join(prefix, "?.lua"),
		filepath.Join(prefix, "?/init.lua"),
	}, ";") + ";;"
}

func getPath(currentDir string) string {
	return strings.Join([]string{
		filepath.Join(currentDir, "bin"),
		filepath.Join(currentDir, "lua_modules/bin"),
	}, ":")
}

// Single quotes suppress shell expansion; emit embedded quotes separately.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func printAll(currentDir string) {
	cli.Println(`
#
# please add the following lenv settings to your environment
#`)

	for _, name := range []string{"PATH", "LUA_PATH", "LUA_CPATH"} {
		var value string
		switch name {
		case "PATH":
			value = shellQuote(getPath(currentDir)) + `:"$PATH"`

		case "LUA_PATH":
			value = shellQuote(getLuaPath(currentDir))

		case "LUA_CPATH":
			value = shellQuote(getLuaCPath(currentDir))
		}
		cli.Printf(`export %s=%s`, name, value)
	}
	cli.Println("")
}

func CmdPath(currentDir string, opts []string) {
	if len(opts) > 0 {
		switch opts[0] {
		case "bin":
			cli.Println(getPath(currentDir))
			return

		case "lualib":
			cli.Println(getLuaPath(currentDir))
			return

		case "luaclib":
			cli.Println(getLuaCPath(currentDir))
			return
		}
	}
	printAll(currentDir)
}
