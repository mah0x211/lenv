package internal

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/testutil"
	version "github.com/mah0x211/lenv/internal/versions"
	"github.com/stretchr/testify/require"
)

// Prefer the actual instrumented CLI when coverage is collected by make test.
// Direct go test runs keep using the existing subprocess entry points.
func commandForLenvTest(helper string, args ...string) *exec.Cmd {
	if executable := os.Getenv("LENV_TEST_EXECUTABLE"); executable != "" {
		return exec.Command(executable, args...)
	}
	return exec.Command(os.Args[0], "-test.run=^"+helper+"$")
}

func newCLIProject(t *testing.T) (string, string) {
	t.Helper()
	project, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	root := filepath.Join(project, ".lenv")
	for _, name := range []string{"src", "lua/5.4.8", "lua/5.5.1", "luajit/v2.1"} {
		require.NoError(t, os.MkdirAll(filepath.Join(root, name), 0755))
	}
	lua, jit, rocks := version.NewVersions(), version.NewVersions(), version.NewVersions()
	for _, version := range []string{"5.4.8", "5.5.1"} {
		lua.Add("lua-"+version+".tar.gz", version, "", ".tar.gz")
		require.NoError(t, os.WriteFile(filepath.Join(root, "lua", version, "marker"), []byte(version), 0600))
	}
	jit.AddBranch("v2.1", "origin")
	for _, version := range []string{"3.12.2", "3.13.0"} {
		rocks.Add("luarocks-"+version+".tar.gz", version, "", ".tar.gz")
		dir := filepath.Join(root, "lua/5.4.8/luarocks", version, "lua_modules")
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "marker"), []byte(version), 0600))
	}
	for name, versions := range map[string]*version.Versions{"lua": lua, "luajit": jit, "luarocks": rocks} {
		require.NoError(t, versions.WriteFile(filepath.Join(root, name+"_vers.txt")))
	}
	require.NoError(t, os.Symlink("lua/5.4.8", filepath.Join(root, "current")))
	require.NoError(t, os.Symlink("luarocks/3.12.2/lua_modules", filepath.Join(root, "lua/5.4.8/lua_modules")))
	return project, root
}

func runCLICommand(t *testing.T, project string, args ...string) (int, string, string) {
	t.Helper()
	data, err := json.Marshal(args)
	require.NoError(t, err)
	cmd := commandForLenvTest("TestStartupDiagnosisSubprocess", args...)
	cmd.Dir = project
	cmd.Env = append(os.Environ(), "LENV_DIAG_PROCESS=1", "LENV_DIAG_ARGS="+string(data))
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	code := 0
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		require.ErrorAs(t, err, &exit, "%s", stderr.String())
		code = exit.ExitCode()
	}
	return code, stdout.String(), stderr.String()
}

func TestDirectoryOptionsAndParentEnvironmentSelection(t *testing.T) {
	project, root := newCLIProject(t)
	child := filepath.Join(project, "child", "nested")
	require.NoError(t, os.MkdirAll(child, 0755))
	code, out, stderr := runCLICommand(t, child, "path", "bin")
	require.Equal(t, 0, code, stderr)
	require.Contains(t, out, filepath.Join(root, "current/bin"))
	code, out, stderr = runCLICommand(t, project, "--global", "help")
	require.Equal(t, 0, code, stderr)
	require.Contains(t, out, "Usage:")
	require.NotContains(t, stderr, root)
}

func TestHelpWithoutDirectoryOptionLeavesProjectUntouched(t *testing.T) {
	project := t.TempDir()
	code, out, stderr := runCLICommand(t, project, "help")
	require.Equal(t, 0, code, stderr)
	require.Contains(t, out, "Usage:")
	entries, err := os.ReadDir(project)
	require.NoError(t, err)
	require.Empty(t, entries, "help must not initialize an environment")
}

func TestUnknownCommandReportsUsageWithoutChangingSelection(t *testing.T) {
	project, root := newCLIProject(t)
	code, out, stderr := runCLICommand(t, project, "-p", "not-a-command")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, out, "unknown command")
	require.Contains(t, out, "Usage:")
	link, err := os.Readlink(filepath.Join(root, "current"))
	require.NoError(t, err)
	require.Equal(t, "lua/5.4.8", link)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func installedProjectTestDirectory(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(path, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(path, "marker"), []byte("keep"), 0600))
}

