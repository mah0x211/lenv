package env

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/fs"
)

func UseInstalledVersion(cfg *config.TargetConfig, ver string) {
	// change workdir
	wd := config.LenvDir
	dst := config.CurrentDir
	suffix := ""
	if cfg.Name == "luarocks" {
		wd = config.CurrentUsed
		dst = filepath.Join(wd, "lua_modules")
		suffix = "lua_modules"
	}

	if err := os.Chdir(wd); err != nil {
		cli.Fatalf("failed to chdir: %v", err)
	}

	// change rootdir from absolute to relative
	rootdir := strings.TrimPrefix(cfg.RootDir, wd)
	if rootdir != cfg.RootDir {
		rootdir = filepath.Clean("./" + rootdir)
	}

	infos, err := os.ReadDir(rootdir)
	if err != nil {
		cli.Fatalf("failed to readdir: %v", err)
	}
	for _, info := range infos {
		if info.Name() == ver {
			src := filepath.Join(rootdir, ver, suffix)
			if !info.IsDir() {
				cli.Fatalf("found %s %s (%q) but it is not a directory.\nplease remove it yourself.", cfg.Name, ver, src)
			} else if err := fs.CreateSymlink(src, dst); err != nil {
				cli.Fatalf("failed to create symlink: %v", err)
			}

			cli.Printf("use %s version %s (%q)", cfg.Name, ver, src)

			if err := os.Chdir(config.CWD); err != nil {
				cli.Fatalf("failed to chdir to %q: %v", config.CWD, err)
			} else if cfg.Name != "luarocks" {
				// resolve current dir if it is not luarocks
				if err := ResolveCurrentDir(); err != nil {
					cli.Fatalf("%v", err)
				}
			}
			return
		}
	}

	cli.Fatalf("%s version %q is not installed", cfg.Name, ver)
}
