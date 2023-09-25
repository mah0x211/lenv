package main

import (
	"io"
	"os"
	"os/exec"
)

func DoExecEx(name string, stdout, stderr io.Writer, argv ...string) error {
	cmd := exec.Command(name, argv...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}

func DoExec(name string, argv ...string) error {
	return DoExecEx(name, os.Stdout, os.Stderr, argv...)
}
