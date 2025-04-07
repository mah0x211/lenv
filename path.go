package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetLuaPackagePaths() map[string][]string {
	// get luarocks pathnames
	var buf bytes.Buffer
	if err := DoExecEx("lua", &buf, os.Stderr, "-e", `
for k, s in  pairs({
    LUA_PATH = package.path,
    LUA_CPATH = package.cpath,
}) do
    io.write(k .. '=')
    local patterns = {}
    s:gsub('%?/?[^;/]+', function(v)
        if not patterns[v] then
            patterns[v] = true
            patterns[#patterns + 1] = v
        end
    end)
    table.sort(patterns)
    io.write(table.concat(patterns, ';'), '\n')
end
	`); err != nil {
		fatalf("failed to get lua path: %v", err)
	}

	pathnames := map[string][]string{}
	for _, line := range strings.Split(buf.String(), "\n") {
		arr := strings.SplitN(line, "=", 2)
		if len(arr) != 2 {
			continue
		}
		name := arr[0]
		paths := arr[1]

		switch name {
		case "LUA_PATH", "LUA_CPATH":
			for _, v := range strings.Split(paths, ";") {
				// extract object extension
				pathname := pathnames[name]
				if pathname == nil {
					pathname = []string{}
				}
				pathname = append(pathname, v)
				pathnames[name] = pathname
			}
		}
	}

	return pathnames
}

func getLuaCPath(rockspath map[string][]string) string {
	if len(rockspath) == 0 {
		// extract LUA_CPATH from luarocks
		rockspath = GetLuaPackagePaths()
	}
	cpath := rockspath["LUA_CPATH"]
	if cpath == nil {
		fatalf("failed to get LUA_CPATH from luarocks path command")
	}

	// get unique basenames
	basenames := map[string]bool{}
	for _, basename := range cpath {
		basenames[basename] = true
	}

	prefix := filepath.Join(CurrentDir, "lua_modules/luaclib")
	pathnames := []string{}
	for basename := range basenames {
		pathnames = append(pathnames, fmt.Sprintf("%s/%s", prefix, basename))
	}
	return strings.Join(pathnames, ";") + ";;"
}

func getLuaPath(rockspath map[string][]string) string {
	if len(rockspath) == 0 {
		// extract LUA_PATH from luarocks
		rockspath = GetLuaPackagePaths()
	}
	luapath := rockspath["LUA_PATH"]
	if luapath == nil {
		fatalf("failed to get LUA_PATH")
	}

	// get unique basenames
	basenames := map[string]bool{}
	for _, basename := range luapath {
		basenames[basename] = true
	}

	prefix := filepath.Join(CurrentDir, "lua_modules/lualib")
	pathnames := []string{}
	for basename := range basenames {
		pathnames = append(pathnames, fmt.Sprintf("%s/%s", prefix, basename))
	}
	return strings.Join(pathnames, ";") + ";;"
}

func getPath() string {
	return strings.Join([]string{
		filepath.Join(CurrentDir, "bin"),
		filepath.Join(CurrentDir, "lua_modules/bin"),
	}, ":")
}

func printAll() {
	rockspath := GetLuaPackagePaths()

	printf(`
#
# please add the following lenv settings to your environment
#
`)

	for _, name := range []string{"PATH", "LUA_PATH", "LUA_CPATH"} {
		var value string
		switch name {
		case "PATH":
			value = getPath() + ":$PATH"

		case "LUA_PATH":
			value = getLuaPath(rockspath)

		case "LUA_CPATH":
			value = getLuaCPath(rockspath)
		}
		printf(`export %s="%s"`, name, value)
	}
	printf("")
}

func CmdPath(opts []string) {
	if len(opts) > 0 {
		switch opts[0] {
		case "bin":
			printf(getPath())
			return

		case "lualib":
			printf(getLuaPath(nil))
			return

		case "luaclib":
			printf(getLuaCPath(nil))
			return
		}
	}
	printAll()
}
