package uninstall

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/testutil"
	version "github.com/mah0x211/lenv/internal/versions"
	"github.com/stretchr/testify/require"
)

func newUninstallProject(t *testing.T) string {
	t.Helper()
	root := testutil.NewInstalledProject(t)

	oldRoot, oldCurrent, oldUsed := config.LenvDir, config.CurrentDir, config.CurrentUsed
	oldLua, oldJit, oldRocks, oldExit := config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg, cli.ExitCode
	t.Cleanup(func() {
		config.LenvDir, config.CurrentDir, config.CurrentUsed = oldRoot, oldCurrent, oldUsed
		config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg, cli.ExitCode = oldLua, oldJit, oldRocks, oldExit
	})
	config.LenvDir, config.CurrentDir, config.CurrentUsed = root, filepath.Join(root, "current"), filepath.Join(root, "lua/5.4.8")
	config.LuaCfg = &config.TargetConfig{Name: "lua", RootDir: filepath.Join(root, "lua"), VersionFile: filepath.Join(root, "lua_vers.txt")}
	config.LuaJitCfg = &config.TargetConfig{Name: "luajit", RootDir: filepath.Join(root, "luajit"), VersionFile: filepath.Join(root, "luajit_vers.txt")}
	config.LuaRocksCfg = &config.TargetConfig{Name: "luarocks", RootDir: filepath.Join(root, "lua/5.4.8/luarocks"), VersionFile: filepath.Join(root, "luarocks_vers.txt")}
	cli.ExitCode = 0
	return root
}

func runUninstallCommand(t *testing.T, target string) (string, string, bool) {
	t.Helper()
	return testutil.CaptureCommandOutput(t, func() { CmdUninstall([]string{target}) })
}

func TestUninstallCommandRemovesOnlyRequestedRuntime(t *testing.T) {
	for _, tc := range []struct{ target, dir string }{
		{"5.5.1", "lua/5.5.1"}, {"5.4.8", "lua/5.4.8"}, {"lj-v2.1", "luajit/v2.1"},
	} {
		t.Run(tc.target, func(t *testing.T) {
			root := newUninstallProject(t)
			out, stderr, completed := runUninstallCommand(t, tc.target)
			require.True(t, completed)
			require.Equal(t, 0, cli.ExitCode, stderr)
			require.Contains(t, out, "has been uninstalled")
			_, err := os.Lstat(filepath.Join(root, tc.dir))
			require.True(t, os.IsNotExist(err), "%v", err)
			for _, other := range []string{"lua/5.4.8", "lua/5.5.1", "luajit/v2.1"} {
				if other != tc.dir {
					require.DirExists(t, filepath.Join(root, other))
				}
			}
			current, err := os.Readlink(filepath.Join(root, "current"))
			require.NoError(t, err)
			require.Equal(t, "lua/5.4.8", current)
		})
	}
}

func TestUninstallCommandRejectsMissingAndNonDirectoryRuntime(t *testing.T) {
	for _, file := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "file"}[file], func(t *testing.T) {
			root := newUninstallProject(t)
			target := filepath.Join(root, "lua/5.5.1")
			require.NoError(t, os.RemoveAll(target))
			message := "is not installed"
			if file {
				require.NoError(t, os.WriteFile(target, []byte("keep"), 0600))
				message = "not a directory"
			}
			_, stderr, completed := runUninstallCommand(t, "5.5.1")
			require.False(t, completed)
			require.Equal(t, 1, cli.ExitCode, stderr)
			require.Contains(t, stderr, message)
			testutil.AssertFileContent(t, filepath.Join(root, "lua/5.4.8/marker"), "5.4.8")
			if file {
				testutil.AssertFileContent(t, target, "keep")
			}
		})
	}
}

func TestUninstallCommandReportsRemovalFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := newUninstallProject(t)
	lua := filepath.Join(root, "lua")
	version := filepath.Join(lua, "5.5.1")
	require.NoError(t, os.Chmod(version, 0555))
	require.NoError(t, os.Chmod(lua, 0555))
	t.Cleanup(func() {
		require.NoError(t, os.Chmod(lua, 0755))
		require.NoError(t, os.Chmod(version, 0755))
	})
	_, stderr, completed := runUninstallCommand(t, "5.5.1")
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "failed to uninstall version")
	testutil.AssertFileContent(t, filepath.Join(version, "marker"), "5.5.1")
	testutil.AssertFileContent(t, filepath.Join(lua, "5.4.8/marker"), "5.4.8")
}

