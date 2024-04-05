package main

import (
	"os"
	"path/filepath"
)

func uninstall(t *Target) {
	dir := filepath.Join(t.Config.RootDir, t.Version.Ver)
	stat, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			fatalf("failed to stat %q: %v", dir, err)
		}
		fatalf("%s version %s is not installed.", t.Config.Name, t.Version.Ver)
	}
	if !stat.IsDir() {
		fatalf("found %s %s (%q) but it is not a directory.\nplease remove it yourself.", t.Config.Name, t.Version.Ver, dir)
	} else if err = os.RemoveAll(dir); err != nil {
		fatalf("failed to uninstall version %s: %v", t.Version.Ver, err)
	}
	printf("%s version %s (%q) has been uninstalled.", t.Config.Name, t.Version.Ver, dir)
}

func CmdUninstall(opts []string) {
	target := PickTargetVersion(opts[0], true)

	// uninstall the specified version of lua
	if target.Lua != nil {
		uninstall(target.Lua)
		// it is remove all the versions of luarocks
		return
	}

	// uninstall the specified version of luarocks
	if target.LuaRocks != nil {
		CheckLuaRocksRootDir()
		uninstall(target.LuaRocks)
	}
}
