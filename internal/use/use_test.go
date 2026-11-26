package use

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/env"
	"github.com/mah0x211/lenv/internal/testutil"
	version "github.com/mah0x211/lenv/internal/versions"
	"github.com/stretchr/testify/require"
)

func newUseProject(t *testing.T) string {
	t.Helper()
	testutil.PreserveWorkingDirectory(t)
	root := testutil.NewInstalledProject(t)

	wd, err := os.Getwd()
	require.NoError(t, err)
	oldRoot, oldCurrent, oldUsed, oldCWD := config.LenvDir, config.CurrentDir, config.CurrentUsed, config.CWD
	oldLua, oldJit, oldRocks, oldExit := config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg, cli.ExitCode
	t.Cleanup(func() {
		config.LenvDir, config.CurrentDir, config.CurrentUsed, config.CWD = oldRoot, oldCurrent, oldUsed, oldCWD
		config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg, cli.ExitCode = oldLua, oldJit, oldRocks, oldExit
	})
	config.LenvDir, config.CurrentDir, config.CurrentUsed, config.CWD = root, filepath.Join(root, "current"), filepath.Join(root, "lua/5.4.8"), wd
	config.LuaCfg = &config.TargetConfig{Name: "lua", RootDir: filepath.Join(root, "lua"), VersionFile: filepath.Join(root, "lua_vers.txt")}
	config.LuaJitCfg = &config.TargetConfig{Name: "luajit", RootDir: filepath.Join(root, "luajit"), VersionFile: filepath.Join(root, "luajit_vers.txt")}
	config.LuaRocksCfg = &config.TargetConfig{Name: "luarocks", RootDir: filepath.Join(root, "lua/5.4.8/luarocks"), VersionFile: filepath.Join(root, "luarocks_vers.txt")}
	cli.ExitCode = 0
	return root
}

func runUse(t *testing.T, target string) (string, bool) {
	t.Helper()
	oldRecovery := env.StartupRecovery
	t.Cleanup(func() { env.StartupRecovery = oldRecovery })
	_, stderr, completed := testutil.CaptureCommandOutput(t, func() {
		env.StartupRecovery = env.NewRecoverySession(os.Stdin, os.Stderr)
		CmdUse([]string{target})
	})
	return stderr, completed
}

func TestUseCommandSelections(t *testing.T) {
	for _, tc := range []struct{ target, runtime, rocks string }{
		{"5.5.1", "lua/5.5.1", ""},
		{"5.5", "lua/5.5.1", ""},
		{"latest", "lua/5.5.1", ""},
		{"lj-v2.1", "luajit/v2.1", ""},
		{":3.13.0", "lua/5.4.8", "3.13.0"},
		{"5.4.8:3.13", "lua/5.4.8", "3.13.0"},
	} {
		t.Run(tc.target, func(t *testing.T) {
			root := newUseProject(t)
			stderr, completed := runUse(t, tc.target)
			require.True(t, completed)
			require.Equal(t, 0, cli.ExitCode, stderr)
			link, err := os.Readlink(filepath.Join(root, "current"))
			require.NoError(t, err)
			require.Equal(t, tc.runtime, link)
			info, err := os.Lstat(filepath.Join(root, tc.runtime, "luarocks"))
			require.NoError(t, err)
			require.True(t, info.IsDir())
			if tc.rocks != "" {
				link, err = os.Readlink(filepath.Join(root, tc.runtime, "lua_modules"))
				require.NoError(t, err)
				require.Equal(t, filepath.Join("luarocks", tc.rocks, "lua_modules"), link)
			}
			testutil.AssertFileContent(t, filepath.Join(root, "lua/5.4.8/marker"), "5.4.8")
		})
	}
}

func TestLuaRocksSelectionRequiresSelectedRuntime(t *testing.T) {
	root := newUseProject(t)
	require.NoError(t, os.Remove(config.CurrentDir))
	config.CurrentUsed, config.LuaRocksCfg.RootDir = "", ""
	stderr, completed := runUse(t, ":3.13.0")
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "please run `lenv use <ver>`")
	require.DirExists(t, filepath.Join(root, "lua/5.4.8/luarocks/3.13.0"))
}

func TestLatestSelectionFailureKeepsCurrentRuntime(t *testing.T) {
	root := newUseProject(t)
	runtime := filepath.Join(root, "lua", "5.4.8")
	require.NoError(t, os.Remove(filepath.Join(runtime, "lua_modules")))
	require.NoError(t, version.NewVersions().WriteFile(config.LuaCfg.VersionFile))

	stderr, completed := runUse(t, "latest")
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "version \"latest\" does not defined")
	testutil.AssertFileContent(t, filepath.Join(runtime, "marker"), "5.4.8")
	link, err := os.Readlink(config.CurrentDir)
	require.NoError(t, err)
	require.Equal(t, "lua/5.4.8", link)
	_, err = os.Lstat(filepath.Join(runtime, "lua_modules"))
	require.True(t, os.IsNotExist(err))
}