func TestProjectOptionWithIfMissing(t *testing.T) {
	project, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	manager := filepath.Join(project, ".lenv")
	selected := filepath.Join(manager, "lua", "5.5.1")
	installedProjectTestDirectory(t, selected)
	require.NoError(t, os.MkdirAll(filepath.Join(manager, "src"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(manager, "luajit"), 0755))
	versions := version.NewVersions()
	versions.Add("lua-5.5.1.tar.gz", "5.5.1", "sha256:unused", ".tar.gz")
	require.NoError(t, versions.WriteFile(filepath.Join(manager, "lua_vers.txt")))
	oldArgs := os.Args
	oldClient := http.DefaultClient
	oldCWD, oldLenv, oldSrc, oldCurrent, oldUsed, oldExit := config.CWD, config.LenvDir, config.SrcDir, config.CurrentDir, config.CurrentUsed, cli.ExitCode
	oldLua, oldJit, oldRocks := config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Args, http.DefaultClient = oldArgs, oldClient
		config.CWD, config.LenvDir, config.SrcDir, config.CurrentDir, config.CurrentUsed, cli.ExitCode = oldCWD, oldLenv, oldSrc, oldCurrent, oldUsed, oldExit
		config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg = oldLua, oldJit, oldRocks
		require.NoError(t, os.Chdir(wd))
	})
	var requests int32
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&requests, 1)
		return nil, fmt.Errorf("unexpected network request to %s", req.URL)
	})}
	os.Args = []string{"lenv", "--project", "install", "--if-missing", "5.5.1"}
	require.NoError(t, os.Chdir(project))
	start()
	require.Zero(t, cli.ExitCode)
	require.Equal(t, selected, config.CurrentUsed)
	require.Zero(t, atomic.LoadInt32(&requests))
}

func TestStartupDiagnosisSubprocess(t *testing.T) {
	if os.Getenv("LENV_DIAG_PROCESS") != "1" {
		return
	}
	var args []string
	require.NoError(t, json.Unmarshal([]byte(os.Getenv("LENV_DIAG_ARGS")), &args))
	os.Args = append([]string{"lenv"}, args...)
	main()
}

func startupDiagnosisProcess(t *testing.T, project string, args ...string) (int, string) {
	t.Helper()
	data, err := json.Marshal(args)
	require.NoError(t, err)
	cmd := commandForLenvTest("TestStartupDiagnosisSubprocess", args...)
	cmd.Dir = project
	cmd.Env = append(os.Environ(), "LENV_DIAG_PROCESS=1", "LENV_DIAG_ARGS="+string(data))
	output, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		var exit *exec.ExitError
		require.ErrorAs(t, err, &exit, "%s", output)
		code = exit.ExitCode()
	}
	return code, string(output)
}

func TestStartupDiagnosisHelpAvailableForInvalidCurrent(t *testing.T) {
	for _, kind := range []string{"file", "directory", "dangling", "loop"} {
		t.Run(kind, func(t *testing.T) {
			project := t.TempDir()
			root := filepath.Join(project, ".lenv")
			for _, name := range []string{"src", "lua", "luajit"} {
				require.NoError(t, os.MkdirAll(filepath.Join(root, name), 0755))
			}
			current := filepath.Join(root, "current")
			switch kind {
			case "file":
				require.NoError(t, os.WriteFile(current, []byte("keep"), 0600))
			case "directory":
				require.NoError(t, os.Mkdir(current, 0755))
			case "dangling":
				require.NoError(t, os.Symlink("missing", current))
			case "loop":
				require.NoError(t, os.Symlink("current", current))
			}
			before, err := os.Lstat(current)
			require.NoError(t, err)
			code, output := startupDiagnosisProcess(t, project, "-p", "help")
			require.Equal(t, 0, code, output)
			require.Contains(t, output, "Usage:")
			require.Contains(t, output, "Cannot initialize lenv")
			require.NotContains(t, output, "Delete and recover?")
			after, err := os.Lstat(current)
			require.NoError(t, err)
			require.True(t, os.SameFile(before, after))
		})
	}
}

