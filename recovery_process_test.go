package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Python supplies a real pseudo-terminal; all recovery decisions run in lenv.
const recoveryPTY = `
import errno, json, os, pty, select, subprocess, sys, time
exe, directory, answers, mutation, args = sys.argv[1:]
answers, args = json.loads(answers), json.loads(args)
master, slave = pty.openpty()
env = os.environ.copy()
env["LENV_RECOVERY_PROCESS"] = "1"
env["LENV_RECOVERY_ARGS"] = json.dumps(args)
command = [exe, "-test.run=^TestRecoverySubprocess$"]
executable = env.get("LENV_TEST_EXECUTABLE") or env.get("LENV_RECOVERY_EXECUTABLE")
if executable:
    command = [executable, *args]
p = subprocess.Popen(command, cwd=directory,
                     stdin=slave, stdout=slave, stderr=slave, env=env)
os.close(slave)
output = bytearray()
answered = 0
deadline = time.monotonic() + 20
try:
    while True:
        if time.monotonic() > deadline:
            raise RuntimeError("recovery process timed out: " + output.decode(errors="replace"))
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
            prompts = output.count(b"Delete and recover? [y/N] ")
            if prompts > answered:
                if answered == 0 and mutation:
                    path = os.path.join(directory, ".lenv", "current")
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

func TestRecoverySubprocess(t *testing.T) {
	if os.Getenv("LENV_RECOVERY_PROCESS") != "1" {
		return
	}
	var args []string
	if err := json.Unmarshal([]byte(os.Getenv("LENV_RECOVERY_ARGS")), &args); err != nil {
		panic(err)
	}
	os.Args = append([]string{"lenv"}, args...)
	main()
}

type recoveryProcessResult struct {
	Exit   int    `json:"exit"`
	Output string `json:"output"`
}

func recoveryProcess(t *testing.T, project string, answers []string, mutation string, args ...string) recoveryProcessResult {
	t.Helper()
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" && runtime.GOOS != "freebsd" {
		t.Skip("interactive recovery supports Linux, macOS, and FreeBSD")
	}
	answerJSON, err := json.Marshal(answers)
	require.NoError(t, err)
	argJSON, err := json.Marshal(args)
	require.NoError(t, err)
	cmd := exec.Command("python3", "-c", recoveryPTY, os.Args[0], project, string(answerJSON), mutation, string(argJSON))
	data, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", data)
	var result recoveryProcessResult
	require.NoError(t, json.Unmarshal(data, &result), "%s", data)
	return result
}

func recoveryProject(t *testing.T, kind string) (string, string) {
	t.Helper()
	project, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	root := filepath.Join(project, ".lenv")
	target := filepath.Join(project, "kept")
	require.NoError(t, os.Mkdir(target, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "marker"), []byte("keep"), 0600))
	if kind == "root-file" {
		require.NoError(t, os.WriteFile(root, nil, 0600))
		return project, target
	}
	for _, dir := range []string{"src", "lua/5.5.1", "luajit"} {
		require.NoError(t, os.MkdirAll(filepath.Join(root, dir), 0755))
	}
	current := filepath.Join(root, "current")
	switch kind {
	case "current-directory":
		require.NoError(t, os.Mkdir(current, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(current, "marker"), []byte("keep"), 0600))
	case "current-file":
		require.NoError(t, os.WriteFile(current, []byte("keep"), 0600))
	case "current-dangling":
		require.NoError(t, os.Symlink("missing", current))
	case "current-loop":
		require.NoError(t, os.Symlink("current", current))
	case "current-intermediate-file":
		target = filepath.Join(root, "lua", "entry")
		require.NoError(t, os.WriteFile(target, []byte("keep"), 0600))
		require.NoError(t, os.Symlink("lua/entry/subdir", current))
	case "current-to-file":
		target = filepath.Join(project, "kept-file")
		require.NoError(t, os.WriteFile(target, []byte("keep"), 0600))
		require.NoError(t, os.Symlink(target, current))
	case "rocks-file", "rocks-link", "rocks-dangling":
		lua := filepath.Join(root, "lua/5.5.1")
		require.NoError(t, os.Symlink("lua/5.5.1", current))
		rocks := filepath.Join(lua, "luarocks")
		switch kind {
		case "rocks-file":
			require.NoError(t, os.WriteFile(rocks, []byte("keep"), 0600))
		case "rocks-link":
			require.NoError(t, os.Symlink(target, rocks))
		case "rocks-dangling":
			require.NoError(t, os.Symlink("missing", rocks))
		}
	case "required-file", "required-link", "multiple":
		src := filepath.Join(root, "src")
		require.NoError(t, os.Remove(src))
		if kind == "required-link" {
			require.NoError(t, os.Symlink(target, src))
		} else {
			require.NoError(t, os.WriteFile(src, []byte("keep"), 0600))
		}
		if kind == "multiple" {
			jit := filepath.Join(root, "luajit")
			require.NoError(t, os.Remove(jit))
			require.NoError(t, os.WriteFile(jit, nil, 0600))
		}
	}
	return project, target
}

func TestInteractiveRecoveryApproval(t *testing.T) {
	kinds := []string{"current-directory", "current-file", "current-dangling", "current-loop", "current-to-file", "current-intermediate-file",
		"rocks-file", "rocks-link", "rocks-dangling", "required-file", "required-link", "root-file", "multiple"}
	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			project, kept := recoveryProject(t, kind)
			root := filepath.Join(project, ".lenv")
			path := filepath.Join(root, "current")
			wantType, wantAfter := "symbolic link", "leave Lua and LuaJIT unselected"
			switch {
			case strings.HasPrefix(kind, "rocks-"):
				path = filepath.Join(root, "lua/5.5.1/luarocks")
				wantAfter = "create an actual directory here"
			case kind == "root-file":
				path = root
				wantAfter = "create an actual directory here"
			case strings.HasPrefix(kind, "required-") || kind == "multiple":
				path = filepath.Join(root, "src")
				wantAfter = "create an actual directory here"
			}
			switch kind {
			case "current-directory":
				wantType = "directory"
			case "current-file", "rocks-file", "required-file", "root-file", "multiple":
				wantType = "regular file"
			}
			var linkTarget string
			if wantType == "symbolic link" {
				var err error
				linkTarget, err = os.Readlink(path)
				require.NoError(t, err)
			}
			result := recoveryProcess(t, project, []string{"yes", "y"}, "", "-p", "setup", "--recover")
			require.Equal(t, 0, result.Exit, result.Output)
			output := strings.ReplaceAll(result.Output, "\r\n", "\n")
			require.Contains(t, output, fmt.Sprintf("\nPath: %q\nType: %s\n", path, wantType))
			require.Contains(t, output, "\nAfter recovery: "+wantAfter+"\n")
			switch wantType {
			case "symbolic link":
				require.Contains(t, output, fmt.Sprintf("\nLink target: %q\n", linkTarget))
				require.Contains(t, output, "Delete only this link; its target will be kept.\n")
			case "directory":
				require.Contains(t, output, "Delete this directory and ALL of its contents.\n")
			case "regular file":
				require.Contains(t, output, "Delete this file and its contents.\n")
			}
			for _, name := range []string{"", "src", "lua", "luajit"} {
				info, err := os.Lstat(filepath.Join(root, name))
				require.NoError(t, err)
				require.True(t, info.IsDir())
			}
			if strings.HasPrefix(kind, "current-") {
				_, err := os.Lstat(filepath.Join(root, "current"))
				require.True(t, os.IsNotExist(err))
			}
			if strings.HasPrefix(kind, "rocks-") {
				info, err := os.Lstat(filepath.Join(root, "lua/5.5.1/luarocks"))
				require.NoError(t, err)
				require.True(t, info.IsDir())
			}
			info, err := os.Stat(kept)
			require.NoError(t, err)
			if info.IsDir() {
				kept = filepath.Join(kept, "marker")
			}
			data, err := os.ReadFile(kept)
			require.NoError(t, err)
			require.Equal(t, "keep", string(data))
		})
	}
}

func TestInteractiveRecoveryCancellation(t *testing.T) {
	for _, answer := range []string{"n", "", "EOF", "not-yes"} {
		t.Run(answer, func(t *testing.T) {
			project, _ := recoveryProject(t, "current-directory")
			result := recoveryProcess(t, project, []string{answer}, "", "-p", "setup", "--recover")
			require.Equal(t, 1, result.Exit, result.Output)
			require.Contains(t, result.Output, "recovery cancelled")
			require.Contains(t, result.Output, "removed or moved aside")
			data, err := os.ReadFile(filepath.Join(project, ".lenv/current/marker"))
			require.NoError(t, err)
			require.Equal(t, "keep", string(data))
		})
	}
}

func TestInteractiveRecoveryRechecksTarget(t *testing.T) {
	project, _ := recoveryProject(t, "current-file")
	result := recoveryProcess(t, project, []string{"yes"}, "replace", "-p", "setup", "--recover")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "changed while awaiting confirmation")
	data, err := os.ReadFile(filepath.Join(project, ".lenv/current"))
	require.NoError(t, err)
	require.Equal(t, "replacement", string(data))
	project, _ = recoveryProject(t, "current-directory")
	result = recoveryProcess(t, project, []string{"yes"}, "contents", "-p", "setup", "--recover")
	require.Equal(t, 0, result.Exit, result.Output)
}

func TestNoninteractiveRecoveryRejectsPipedApproval(t *testing.T) {
	project, _ := recoveryProject(t, "current-file")
	cmd := commandForLenvTest("TestRecoverySubprocess", "-p", "setup", "--recover")
	cmd.Dir = project
	cmd.Env = append(os.Environ(), "LENV_RECOVERY_PROCESS=1", `LENV_RECOVERY_ARGS=["-p","setup","--recover"]`)
	cmd.Stdin = strings.NewReader("yes\n")
	out, err := cmd.CombinedOutput()
	require.Error(t, err)
	require.Contains(t, string(out), "interactive terminal")
	require.Contains(t, string(out), "recovery cancelled")
	data, err := os.ReadFile(filepath.Join(project, ".lenv/current"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
}

func TestRecoveryCreatesMissingDirectories(t *testing.T) {
	project := t.TempDir()
	result := recoveryProcess(t, project, nil, "", "-p", "setup")
	require.Equal(t, 0, result.Exit, result.Output)
	require.NotContains(t, result.Output, "Delete and recover?")
	for _, dir := range []string{".lenv", ".lenv/src", ".lenv/lua", ".lenv/luajit"} {
		info, err := os.Lstat(filepath.Join(project, dir))
		require.NoError(t, err)
		require.True(t, info.IsDir())
	}
}

func TestRecoveryPreservesAutomaticDiscovery(t *testing.T) {
	for _, kind := range []string{"file", "link", "dangling"} {
		t.Run(kind, func(t *testing.T) {
			parent := t.TempDir()
			root := filepath.Join(parent, ".lenv")
			require.NoError(t, os.Mkdir(root, 0755))
			child := filepath.Join(parent, "child")
			require.NoError(t, os.Mkdir(child, 0755))
			candidate := filepath.Join(child, ".lenv")
			if kind == "file" {
				require.NoError(t, os.WriteFile(candidate, []byte("keep"), 0600))
			} else {
				target := root
				if kind == "dangling" {
					target = "missing"
				}
				require.NoError(t, os.Symlink(target, candidate))
			}
			before, err := os.Lstat(candidate)
			require.NoError(t, err)
			result := recoveryProcess(t, child, nil, "", "help")
			require.Equal(t, 0, result.Exit, result.Output)
			require.NotContains(t, result.Output, "Delete and recover?")
			after, err := os.Lstat(candidate)
			require.NoError(t, err)
			require.True(t, os.SameFile(before, after))
			_, err = os.Lstat(filepath.Join(root, "src"))
			require.True(t, os.IsNotExist(err))
		})
	}
}

func TestInteractiveRecoveryRemovalFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can bypass directory write permissions")
	}
	project, _ := recoveryProject(t, "current-file")
	root := filepath.Join(project, ".lenv")
	require.NoError(t, os.Chmod(root, 0555))
	defer os.Chmod(root, 0755)
	result := recoveryProcess(t, project, []string{"yes"}, "", "-p", "setup", "--recover")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "cannot remove regular file")
	require.Contains(t, result.Output, filepath.Join(root, "current"))
	data, err := os.ReadFile(filepath.Join(root, "current"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
}
