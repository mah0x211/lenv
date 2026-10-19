package main

import (
	"os"
	"os/exec"
)

// Prefer the actual instrumented CLI when coverage is collected by make test.
// Direct go test runs keep using the existing subprocess entry points.
func commandForLenvTest(helper string, args ...string) *exec.Cmd {
	if executable := os.Getenv("LENV_TEST_EXECUTABLE"); executable != "" {
		return exec.Command(executable, args...)
	}
	return exec.Command(os.Args[0], "-test.run=^"+helper+"$")
}
