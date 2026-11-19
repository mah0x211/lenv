package uninstall

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/fs"
	targetmodel "github.com/mah0x211/lenv/internal/target"
)

func uninstall(t *targetmodel.Target) {
	dir := filepath.Join(t.Config.RootDir, t.Version.Ver)
	stat, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			cli.Fatalf("failed to stat %q: %v", dir, err)
		}
		cli.Fatalf("%s version %s is not installed.", t.Config.Name, t.Version.Ver)
	}
	if !stat.IsDir() {
		cli.Fatalf("found %s %s (%q) but it is not a directory.\nplease remove it yourself.", t.Config.Name, t.Version.Ver, dir)
	} else if err = os.RemoveAll(dir); err != nil {
		cli.Fatalf("failed to uninstall version %s: %v", t.Version.Ver, err)
	}
	cli.Printf("%s version %s (%q) has been uninstalled.", t.Config.Name, t.Version.Ver, dir)
}

func runUninstall(opts []string) {
	if len(opts) == 0 {
		cli.Help(1, "no version specified")
	}
	target := targetmodel.PickTargetVersion(opts[0], true)

	// uninstall the specified version of lua
	if target.Lua != nil {
		uninstall(target.Lua)
		// it is remove all the versions of luarocks
		return
	}

	// uninstall the specified version of luarocks
	if target.LuaRocks != nil {
		config.CheckLuaRocksRootDir()

		// check if lua_modules is a symlink to target luarocks version
		shouldRemoveSymlink := false
		symlink := fmt.Sprintf("%s/lua_modules", config.CurrentDir)
		rocksdir := filepath.Join(target.LuaRocks.Config.RootDir, target.LuaRocks.Version.Ver)
		if pathname, err := fs.EvalSymlink(symlink); err != nil && !errors.Is(err, fs.ErrNotSymlink) && !os.IsNotExist(err) {
			cli.Fatalf("failed to EvalSymlink(): %#v", err)
		} else {
			// A version-name prefix alone also matches sibling versions.
			shouldRemoveSymlink = pathname == rocksdir ||
				strings.HasPrefix(pathname, rocksdir+string(filepath.Separator))
		}

		uninstall(target.LuaRocks)

		if shouldRemoveSymlink {
			if err := os.Remove(symlink); err != nil {
				cli.Fatalf("failed to remove symlink %q: %v", symlink, err)
			}
			cli.Printf("symlink %q has been removed.", symlink)
		}
	}
}

func CmdUninstall(opts []string) {
	runUninstall(opts)
}
