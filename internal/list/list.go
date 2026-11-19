package list

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/fs"
	version "github.com/mah0x211/lenv/internal/versions"
)

func listRocks(luadir string) {
	if err := os.Chdir(luadir); err != nil {
		cli.Eprintf("%v", err)
		return
	}

	var verUsed string
	if dir, err := fs.ReadSymlink("./lua_modules"); err != nil {
		if errors.Is(err, fs.ErrNotSymlink) {
			cli.Eprintf("%s is not symlink.\nplease remove it yourself", "lua_modules")
		}
		cli.Eprintf("failed to read symlink: %v", err)
		return
	} else if dir != "" {
		if !strings.HasPrefix(dir, "luarocks/") {
			cli.Eprintf("./lua_modules -> %s is not relative symlink to ./luarocks directory.\nplease remove it yourself", dir)
			return
		} else if ok, err := fs.IsDir(dir); err != nil {
			cli.Eprintf("failed to check %q type: %v", dir, err)
			return
		} else if !ok {
			cli.Eprintf("%s -> %s is not symlink to directory.\nplease remove it yourself", dir, dir)
			return
		}
		verUsed = filepath.Base(filepath.Dir(dir))
	}

	infos, err := ioutil.ReadDir("./luarocks")
	if err != nil {
		if !os.IsNotExist(err) {
			cli.Eprintf("failed to readdir: %v", err)
		}
		return
	}

	vers := []string{}
	maxlen := 0
	for _, info := range infos {
		ver := info.Name()
		if version.IsSemVer(ver) {
			vers = append(vers, ver)
			if len(ver) > maxlen {
				maxlen = len(ver)
			}
		}
	}
	version.SortVersions(vers)

	if len(vers) > 0 {
		format := fmt.Sprintf("    %%s %%-%ds", maxlen)
		cli.Println(" └ luarocks")
		tail := len(vers) - 1
		for i, v := range vers {
			c := "├"
			if i == tail {
				c = "└"
			}

			if v == verUsed {
				cli.Printf(format+" (used)", c, v)
			} else {
				cli.Printf(format, c, v)
			}
		}
	}
}

func printInstalled() {
	for _, cfg := range []*config.TargetConfig{
		config.LuaCfg, config.LuaJitCfg,
	} {
		cli.Printf("list installed %s and luarocks versions:", cfg.Name)
		infos, err := ioutil.ReadDir(cfg.RootDir)
		if err != nil {
			cli.Eprintf("failed to readdir: %v", err)
			continue
		}

		vers, err := version.NewVersionsFromFile(cfg.VersionFile)
		if err != nil {
			cli.Eprintf("failed to read version file %q: %v", cfg.VersionFile, err)
			continue
		}

		inst_vers := []string{}
		inst_dirs := []string{}
		maxlen := 0
		for _, info := range infos {
			ver := info.Name()
			dir := filepath.Join(cfg.RootDir, ver)

			// verify version
			if item := vers.GetItem(ver); item == nil {
				// ignore unknown file/directory
				cli.Eprintf("ignore unknown version: %s (%q)", ver, dir)
				continue
			} else if !info.IsDir() {
				cli.Eprintf("found %s %s (%q) but it is not a directory.\nplease remove it yourself.", cfg.Name, ver, dir)
				continue
			} else if len(ver) > maxlen {
				maxlen = len(ver)
			}

			inst_dirs = append(inst_dirs, dir)
			inst_vers = append(inst_vers, ver)
		}

		format := fmt.Sprintf("%%-%ds (%%s)", maxlen)
		for i, v := range inst_vers {
			dir := inst_dirs[i]
			if dir == config.CurrentUsed {
				cli.Printf(format+" (used)", v, dir)
			} else {
				cli.Printf(format, v, dir)
			}
			listRocks(dir)
		}
		cli.Println("")
	}
}

func CmdList() {
	printInstalled()
}
