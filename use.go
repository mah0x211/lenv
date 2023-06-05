package main

import (
	"io/ioutil"
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

	infos, err := ioutil.ReadDir(rootdir)
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
			return
		}
	}

	fatalf("%s version %q is not installed", cfg.Name, ver)
}

func CmdUse(cfg *TargetConfig, opts []string) {
	// check target version
	if len(opts) == 0 || (cfg != LuaRocksCfg && opts[0] == ":") {
		CmdHelp(1, "no version specified")
	}
	ver := opts[0]

	// check :<luarocks-version>
	var rocksVer string
	if cfg != LuaRocksCfg {
		if delim := strings.Index(ver, ":"); delim != -1 {
			rocksVer = ver[delim+1:]
			ver = ver[:delim]
		}
	}

	var verItem *VerItem
	if len(ver) > 0 {
		verItem = PickTargetVersionItem(cfg, ver)
	}
	var rocksItem *VerItem
	if len(rocksVer) > 0 {
		rocksItem = PickTargetVersionItem(LuaRocksCfg, rocksVer)
	}

	if verItem != nil {
		UseInstalledVersion(cfg, verItem.Ver)
	}
	if rocksItem != nil {
		UseInstalledVersion(LuaRocksCfg, rocksItem.Ver)
	}
}
