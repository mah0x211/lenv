package vers

import (
	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/target"
)

func CmdVers() {
	for _, cfg := range []*config.TargetConfig{
		config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg,
	} {
		cli.Print("%s", target.ListTargetVersions(cfg))
		cli.Print("\n")
	}
}