func TestUseCommandRejectsUnavailableTargets(t *testing.T) {
	for _, state := range []string{"missing", "file", "modules directory"} {
		t.Run(state, func(t *testing.T) {
			root := newUseProject(t)
			target := "5.5.1"
			message := "is not installed"
			entry := filepath.Join(root, "lua/5.5.1")
			require.NoError(t, os.RemoveAll(entry))
			switch state {
			case "file":
				require.NoError(t, os.WriteFile(entry, []byte("keep"), 0600))
				message = "not a directory"
			case "modules directory":
				target = ":3.13.0"
				modules := filepath.Join(root, "lua/5.4.8/lua_modules")
				require.NoError(t, os.Remove(modules))
				require.NoError(t, os.Mkdir(modules, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(modules, "marker"), []byte("keep"), 0600))
				message = "failed to create symlink"
			}
			stderr, completed := runUse(t, target)
			require.False(t, completed)
			require.Equal(t, 1, cli.ExitCode, stderr)
			require.Contains(t, stderr, message)
			current, err := os.Readlink(filepath.Join(root, "current"))
			require.NoError(t, err)
			require.Equal(t, "lua/5.4.8", current)
			if state == "file" {
				testutil.AssertFileContent(t, entry, "keep")
			}
			if state == "modules directory" {
				testutil.AssertFileContent(t, filepath.Join(root, "lua/5.4.8/lua_modules/marker"), "keep")
			}
		})
	}
}

func TestUseCommandRejectsInvalidVersionMetadata(t *testing.T) {
	for _, state := range []string{"missing", "invalid", "unsupported archive"} {
		t.Run(state, func(t *testing.T) {
			root := newUseProject(t)
			file := filepath.Join(root, "lua_vers.txt")
			message := "failed to read version file"
			switch state {
			case "missing":
				require.NoError(t, os.Remove(file))
			case "invalid":
				require.NoError(t, os.WriteFile(file, []byte("{"), 0600))
			default:
				versions := version.NewVersions()
				versions.Add("lua-5.5.1.tar.xz", "5.5.1", "", ".tar.xz")
				require.NoError(t, versions.WriteFile(file))
				message = "unsupported media-type"
			}
			stderr, completed := runUse(t, "5.5.1")
			require.False(t, completed)
			require.Equal(t, 1, cli.ExitCode, stderr)
			require.Contains(t, stderr, message)
			link, err := os.Readlink(filepath.Join(root, "current"))
			require.NoError(t, err)
			require.Equal(t, "lua/5.4.8", link)
		})
	}
}

func TestUseCommandReportsUnwritableEnvironment(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := newUseProject(t)
	require.NoError(t, os.Chmod(root, 0555))
	t.Cleanup(func() { require.NoError(t, os.Chmod(root, 0755)) })
	stderr, completed := runUse(t, "5.5.1")
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "failed to tempfile")
	link, err := os.Readlink(filepath.Join(root, "current"))
	require.NoError(t, err)
	require.Equal(t, "lua/5.4.8", link)
}

func TestUseCommandReportsConflictingLuaRocksDirectory(t *testing.T) {
	root := newUseProject(t)
	rocks := filepath.Join(root, "lua/5.5.1/luarocks")
	require.NoError(t, os.WriteFile(rocks, []byte("keep"), 0600))
	stderr, completed := runUse(t, "5.5.1")
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "Recovery is disabled")
	require.Contains(t, stderr, "setup --recover")
	testutil.AssertFileContent(t, rocks, "keep")
}

func TestUseCommandRequiresVersion(t *testing.T) {
	for _, selected := range []bool{false, true} {
		state := "unselected"
		if selected {
			state = "selected"
		}
		for _, tc := range []struct {
			name string
			args []string
		}{
			{"missing", nil},
			{"empty", []string{""}},
		} {
			t.Run(state+"/"+tc.name, func(t *testing.T) {
				root := newUseProject(t)
				current := filepath.Join(root, "current")
				if !selected {
					require.NoError(t, os.Remove(current))
					config.CurrentUsed, config.LuaRocksCfg.RootDir = "", ""
				}
				before, err := os.ReadDir(root)
				require.NoError(t, err)
				output, code := testutil.CaptureHelpExit(t, func() { CmdUse(tc.args) })
				require.Equal(t, 1, code, output)
				require.Contains(t, output, "no version specified")
				require.Contains(t, output, "Usage:")
				require.NotContains(t, output, "panic:")
				testutil.AssertFileContent(t, filepath.Join(root, "lua/5.5.1/marker"), "5.5.1")
				after, err := os.ReadDir(root)
				require.NoError(t, err)
				require.Len(t, after, len(before))
				for i := range before {
					require.Equal(t, before[i].Name(), after[i].Name())
				}
				if selected {
					link, err := os.Readlink(current)
					require.NoError(t, err)
					require.Equal(t, "lua/5.4.8", link)
				} else {
					_, err := os.Lstat(current)
					require.True(t, os.IsNotExist(err))
				}
			})
		}
	}
}

func TestUseCommandReportsInaccessibleVersionDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := newUseProject(t)
	blocked := filepath.Join(root, "luajit")
	require.NoError(t, os.Chmod(blocked, 0))
	t.Cleanup(func() { require.NoError(t, os.Chmod(blocked, 0755)) })
	stderr, completed := runUse(t, "lj-v2.1")
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "failed to readdir")
	current, err := os.Readlink(filepath.Join(root, "current"))
	require.NoError(t, err)
	require.Equal(t, "lua/5.4.8", current)
	require.NoError(t, os.Chmod(blocked, 0755))
	require.DirExists(t, filepath.Join(blocked, "v2.1"))
}
