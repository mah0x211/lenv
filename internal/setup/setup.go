package setup

import "github.com/mah0x211/lenv/internal/env"

func CmdSetup(diagnosis *env.StartupDiagnosis, opts []string) {
	env.Setup(diagnosis, opts)
}