func TestStartupDiagnosisSetupReachableForInvalidCurrent(t *testing.T) {
	project := t.TempDir()
	root := filepath.Join(project, ".lenv")
	require.NoError(t, os.Mkdir(root, 0755))
	current := filepath.Join(root, "current")
	require.NoError(t, os.WriteFile(current, []byte("keep"), 0600))
	// Setup must reach directory creation despite the invalid selection. It may
	// stop later if repairing the selection requires explicit confirmation.
	_, output := startupDiagnosisProcess(t, project, "-p", "setup")
	for _, name := range []string{"src", "lua", "luajit"} {
		info, err := os.Lstat(filepath.Join(root, name))
		require.NoError(t, err, "%s", output)
		require.True(t, info.IsDir())
	}
	data, err := os.ReadFile(current)
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
	require.NotContains(t, output, "Delete and recover?")
}

func TestStartupDiagnosisBlocksCommandsWithoutInitializing(t *testing.T) {
	project := t.TempDir()
	code, output := startupDiagnosisProcess(t, project, "-p", "current")
	require.Equal(t, 1, code, output)
	require.Contains(t, output, "Run lenv -p setup")
	_, err := os.Lstat(filepath.Join(project, ".lenv"))
	require.True(t, os.IsNotExist(err))
	code, output = startupDiagnosisProcess(t, project, "-p", "help")
	require.Equal(t, 0, code, output)
	require.Contains(t, output, "Usage:")
	_, err = os.Lstat(filepath.Join(project, ".lenv"))
	require.True(t, os.IsNotExist(err))
}

func TestStartupOutputForAllCommands(t *testing.T) {
	for _, command := range []struct {
		name string
		args []string
	}{
		{"help", []string{"help"}},
		{"default-help", nil},
		{"current", []string{"current"}},
		{"list", []string{"ls"}},
		{"path", []string{"path"}},
		{"path-bin", []string{"path", "bin"}},
		{"setup", []string{"setup"}},
		{"setup-recover", []string{"setup", "--recover"}},
		{"vers", []string{"vers"}},
	} {
		for _, healthy := range []bool{false, true} {
			name := command.name + "/missing"
			if healthy {
				name = command.name + "/healthy"
			}
			t.Run(name, func(t *testing.T) {
				project := t.TempDir()
				if healthy {
					for _, dir := range []string{"src", "lua", "luajit"} {
						require.NoError(t, os.MkdirAll(filepath.Join(project, ".lenv", dir), 0755))
					}
					for _, name := range []string{"lua", "luajit", "luarocks"} {
						require.NoError(t, version.NewVersions().WriteFile(filepath.Join(project, ".lenv", name+"_vers.txt")))
					}
				}
				args := append([]string{"-p"}, command.args...)
				code, stdout, stderr := runCLICommand(t, project, args...)
				message := "Checking required directories and current selection..."
				setup := strings.HasPrefix(command.name, "setup")
				help := command.name == "help" || command.name == "default-help"
				require.NotContains(t, stdout, message)
				if setup {
					require.Equal(t, 1, strings.Count(stderr, message), stderr)
				} else {
					require.NotContains(t, stderr, message)
					if healthy {
						require.Empty(t, stderr)
					} else {
						require.Contains(t, stderr, "Error: Cannot initialize lenv:")
						require.Contains(t, stderr, "Run lenv -p setup")
						require.NotContains(t, stdout, "Error:")
					}
				}
				if !healthy && !setup && !help {
					require.Equal(t, 1, code, stderr)
					require.Empty(t, stdout)
				} else {
					require.Equal(t, 0, code, stderr)
				}
				if help {
					require.Contains(t, stdout, "Usage:")
				}
				if healthy && command.name == "path" {
					require.Contains(t, stdout, "export PATH=")
				}
			})
		}
	}
}

