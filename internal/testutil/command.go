package testutil

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
)

// CaptureCommandOutput runs a command function in a goroutine so tests can
// observe functions that stop with runtime.Goexit through cli.Fatalf.
func CaptureCommandOutput(t testing.TB, command func()) (stdout, stderr string, completed bool) {
	t.Helper()
	directory := t.TempDir()
	stdoutFile, err := os.CreateTemp(directory, "stdout.*")
	if err != nil {
		t.Fatal(err)
	}
	stderrFile, err := os.CreateTemp(directory, "stderr.*")
	if err != nil {
		stdoutFile.Close()
		t.Fatal(err)
	}
	oldStdout, oldStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = stdoutFile, stderrFile
	t.Cleanup(func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
		stdoutFile.Close()
		stderrFile.Close()
	})

	type result struct {
		completed bool
		panic     any
	}
	done := make(chan result)
	go func() {
		finished := false
		var panicValue any
		defer func() {
			panicValue = recover()
			done <- result{completed: finished, panic: panicValue}
		}()
		command()
		finished = true
	}()
	commandResult := <-done
	os.Stdout, os.Stderr = oldStdout, oldStderr
	if commandResult.panic != nil {
		panic(commandResult.panic)
	}
	if _, err := stdoutFile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := stderrFile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	stdoutData, err := io.ReadAll(stdoutFile)
	if err != nil {
		t.Fatal(err)
	}
	stderrData, err := io.ReadAll(stderrFile)
	if err != nil {
		t.Fatal(err)
	}
	return string(stdoutData), string(stderrData), commandResult.completed
}

// CaptureHelpExit observes the output and exit code of a command that calls
// cli.Help. Terminating the command goroutine preserves its non-returning exit.
func CaptureHelpExit(t testing.TB, command func()) (string, int) {
	t.Helper()
	oldExit := cli.OsExit
	defer func() { cli.OsExit = oldExit }()
	code := -1
	cli.OsExit = func(status int) {
		code = status
		runtime.Goexit()
	}
	stdout, stderr, completed := CaptureCommandOutput(t, command)
	output := stdout + stderr
	if completed || code == -1 {
		t.Fatalf("command did not exit through help: %s", output)
	}
	return output, code
}

func SourcePathSettings(t testing.TB, shell, file, inherited string) []string {
	t.Helper()
	cmd := exec.Command(shell, "-c", `. "$1"; printf '%s\000%s\000%s\000' "$PATH" "$LUA_PATH" "$LUA_CPATH"`, "path-settings", file)
	// The inherited values deliberately differ from the generating process.
	// Quoted expansions must preserve them without interpreting their contents.
	cmd.Env = []string{"PATH=" + inherited, "LUA_PATH=previous", "LUA_CPATH=previous"}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	data, err := cmd.Output()
	if err != nil {
		t.Fatalf("source settings with %s: %v\n%s", shell, err, stderr.String())
	}
	values := strings.Split(string(data), "\x00")
	if len(values) != 4 || values[3] != "" {
		t.Fatalf("invalid environment output: %q", data)
	}
	return values[:3]
}
