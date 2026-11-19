package install

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/env"
	commandexec "github.com/mah0x211/lenv/internal/exec"
	"github.com/mah0x211/lenv/internal/fs"
	version "github.com/mah0x211/lenv/internal/versions"
)

func installRocks(instdir string, cfg *config.TargetConfig) error {
	luadir := filepath.Dir(cfg.RootDir)
	if err := commandexec.DoExec("./configure", "--help"); err != nil {
		return err
	}
	opts := []string{
		fmt.Sprintf("--prefix=%s", instdir),
		fmt.Sprintf("--with-lua=%s", luadir),
		// fmt.Sprintf("--with-lua-include=%s/include", luadir),
		// fmt.Sprintf("--with-lua-lib=%s/lib", luadir),
	}
	cli.Printf("./configure %s", strings.Join(opts, " "))
	if err := commandexec.DoExec("./configure", opts...); err != nil {
		return err
	} else if err = commandexec.DoExec("make"); err != nil {
		return err
	} else if err = commandexec.DoExec("make", "install"); err != nil {
		return err
	}

	cli.Println("postflight...")
	return postInstallLuaRocks(instdir)
}

func installLuaJit(instdir string, opts []string) error {
	if runtime.GOOS == "darwin" {
		// append MACOSX_DEPLOYMENT_TARGET=10.8 by default on macOS platform
		found := false
		for _, opt := range opts {
			found = strings.HasPrefix(opt, "MACOSX_DEPLOYMENT_TARGET")
			if found {
				break
			}
		}
		if !found {
			opts = append(opts, "MACOSX_DEPLOYMENT_TARGET=10.8")
		}
	}

	cli.Printf("make %s", strings.Join(opts, " "))
	if err := commandexec.DoExec("make", opts...); err != nil {
		return err
	}

	cli.Printf("make install PREFIX=%s", instdir)
	if err := commandexec.DoExec("make", "install", "PREFIX="+instdir); err != nil {
		return err
	}

	cli.Println("postflight...")
	return postInstallLuaJit(instdir)
}

func installLua(instdir string, opts []string) error {
	cli.Printf("make %s", strings.Join(opts, " "))
	if err := commandexec.DoExec("make", opts...); err != nil {
		return err
	}

	cli.Printf("make install INSTALL_TOP=%s", instdir)
	return commandexec.DoExec("make", "install", "INSTALL_TOP="+instdir)
}

func doInstall(cfg *config.TargetConfig, item *version.VerItem, options *options) {
	instdir := filepath.Join(cfg.RootDir, item.Ver)
	if options.IfMissing {
		installed, err := fs.IsDir(instdir)
		if err != nil {
			cli.Fatalf("failed to check installation directory: %v", err)
		}
		if installed {
			cli.Printf("%s version %s is already installed.", cfg.Name, item.Ver)
			env.UseInstalledVersion(cfg, item.Ver)
			return
		}
	}
	cli.Printf("install %q", item.Ver)
	dir, err := prepareBuildSource(cfg, item, options)
	if err != nil {
		cli.Fatalf("failed to prepare build sources: %v", err)
	}
	if !options.Rebuild {
		defer os.RemoveAll(dir)
	}

	cli.Printf("remove old directory: %s", instdir)
	if err := os.RemoveAll(instdir); err != nil {
		cli.Fatalf("failed to os.RemoveAll(): %v", err)
	}

	cli.Printf("chdir %q", dir)
	if err := os.Chdir(dir); err != nil {
		cli.Fatalf("failed to chdir(): %v", err)
	}

	switch cfg.Name {
	case "lua":
		err = installLua(instdir, options.Build)
	case "luajit":
		err = installLuaJit(instdir, options.Build)
	case "luarocks":
		err = installRocks(instdir, cfg)
	default:
		cli.Fatalf("unsupported target name %q", cfg.Name)
	}

	if err != nil {
		cli.Fatalf("failed to install %s version %s: %v", cfg.Name, item.Ver, err)
	}
	cli.Printf("\n%s version %s (%q) has been installed.", cfg.Name, item.Ver, instdir)

	// automatically use the installed version
	env.UseInstalledVersion(cfg, item.Ver)
}