func TestCommandsRequireVersion(t *testing.T) {
	for _, args := range [][]string{{"-p", "use"}, {"-p", "uninstall", ""}} {
		project, _ := newCLIProject(t)
		code, output := startupDiagnosisProcess(t, project, args...)
		require.Equal(t, 1, code, output)
		require.Contains(t, output, "no version specified")
		require.Contains(t, output, "Usage:")
		require.NotContains(t, output, "panic:")
	}
}

func TestFetchExitSubprocess(t *testing.T) {
	if os.Getenv("LENV_FETCH_EXIT_CHILD") != "1" {
		return
	}
	pool := x509.NewCertPool()
	require.True(t, pool.AppendCertsFromPEM([]byte(os.Getenv("LENV_FETCH_EXIT_CERT"))))
	http.DefaultClient = &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12},
	}}
	address := os.Getenv("LENV_FETCH_EXIT_URL")
	config.LuaCfg.ReleaseURL = address + "/lua"
	config.LuaJitCfg.ReleaseURL = address + "/jit"
	config.LuaRocksCfg.ReleaseURL = address + "/rocks"
	os.Args = []string{"lenv", "--project", "fetch"}
	main()
}

func TestFetchExitStatus(t *testing.T) {
	for _, failure := range []string{"", "lua"} {
		name := failure
		if name == "" {
			name = "success"
		}
		t.Run(name, func(t *testing.T) {
			project := t.TempDir()
			root := filepath.Join(project, ".lenv")
			for _, dir := range []string{"src", "lua", "luajit"} {
				require.NoError(t, os.MkdirAll(filepath.Join(root, dir), 0755))
			}
			server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/"+failure {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				switch r.URL.Path {
				case "/lua":
					fmt.Fprint(w, "<tr><td class=\"name\"><a href=\"lua-5.5.1.tar.gz\">lua</a></td><td class=\"sum\">aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa</td></tr>")
				case "/jit":
					fmt.Fprint(w, `[{"name":"v2.1"}]`)
				case "/rocks":
					fmt.Fprint(w, `<a href="luarocks-3.13.0.tar.gz">archive</a><a href="luarocks-3.13.0.tar.gz.asc">signature</a>`)
				}
			}))
			certificate := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
			cmd := exec.Command(os.Args[0], "-test.run=^TestFetchExitSubprocess$")
			cmd.Dir = project
			cmd.Env = append(os.Environ(), "LENV_FETCH_EXIT_CHILD=1", "LENV_FETCH_EXIT_URL="+server.URL, "LENV_FETCH_EXIT_CERT="+string(certificate))
			output, err := cmd.CombinedOutput()
			if failure == "" {
				require.NoError(t, err, "%s", output)
			} else {
				var exit *exec.ExitError
				require.ErrorAs(t, err, &exit, "%s", output)
				require.Equal(t, 1, exit.ExitCode(), "%s", output)
				require.Contains(t, string(output), "503 Service Unavailable")
			}
			require.NotContains(t, string(output), "panic:")
		})
	}
}

func TestLatestSelectionCLIError(t *testing.T) {
	for _, command := range []string{"install", "use"} {
		t.Run(command, func(t *testing.T) {
			project, root := newCLIProject(t)
			require.NoError(t, version.NewVersions().WriteFile(filepath.Join(root, "lua_vers.txt")))
			code, output := startupDiagnosisProcess(t, project, "--project", command, "latest")
			require.Equal(t, 1, code, output)
			require.Contains(t, output, "version \"latest\" does not defined")
			require.Contains(t, output, "lua_vers.txt")
			require.NotContains(t, output, "panic:")
		})
	}
}

func TestPathWithoutLua(t *testing.T) {
	project, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	for _, name := range []string{"src", "lua", "luajit"} {
		require.NoError(t, os.MkdirAll(filepath.Join(project, ".lenv", name), 0755))
	}
	current := filepath.Join(project, ".lenv", "current")
	t.Setenv("PATH", t.TempDir())
	t.Setenv("LUA_INIT", "error('must not run')")
	code, output, stderr := runCLICommand(t, project, "-p", "path")
	require.Equal(t, 0, code, stderr)
	require.Contains(t, output, filepath.Join(current, "bin"))
	// Generating settings must not install or select a runtime.
	_, err = os.Lstat(current)
	require.True(t, os.IsNotExist(err))
	for _, name := range []string{"lua", "luajit"} {
		entries, err := os.ReadDir(filepath.Join(project, ".lenv", name))
		require.NoError(t, err)
		require.Empty(t, entries)
	}
}