func TestUninstallCommandRequiresExactVersion(t *testing.T) {
	root := newUninstallProject(t)
	_, stderr, completed := runUninstallCommand(t, "5.5")
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, `version "5.5" does not defined`)
	testutil.AssertFileContent(t, filepath.Join(root, "lua/5.5.1/marker"), "5.5.1")
}

func TestUninstallCommandRejectsCyclicModulesSelection(t *testing.T) {
	root := newUninstallProject(t)
	modules := filepath.Join(root, "lua/5.4.8/lua_modules")
	require.NoError(t, os.Remove(modules))
	require.NoError(t, os.Symlink("lua_modules", modules))
	_, stderr, completed := runUninstallCommand(t, ":3.12.2")
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "failed to EvalSymlink")
	require.DirExists(t, filepath.Join(root, "lua/5.4.8/luarocks/3.12.2"))
	link, err := os.Readlink(modules)
	require.NoError(t, err)
	require.Equal(t, "lua_modules", link)
}

func TestUninstallCommandRequiresVersion(t *testing.T) {
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
				root := newUninstallProject(t)
				current := filepath.Join(root, "current")
				if !selected {
					require.NoError(t, os.Remove(current))
					config.CurrentUsed, config.LuaRocksCfg.RootDir = "", ""
				}
				before, err := os.ReadDir(root)
				require.NoError(t, err)
				output, code := testutil.CaptureHelpExit(t, func() { CmdUninstall(tc.args) })
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

func TestUninstallCommandReportsInaccessibleVersionDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := newUninstallProject(t)
	blocked := filepath.Join(root, "luajit")
	require.NoError(t, os.Chmod(blocked, 0))
	t.Cleanup(func() { require.NoError(t, os.Chmod(blocked, 0755)) })
	_, stderr, completed := runUninstallCommand(t, "lj-v2.1")
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "failed to stat")
	current, err := os.Readlink(filepath.Join(root, "current"))
	require.NoError(t, err)
	require.Equal(t, "lua/5.4.8", current)
	require.NoError(t, os.Chmod(blocked, 0755))
	require.DirExists(t, filepath.Join(blocked, "v2.1"))
}

var uninstallSelectionCases = []struct {
	name, selected string
}{
	{"selected target", "3.0.1"},
	{"selected prerelease", "3.0.1-rc2"},
	{"selected longer patch", "3.0.10"},
	{"selected unrelated", "3.1.0"},
	{"no selection", ""},
}

type uninstallSelectionFixture struct {
	current, rocks, modules, versionFile string
	selected                             string
	selectionInfo                        os.FileInfo
}

func newUninstallSelectionFixture(t *testing.T, selected string) *uninstallSelectionFixture {
	t.Helper()
	project, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(project, ".lenv")
	fixture := &uninstallSelectionFixture{
		current:     filepath.Join(root, "current"),
		rocks:       filepath.Join(root, "lua/5.5.1/luarocks"),
		versionFile: filepath.Join(root, "luarocks_vers.txt"), selected: selected,
	}
	for _, name := range []string{"src", "luajit"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	versions := version.NewVersions()
	for _, version := range []string{"3.0.1", "3.0.1-rc2", "3.0.10", "3.1.0"} {
		dir := filepath.Join(fixture.rocks, version, "lua_modules")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "marker"), []byte(version), 0600); err != nil {
			t.Fatal(err)
		}
		versions.Add("luarocks-"+version+".tar.gz", version, "", ".tar.gz")
	}
	if err := versions.WriteFile(fixture.versionFile); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("lua/5.5.1", fixture.current); err != nil {
		t.Fatal(err)
	}
	fixture.modules = filepath.Join(fixture.current, "lua_modules")
	if selected != "" {
		if err := os.Symlink(filepath.Join("luarocks", selected, "lua_modules"), fixture.modules); err != nil {
			t.Fatal(err)
		}
		fixture.selectionInfo, err = os.Lstat(fixture.modules)
		if err != nil {
			t.Fatal(err)
		}
	}
	oldRocks, oldCurrent, oldExit := config.LuaRocksCfg, config.CurrentDir, cli.ExitCode
	t.Cleanup(func() { config.LuaRocksCfg, config.CurrentDir, cli.ExitCode = oldRocks, oldCurrent, oldExit })
	config.LuaRocksCfg = &config.TargetConfig{Name: "luarocks", RootDir: fixture.rocks, VersionFile: fixture.versionFile}
	config.CurrentDir, cli.ExitCode = fixture.current, 0
	return fixture
}

