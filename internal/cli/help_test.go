package cli_test

import (
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_cmdHelp(t *testing.T) {
	exit := testutil.NewFakeExit(t)
	defer exit.Close()
	stdout := testutil.NewFakeStdout(t)
	defer stdout.CloseAll()

	// test that cmdHelp output to stdout and exit with rc
	cli.Help(123, "message")
	stdout.Close()
	exit.Close()
	assert.Equalf(t, 123, exit.Code, "not equal")
}