func sourcePathSettings(t *testing.T, shell, file, inherited string) []string {
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

func TestPathShellSettingsFromProjectCommand(t *testing.T) {
	base := t.TempDir()
	project := filepath.Join(base, "project ' \"$HOME $(printf changed) `printf changed`")
	for _, name := range []string{"src", "lua", "luajit"} {
		if err := os.MkdirAll(filepath.Join(project, ".lenv", name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	// main discovers the physical working directory, including on macOS where
	// temporary directory names may start with the /var alias.
	project, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatal(err)
	}
	code, settings, stderr := runCLICommand(t, project, "-p", "path")
	require.Equal(t, 0, code, stderr)
	rc := filepath.Join(base, "lenvrc")
	if err := os.WriteFile(rc, []byte(settings), 0600); err != nil {
		t.Fatal(err)
	}
	for _, shell := range []string{"/bin/sh", "/bin/bash"} {
		values := sourcePathSettings(t, shell, rc, "/inherited/bin")
		current := filepath.Join(project, ".lenv", "current")
		want := filepath.Join(current, "bin") + ":" + filepath.Join(current, "lua_modules/bin") + ":/inherited/bin"
		if values[0] != want {
			t.Fatalf("%s project PATH = %q; want %q", shell, values[0], want)
		}
		for i, suffix := range []string{"lua_modules/lualib/", "lua_modules/luaclib/"} {
			if !strings.HasPrefix(values[i+1], filepath.Join(current, suffix)+string(filepath.Separator)) {
				t.Fatalf("project Lua setting %q lacks expected directory", values[i+1])
			}
			if !strings.HasSuffix(values[i+1], ";;") {
				t.Fatalf("project Lua setting %q lacks default search path", values[i+1])
			}
		}
	}
}

func TestRecoveryMissingRocksRequiresSetupAtStartup(t *testing.T) {
	project := recoveryProject(t, "rocks-file")
	rocks := filepath.Join(project, ".lenv/lua/5.5.1/luarocks")
	require.NoError(t, os.Remove(rocks))
	result := recoveryProcess(t, project, nil, "-p", "current")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "Run lenv -p setup to")
	_, err := os.Lstat(rocks)
	require.True(t, os.IsNotExist(err))
	result = recoveryProcess(t, project, nil, "-p", "setup")
	require.Equal(t, 0, result.Exit, result.Output)
	require.NotContains(t, result.Output, "Delete and recover?")
	info, err := os.Lstat(rocks)
	require.NoError(t, err)
	require.True(t, info.IsDir())
	result = recoveryProcess(t, project, nil, "-p", "current")
	require.Equal(t, 0, result.Exit, result.Output)
	require.NotContains(t, result.Output, "Run lenv")
}

func TestRecoveryDisabledByDefault(t *testing.T) {
	project := recoveryProject(t, "current-directory")
	path := filepath.Join(project, ".lenv", "current")
	before, err := os.Lstat(path)
	require.NoError(t, err)
	result := recoveryProcess(t, project, []string{"yes"}, "-p", "current")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "Run lenv -p setup --recover")
	require.Contains(t, result.Output, fmt.Sprintf("Path: %q", path))
	require.NotContains(t, result.Output, "Delete and recover?")
	after, err := os.Lstat(path)
	require.NoError(t, err)
	require.True(t, os.SameFile(before, after))
}

func TestSetupRecoveryFlagWithProjectOption(t *testing.T) {
	for _, option := range []string{"-p", "--project"} {
		project := recoveryProject(t, "current-file")
		result := recoveryProcess(t, project, []string{"yes"}, option, "setup", "--recover")
		require.Equal(t, 0, result.Exit, result.Output)
		require.Contains(t, result.Output, "Delete and recover?")
		_, err := os.Lstat(filepath.Join(project, ".lenv/current"))
		require.True(t, os.IsNotExist(err))
	}
}

func TestRecoveryFlagIsSetupOnly(t *testing.T) {
	for _, args := range [][]string{{"-p", "--recover", "setup"}, {"--recover", "setup"}} {
		project := t.TempDir()
		// Automatic discovery uses this empty real directory for the second case.
		require.NoError(t, os.Mkdir(filepath.Join(project, ".lenv"), 0755))
		result := recoveryProcess(t, project, []string{"yes"}, args...)
		require.Equal(t, 1, result.Exit, result.Output)
		require.NotContains(t, result.Output, "Delete and recover?")
		_, err := os.Lstat(filepath.Join(project, ".lenv/src"))
		require.True(t, os.IsNotExist(err))
	}
}

func TestRecoveryDirectoryOptionConflict(t *testing.T) {
	for _, args := range [][]string{{"--global", "--project", "setup", "--recover"}, {"-g", "-p", "setup"}} {
		project := t.TempDir()
		result := recoveryProcess(t, project, nil, args...)
		require.Equal(t, 1, result.Exit, result.Output)
		require.Contains(t, result.Output, "specify only one installation directory option")
		require.NotContains(t, result.Output, "Delete and recover?")
	}
}

func TestRecoveryHelpDoesNotChangeState(t *testing.T) {
	for _, kind := range []string{"current-file", "root-file"} {
		project := recoveryProject(t, kind)
		result := recoveryProcess(t, project, []string{"yes"}, "-p", "help")
		require.Equal(t, 0, result.Exit, result.Output)
		require.Contains(t, result.Output, "Usage:")
		require.Contains(t, result.Output, "Run lenv -p setup --recover")
		require.NotContains(t, result.Output, "Delete and recover?")
	}
	project := t.TempDir()
	result := recoveryProcess(t, project, nil, "-p", "help")
	require.Equal(t, 0, result.Exit, result.Output)
	require.Contains(t, result.Output, "Run lenv -p setup to")
	_, err := os.Lstat(filepath.Join(project, ".lenv"))
	require.True(t, os.IsNotExist(err))
}

func TestRecoveryBlocksOtherCommandsWithoutChanges(t *testing.T) {
	for _, command := range []string{"path", "fetch", "vers", "ls", "install", "uninstall", "use", "current"} {
		t.Run(command, func(t *testing.T) {
			project := t.TempDir()
			result := recoveryProcess(t, project, nil, "-p", command)
			require.Equal(t, 1, result.Exit, result.Output)
			require.Contains(t, result.Output, "Run lenv -p setup to")
			require.NotContains(t, result.Output, "Delete and recover?")
			_, err := os.Lstat(filepath.Join(project, ".lenv"))
			require.True(t, os.IsNotExist(err))
		})
	}
}

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

func recoveryProcess(t *testing.T, project string, answers []string, args ...string) testutil.PTYResult {
	t.Helper()
	data, err := json.Marshal(args)
	require.NoError(t, err)
	cmd := commandForLenvTest("TestRecoverySubprocess", args...)
	if executable := os.Getenv("LENV_RECOVERY_EXECUTABLE"); os.Getenv("LENV_TEST_EXECUTABLE") == "" && executable != "" {
		cmd = exec.Command(executable, args...)
	}
	cmd.Dir = project
	cmd.Env = append(os.Environ(), "LENV_RECOVERY_PROCESS=1", "LENV_RECOVERY_ARGS="+string(data))
	return testutil.RunPTY(t, cmd, "Delete and recover? [y/N] ", answers, "", "")
}

func recoveryProject(t *testing.T, kind string) string {
	t.Helper()
	project, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	root := filepath.Join(project, ".lenv")
	if kind == "root-file" {
		require.NoError(t, os.WriteFile(root, nil, 0600))
		return project
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
	case "rocks-file":
		require.NoError(t, os.Symlink("lua/5.5.1", current))
		require.NoError(t, os.WriteFile(filepath.Join(root, "lua/5.5.1/luarocks"), []byte("keep"), 0600))
	case "healthy":
	default:
		t.Fatalf("unknown recovery case: %s", kind)
	}
	return project
}

func TestRecoveryHelpWithInspectionError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can bypass directory permissions")
	}
	project := recoveryProject(t, "healthy")
	locked := filepath.Join(project, ".lenv/locked")
	require.NoError(t, os.MkdirAll(filepath.Join(locked, "lua"), 0755))
	require.NoError(t, os.Symlink("locked/lua", filepath.Join(project, ".lenv/current")))
	require.NoError(t, os.Chmod(locked, 0000))
	defer os.Chmod(locked, 0755)
	result := recoveryProcess(t, project, nil, "-p", "help")
	require.Equal(t, 0, result.Exit, result.Output)
	require.Contains(t, result.Output, "cannot resolve current selection")
	require.Contains(t, result.Output, "Usage:")
	require.NotContains(t, result.Output, "Delete and recover?")
}

