package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getCurrentLuaVer() string {
	// determine current used lua version
	pathname, err := EvalSymlink(CurrentDir)
	if err != nil {
		if !errors.Is(err, ErrNotSymlink) && !os.IsNotExist(err) {
			fatalf("failed to EvalSymlink(): %#v", err)
		}
		// no current version set
		return ""
	}
	// extract binary name and version from pathname
	// pathname is like /home/user/.lenv/lua/5.4.4
	list := strings.Split(pathname, string(filepath.Separator))
	if len(list) < 2 {
		return ""
	}
	ver := list[len(list)-1]
	binname := list[len(list)-2]
	return fmt.Sprintf("%s/%s", binname, ver)
}

func getCurrentLuaRocksVer() string {
	// determine current used luarocks version
	pathname, err := EvalSymlink(fmt.Sprintf("%s/lua_modules", CurrentDir))
	if err != nil {
		if !errors.Is(err, ErrNotSymlink) && !os.IsNotExist(err) {
			fatalf("failed to EvalSymlink(): %#v", err)
		}
		// no current version set
		return ""
	}
	// extract binary name and version from pathname
	// pathname is like /home/user/.lenv/lua/5.4.4/luarocks/3.11.1/lua_modules
	list := strings.Split(pathname, string(filepath.Separator))
	if len(list) < 3 {
		return ""
	}
	ver := list[len(list)-2]
	binname := list[len(list)-3]
	return fmt.Sprintf("%s/%s", binname, ver)
}

func CmdCurrent(opts []string) {
	if len(opts) > 0 {
		switch opts[0] {
		case "lua":
			println(getCurrentLuaVer())
			return

		case "luarocks":
			println(getCurrentLuaRocksVer())
			return
		}
	}

	ver := getCurrentLuaVer()
	println(ver)
	if ver != "" {
		if ver = getCurrentLuaRocksVer(); ver != "" {
			println(ver)
		}
	}
}