func (f *uninstallSelectionFixture) assertAfterUninstall(t *testing.T) {
	t.Helper()
	if _, err := os.Lstat(filepath.Join(f.rocks, "3.0.1")); !os.IsNotExist(err) {
		t.Fatalf("target version was not removed: %v", err)
	}
	for _, version := range []string{"3.0.1-rc2", "3.0.10", "3.1.0"} {
		data, err := os.ReadFile(filepath.Join(f.rocks, version, "lua_modules/marker"))
		if err != nil || string(data) != version {
			t.Fatalf("sibling %s changed: %q, %v", version, data, err)
		}
	}
	lua, err := os.Readlink(f.current)
	if err != nil || lua != "lua/5.5.1" {
		t.Fatalf("Lua selection changed: %q, %v", lua, err)
	}
	if f.selected == "" || f.selected == "3.0.1" {
		if _, err := os.Lstat(f.modules); !os.IsNotExist(err) {
			t.Fatalf("unexpected remaining selection: %v", err)
		}
		return
	}
	target, err := os.Readlink(f.modules)
	want := filepath.Join("luarocks", f.selected, "lua_modules")
	if err != nil || target != want {
		t.Fatalf("selection = %q, %v; want %q", target, err, want)
	}
	info, err := os.Lstat(f.modules)
	if err != nil || !os.SameFile(f.selectionInfo, info) {
		t.Fatalf("unrelated selection link was replaced: %v", err)
	}
}

func TestUninstallLuaRocksPreservesOtherSelection(t *testing.T) {
	for _, tc := range uninstallSelectionCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newUninstallSelectionFixture(t, tc.selected)
			done := make(chan struct{})
			completed := false
			go func() {
				defer close(done)
				CmdUninstall([]string{":3.0.1"})
				completed = true
			}()
			<-done
			if !completed || cli.ExitCode != 0 {
				t.Fatalf("uninstall completed = %v, exit code = %d", completed, cli.ExitCode)
			}
			fixture.assertAfterUninstall(t)
		})
	}
}

func TestUninstallLuaRocksSelectionAtVersionRoot(t *testing.T) {
	fixture := newUninstallSelectionFixture(t, "3.0.1")
	if err := os.Remove(fixture.modules); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("luarocks/3.0.1", fixture.modules); err != nil {
		t.Fatal(err)
	}
	_, stderr, completed := runUninstallCommand(t, ":3.0.1")
	if !completed || cli.ExitCode != 0 {
		t.Fatalf("uninstall completed = %v, exit code = %d\n%s", completed, cli.ExitCode, stderr)
	}
	fixture.assertAfterUninstall(t)
}

func TestUninstallLuaRocksReportsSelectionRemovalFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	fixture := newUninstallSelectionFixture(t, "3.0.1")
	// The runtime directory holds the link, while its luarocks subdirectory holds the version.
	runtime := filepath.Dir(fixture.rocks)
	require.NoError(t, os.Chmod(runtime, 0555))
	t.Cleanup(func() { require.NoError(t, os.Chmod(runtime, 0755)) })

	_, stderr, completed := runUninstallCommand(t, ":3.0.1")
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "failed to remove symlink")
	require.Contains(t, stderr, fixture.modules)
	_, err := os.Lstat(filepath.Join(fixture.rocks, "3.0.1"))
	require.True(t, os.IsNotExist(err), "target version was not removed: %v", err)
	testutil.AssertFileContent(t, filepath.Join(fixture.rocks, "3.0.10/lua_modules/marker"), "3.0.10")
	link, err := os.Readlink(fixture.modules)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("luarocks", "3.0.1", "lua_modules"), link)
}
