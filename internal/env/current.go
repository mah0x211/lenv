package env

import "github.com/mah0x211/lenv/internal/config"

// Used when a command selects a newly installed environment, after startup.
func ResolveCurrentDir() error {
	config.CurrentUsed, config.LuaRocksCfg.RootDir = "", ""
	diagnosis := diagnoseCurrent()
	if diagnosis.Err != nil {
		return diagnosis.Err
	}
	for _, step := range diagnosis.Steps {
		if err := StartupRecovery.apply(step); err != nil {
			return err
		}
	}
	diagnosis.SelectCurrent()
	return nil
}
