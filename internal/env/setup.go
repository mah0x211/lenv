package env

import (
	"fmt"

	"github.com/mah0x211/lenv/internal/cli"
)

func Setup(diagnosis *StartupDiagnosis, opts []string) {
	StartupRecovery.Enabled = false
	for _, opt := range opts {
		if opt != "--recover" {
			cli.Help(1, "unknown setup option %q", opt)
		}
		StartupRecovery.Enabled = true
	}
	if diagnosis.Err != nil {
		cli.Fatalf("Error: %v", diagnosis.Err)
	}
	for _, step := range diagnosis.Steps {
		if err := StartupRecovery.apply(step); err != nil {
			cli.Fatalf("Error: %v", err)
		}
	}
	if diagnosis.PendingCurrent {
		// This part was inaccessible or dependent on paths repaired above.
		// Previously inspected entries keep their original diagnosis.
		fmt.Fprintln(StartupRecovery.output, "Checking current selection...")
		diagnosis = diagnoseCurrent()
		if diagnosis.Err != nil {
			cli.Fatalf("Error: %v", diagnosis.Err)
		}
		for _, step := range diagnosis.Steps {
			if err := StartupRecovery.apply(step); err != nil {
				cli.Fatalf("Error: %v", err)
			}
		}
	}
	diagnosis.SelectCurrent()
	cli.Println("The required directories are ready.")
}
