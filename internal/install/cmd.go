package install

import (
	"fmt"
	"strings"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/env"
	"github.com/mah0x211/lenv/internal/fetch"
	targetmodel "github.com/mah0x211/lenv/internal/target"
)

func parseInstallOptions(args []string) (*options, error) {
	options := &options{}
	for len(args) > 0 && strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "--fetch":
			options.Fetch = true
		case "--if-missing":
			options.IfMissing = true
		case "--rebuild":
			options.Rebuild = true
		case "--no-rebuild":
			options.NoRebuild = true
		case "--allow-missing-gpg":
			options.AllowMissingGPG = true
		default:
			return nil, fmt.Errorf("unknown install option %q", args[0])
		}
		args = args[1:]
	}
	if options.Rebuild && options.NoRebuild {
		return nil, fmt.Errorf("--rebuild and --no-rebuild cannot be combined")
	}
	if len(args) == 0 || args[0] == "" || args[0] == ":" {
		return nil, fmt.Errorf("no version specified")
	}
	options.Version = args[0]
	options.Build = args[1:]
	return options, nil
}

// Resolve the requested version only after an optional successful refresh.
func prepareInstall(args []string) (*targetmodel.TargetVersion, *options) {
	options, err := parseInstallOptions(args)
	if err != nil {
		cli.Fatalf("%v", err)
	}
	if options.Fetch && !fetch.CmdFetch() {
		cli.Fatalf("version refresh failed; installation cancelled")
	}
	return targetmodel.PickTargetVersion(options.Version, false), options
}

func CmdInstall(opts []string) {
	target, options := prepareInstall(opts)
	if target.Lua != nil {
		doInstall(target.Lua.Config, target.Lua.Version, options)
	}
	if target.LuaRocks != nil {
		if err := env.ResolveCurrentDir(); err != nil {
			cli.Fatalf("%v", err)
		}
		config.CheckLuaRocksRootDir()
		doInstall(target.LuaRocks.Config, target.LuaRocks.Version, options)
	}
}
