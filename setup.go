package main

import "fmt"

func CmdSetup(diagnosis *startupDiagnosis, opts []string) {
	startupRecovery.enabled = false
	for _, opt := range opts {
		if opt != "--recover" {
			CmdHelp(1, "unknown setup option %q", opt)
		}
		startupRecovery.enabled = true
	}
	if diagnosis.err != nil {
		fatalf("Error: %v", diagnosis.err)
	}
	for _, step := range diagnosis.steps {
		if err := startupRecovery.apply(step); err != nil {
			fatalf("Error: %v", err)
		}
	}
	if diagnosis.pendingCurrent {
		// This part was inaccessible or dependent on paths repaired above.
		// Previously inspected entries keep their original diagnosis.
		fmt.Fprintln(startupRecovery.output, "Checking current selection...")
		diagnosis = diagnoseCurrent()
		if diagnosis.err != nil {
			fatalf("Error: %v", diagnosis.err)
		}
		for _, step := range diagnosis.steps {
			if err := startupRecovery.apply(step); err != nil {
				fatalf("Error: %v", err)
			}
		}
	}
	diagnosis.selectCurrent()
	println("The required directories are ready.")
}
