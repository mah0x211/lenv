package use

import (
	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/env"
	targetmodel "github.com/mah0x211/lenv/internal/target"
)

func CmdUse(opts []string) {
	if len(opts) == 0 {
		cli.Help(1, "no version specified")
	}
	target := targetmodel.PickTargetVersion(opts[0], false)
	if target.Lua != nil {
		env.UseInstalledVersion(target.Lua.Config, target.Lua.Version.Ver)
	}
	if target.LuaRocks != nil {
		config.CheckLuaRocksRootDir()
		env.UseInstalledVersion(target.LuaRocks.Config, target.LuaRocks.Version.Ver)
	}
}
