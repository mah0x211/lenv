package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_cmdHelp(t *testing.T) {
	exit := NewFakeExit(t)
	defer exit.Close()
	stdout := NewFakeStdout(t)
	defer stdout.CloseAll()

	// test that cmdHelp output to stdout and exit with rc
	cmdHelp(123, "message")
	stdout.Close()
	exit.Close()
	assert.Equalf(t, 123, exit.Code, "not equal")
}