func TestSetupOutputForMissingDirectories(t *testing.T) {
	project := t.TempDir()
	result := recoveryProcess(t, project, nil, "-p", "setup")
	require.Equal(t, 0, result.Exit, result.Output)
	check := strings.Index(result.Output, "Checking required directories and current selection...")
	creating := strings.Index(result.Output, "Creating required directory")
	created := strings.Index(result.Output, "Created required directory")
	complete := strings.Index(result.Output, "The required directories are ready.")
	require.True(t, check >= 0 && creating > check && created > creating && complete > created, result.Output)
	require.NotContains(t, result.Output, "Error:")
}

func TestSetupOutputForConflictWithoutRecovery(t *testing.T) {
	project := recoveryProject(t, "current-file")
	result := recoveryProcess(t, project, nil, "-p", "setup")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "Checking required directories and current selection...")
	require.Contains(t, result.Output, "Error: Cannot initialize lenv:")
	require.Contains(t, result.Output, "Nothing was deleted at this path.")
	require.Contains(t, result.Output, "Run lenv -p setup --recover from an interactive terminal")
	require.Contains(t, result.Output, "move the conflicting entry aside and run lenv -p setup again")
	require.NotContains(t, result.Output, "Delete and recover?")
	require.NotContains(t, result.Output, "Removing regular file")
	require.NotContains(t, result.Output, "The required directories are ready.")
	data, err := os.ReadFile(filepath.Join(project, ".lenv/current"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
}

func TestUninstallCommandRemovesOnlyRequestedRuntime(t *testing.T) {
	project, root := newCLIProject(t)
	code, out, stderr := runCLICommand(t, project, "-p", "uninstall", "5.4.8")
	require.Equal(t, 0, code, stderr)
	require.Contains(t, out, "has been uninstalled")
	_, err := os.Lstat(filepath.Join(root, "lua/5.4.8"))
	require.True(t, os.IsNotExist(err), "%v", err)
	current, err := os.Readlink(filepath.Join(root, "current"))
	require.NoError(t, err)
	require.Equal(t, "lua/5.4.8", current)
	// Removing a selected runtime leaves its selection for setup --recover.
	code, _, stderr = runCLICommand(t, project, "-p", "current")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, stderr, "setup")
}

func TestUninstallLuaRocksSelectionThroughCLI(t *testing.T) {
	project, root := newCLIProject(t)
	lua := filepath.Join(root, "lua", "5.4.8")
	modules := filepath.Join(lua, "lua_modules")
	before, err := os.Lstat(modules)
	require.NoError(t, err)
	code, _, stderr := runCLICommand(t, project, "-p", "uninstall", ":3.13.0")
	require.Equal(t, 0, code, stderr)
	_, err = os.Lstat(filepath.Join(lua, "luarocks", "3.13.0"))
	require.True(t, os.IsNotExist(err))
	testutil.AssertFileContent(t, filepath.Join(lua, "luarocks", "3.12.2", "lua_modules", "marker"), "3.12.2")
	selection, err := os.Readlink(filepath.Join(root, "current"))
	require.NoError(t, err)
	require.Equal(t, "lua/5.4.8", selection)
	link, err := os.Readlink(modules)
	require.NoError(t, err)
	require.Equal(t, "luarocks/3.12.2/lua_modules", link)
	after, err := os.Lstat(modules)
	require.NoError(t, err)
	require.True(t, os.SameFile(before, after))
}
