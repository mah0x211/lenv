package testutil

import (
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"testing"
)

const ptyDriver = `
import errno, json, os, pty, select, subprocess, sys, time
command, prompt, answers, path, mutation = sys.argv[1:]
command, answers = json.loads(command), json.loads(answers) or []
master, slave = pty.openpty()
p = subprocess.Popen(command, stdin=slave, stdout=slave, stderr=slave)
os.close(slave)
output = bytearray()
answered = 0
deadline = time.monotonic() + 20
try:
    while True:
        if time.monotonic() > deadline:
            raise RuntimeError("terminal process timed out: " + output.decode(errors="replace"))
        ready, _, _ = select.select([master], [], [], 0.1)
        if ready:
            try:
                chunk = os.read(master, 65536)
            except OSError as e:
                if e.errno == errno.EIO:
                    break
                raise
            if not chunk:
                break
            output.extend(chunk)
            if output.count(prompt.encode()) > answered:
                if answered == 0 and mutation:
                    if mutation == "replace":
                        os.rename(path, path + ".saved")
                        with open(path, "w") as f:
                            f.write("replacement")
                    elif mutation == "move":
                        os.rename(path, path + ".saved")
                    elif mutation == "contents":
                        with open(os.path.join(path, "new-content"), "w") as f:
                            f.write("new")
                answer = answers[answered] if answered < len(answers) else "n"
                os.write(master, b"\x04" if answer == "EOF" else (answer + "\n").encode())
                answered += 1
        elif p.poll() is not None:
            break
    status = p.wait(timeout=5)
    print(json.dumps({"exit": status, "output": output.decode(errors="replace")}))
finally:
    os.close(master)
    if p.poll() is None:
        p.kill()
        p.wait()
`

type PTYResult struct {
	Exit   int    `json:"exit"`
	Output string `json:"output"`
}

// RunPTY drives a command through an actual pseudo-terminal. A requested
// change is applied to the caller's path immediately before the first answer.
func RunPTY(t testing.TB, command *exec.Cmd, prompt string, answers []string, path, mutation string) PTYResult {
	t.Helper()
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" && runtime.GOOS != "freebsd" {
		t.Skip("PTY tests support Linux, macOS, and FreeBSD")
	}
	argv, err := json.Marshal(append([]string{command.Path}, command.Args[1:]...))
	if err != nil {
		t.Fatal(err)
	}
	response, err := json.Marshal(answers)
	if err != nil {
		t.Fatal(err)
	}
	driver := exec.Command("python3", "-c", ptyDriver, string(argv), prompt, string(response), path, mutation)
	driver.Dir, driver.Env = command.Dir, command.Env
	data, err := driver.CombinedOutput()
	if err != nil {
		t.Fatalf("run terminal process: %v\n%s", err, data)
	}
	var result PTYResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("read terminal result: %v\n%s", err, data)
	}
	return result
}

const childArgsEnv = "LENV_TEST_CHILD_ARGS"

// TestChild creates a command for an entry point in the caller's test binary.
func TestChild(t testing.TB, name, directory string, args ...string) *exec.Cmd {
	t.Helper()
	data, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(os.Args[0], "-test.run=^"+name+"$")
	command.Dir = directory
	command.Env = append(os.Environ(), childArgsEnv+"="+string(data))
	return command
}

// TestChildArgs reports whether the current test was started by TestChild.
func TestChildArgs(t testing.TB) ([]string, bool) {
	t.Helper()
	data, ok := os.LookupEnv(childArgsEnv)
	if !ok {
		return nil, false
	}
	var args []string
	if err := json.Unmarshal([]byte(data), &args); err != nil {
		t.Fatal(err)
	}
	return args, true
}
