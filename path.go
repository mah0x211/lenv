package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetLuaRocksPath() map[string][]map[string]string {
	// get luarocks pathnames
	var buf bytes.Buffer
	if err := DoExecEx("luarocks", &buf, os.Stderr, "path"); err != nil {
		fatalf("failed to get luarocks path: %v", err)
	}

	pathnames := map[string][]map[string]string{}
	for _, line := range strings.Split(buf.String(), "\n") {
		arr := strings.SplitN(line, "='", 2)
		if len(arr) != 2 || !strings.HasPrefix(arr[0], "export ") {
			continue
		}
		name := strings.TrimPrefix(arr[0], "export ")
		paths := strings.TrimSuffix(arr[1], "'")

		switch name {
		case "PATH":
			for _, v := range strings.Split(paths, ":") {
				pathname := pathnames[name]
				if pathname == nil {
					pathname = []map[string]string{}
					pathnames[name] = pathname
				}
				pathname = append(pathname, map[string]string{
					"pathname": v,
				})
				// sort by pathname
				// sort.Slice(pathname, func(i, j int) bool {
				// 	return pathname[i]["pathname"] < pathname[j]["pathname"]
				// })
			}

		case "LUA_PATH", "LUA_CPATH":
			for _, v := range strings.Split(paths, ";") {
				// extract object extension
				kv := strings.SplitN(v, "/?", 2)
				if len(kv) == 2 {
					pathname := pathnames[name]
					if pathname == nil {
						pathname = []map[string]string{}
					}
					pathname = append(pathname, map[string]string{
						"pathname": v,
						"dirname":  kv[0],
						"basename": "?" + kv[1],
					})
					// sort by pathname
					// sort.Slice(pathname, func(i, j int) bool {
					// 	return pathname[i]["pathname"] < pathname[j]["pathname"]
					// })
					pathnames[name] = pathname
				}
			}
		}
	}

	return pathnames
}

func getLuaCPath(rockspath map[string][]map[string]string) string {
	if len(rockspath) == 0 {
		// extract LUA_CPATH from luarocks
		rockspath = GetLuaRocksPath()
	}
	cpath := rockspath["LUA_CPATH"]
	if cpath == nil {
		fatalf("failed to get LUA_CPATH from luarocks path command")
	}

	// get unique basenames
	basenames := map[string]bool{}
	for _, pathname := range cpath {
		basenames[pathname["basename"]] = true
	}

	prefix := filepath.Join(CurrentDir, "lua_modules/luaclib")
	pathnames := []string{}
	for basename := range basenames {
		pathnames = append(pathnames, fmt.Sprintf("%s/%s", prefix, basename))
	}
	return strings.Join(pathnames, ";") + ";;"
}

func getLuaPath(rockspath map[string][]map[string]string) string {
	if len(rockspath) == 0 {
		// extract LUA_PATH from luarocks
		rockspath = GetLuaRocksPath()
	}
	luapath := rockspath["LUA_PATH"]
	if luapath == nil {
		fatalf("failed to get LUA_PATH")
	}

	// get unique basenames
	basenames := map[string]bool{}
	for _, pathname := range luapath {
		basenames[pathname["basename"]] = true
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
	rockspath := GetLuaRocksPath()

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
