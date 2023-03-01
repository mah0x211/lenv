package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func listRocks(luadir string) {
	if err := os.Chdir(luadir); err != nil {
		eprintf(err.Error())
		return
	}

	var verUsed string
	if dir, err := readSymlink("./lua_modules"); err != nil {
		if errors.Is(err, ErrNotSymlink) {
			eprintf("%s is not symlink.\nplease remove it yourself", "lua_modules")
		}
		eprintf("failed to read symlink: %v", err)
		return
	} else if dir != "" {
		if !strings.HasPrefix(dir, "luarocks/") {
			eprintf("./lua_modules -> %s is not relative symlink to ./luarocks directory.\nplease remove it yourself", dir)
			return
		} else if ok, err := isDir(dir); err != nil {
			eprintf("failed to check %q type: %v", dir, err)
			return
		} else if !ok {
			eprintf("%s -> %s is not symlink to directory.\nplease remove it yourself", dir, dir)
			return
		}
		verUsed = filepath.Base(filepath.Dir(dir))
	}

	infos, err := ioutil.ReadDir("./luarocks")
	if err != nil {
		if !os.IsNotExist(err) {
			eprintf("failed to readdir: %v", err)
		}
		return
	}

	vers := []string{}
	maxlen := 0
	for _, info := range infos {
		ver := info.Name()
		if isSemVer(ver) {
			vers = append(vers, ver)
			if len(ver) > maxlen {
				maxlen = len(ver)
			}
		}
	}
	sortVersions(vers)

	if len(vers) > 0 {
		format := fmt.Sprintf("    %%s %%-%ds", maxlen)
		printf(" └ luarocks")
		tail := len(vers) - 1
		for i, v := range vers {
			c := "├"
			if i == tail {
				c = "└"
			}

			if v == verUsed {
				printf(format+" (used)", c, v)
			} else {
				printf(format, c, v)
			}
		}
	}
}

func cmdList() {
	for _, cfg := range []*TargetConfig{
		LuaCfg, LuaJitCfg,
	} {
		printf("list installed %s and luarocks versions:", cfg.Name)
		infos, err := ioutil.ReadDir(cfg.RootDir)
		if err != nil {
			eprintf("failed to readdir: %v", err)
			continue
		}

		vers := []string{}
		dirs := []string{}
		maxlen := 0
		for _, info := range infos {
			ver := info.Name()
			dir := filepath.Join(cfg.RootDir, ver)

			// verify version
			if item, err := getVerInfo(cfg.VersionFile, ver); err != nil {
				eprintf("failed to get version info: %v", err)
				break
			} else if item == nil {
				// ignore unknown file/directory
				eprintf("ignore unknown version: %s (%q)", ver, dir)
				continue
			} else if !info.IsDir() {
				eprintf("found %s %s (%q) but it is not a directory.\nplease remove it yourself.", cfg.Name, ver, dir)
				continue
			} else if len(ver) > maxlen {
				maxlen = len(ver)
			}

			dirs = append(dirs, dir)
			vers = append(vers, ver)
		}

		format := fmt.Sprintf("%%-%ds (%%s)", maxlen)
		for i, v := range vers {
			dir := dirs[i]
			if dir == CurrentUsed {
				printf(format+" (used)", v, dir)
			} else {
				printf(format, v, dir)
			}
			listRocks(dir)
		}
		printf("")
	}
}
