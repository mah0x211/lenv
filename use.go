package main

import (
	"os"
	"path/filepath"
	"strings"
)

func UseInstalledVersion(cfg *TargetConfig, ver string) {
	// change workdir
	wd := LenvDir
	dst := CurrentDir
	suffix := ""
	if cfg.Name == "luarocks" {
		wd = CurrentUsed
		dst = filepath.Join(wd, "lua_modules")
		suffix = "lua_modules"
	}

	if err := os.Chdir(wd); err != nil {
		fatalf("failed to chdir: %v", err)
	}

	// change rootdir from absolute to relative
	rootdir := strings.TrimPrefix(cfg.RootDir, wd)
	if rootdir != cfg.RootDir {
		rootdir = filepath.Clean("./" + rootdir)
	}

	infos, err := os.ReadDir(rootdir)
	if err != nil {
		fatalf("failed to readdir: %v", err)
	}
	for _, info := range infos {
		if info.Name() == ver {
			src := filepath.Join(rootdir, ver, suffix)
			if !info.IsDir() {
				fatalf("found %s %s (%q) but it is not a directory.\nplease remove it yourself.", cfg.Name, ver, src)
			} else if err := createSymlink(src, dst); err != nil {
				fatalf("failed to create symlink: %v", err)
			}

			printf("use %s version %s (%q)", cfg.Name, ver, src)

			if err := os.Chdir(CWD); err != nil {
				fatalf("failed to chdir to %q: %v", CWD, err)
			} else if cfg.Name != "luarocks" {
				// resolve current dir if it is not luarocks
				ResolveCurrentDir()
			}
			return
		}
	}

	fatalf("%s version %q is not installed", cfg.Name, ver)
}

func CmdUse(opts []string) {
	target := PickTargetVersion(opts[0], false)
	if target.Lua != nil {
		UseInstalledVersion(target.Lua.Config, target.Lua.Version.Ver)
	}
	if target.LuaRocks != nil {
		CheckLuaRocksRootDir()
		UseInstalledVersion(target.LuaRocks.Config, target.LuaRocks.Version.Ver)
	}
}
