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

func TestHelpRejectsNonStringMessage(t *testing.T) {
	oldCode := cli.ExitCode
	t.Cleanup(func() { cli.ExitCode = oldCode })
	cli.ExitCode = 0
	stdout, stderr, completed := testutil.CaptureCommandOutput(t, func() { cli.Help(0, 42) })
	assert.False(t, completed)
	assert.Equal(t, 1, cli.ExitCode)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "invalid arguments")
}
