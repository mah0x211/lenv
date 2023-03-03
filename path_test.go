package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_cmdPath(t *testing.T) {
	stdout := NewFakeStdout(t)
	defer stdout.CloseAll()

	// test that cmdPath outputs PATH, LUA_PATH and LUA_CPATH to stdout
	CmdPath()
	var b bytes.Buffer
	if _, err := b.ReadFrom(stdout.Close()); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	s := b.String()
	assert.Regexpf(t, "(?m)^export PATH=", s, "not found")
	assert.Regexpf(t, "(?m)^export LUA_PATH=", s, "not found")
	assert.Regexpf(t, "(?m)^export LUA_CPATH=", s, "not found")
}
