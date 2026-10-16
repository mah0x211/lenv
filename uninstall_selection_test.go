package main

import (
	"os"
	"path/filepath"
	"testing"
)

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
	project, current, rocks, modules, versionFile string
	selected                                      string
	selectionInfo                                 os.FileInfo
}

func newUninstallSelectionFixture(t *testing.T, selected string) *uninstallSelectionFixture {
	t.Helper()
	project, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(project, ".lenv")
	fixture := &uninstallSelectionFixture{
		project: project, current: filepath.Join(root, "current"),
		rocks:       filepath.Join(root, "lua/5.5.1/luarocks"),
		versionFile: filepath.Join(root, "luarocks_vers.txt"), selected: selected,
	}
	for _, name := range []string{"src", "luajit"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	versions := NewVersions()
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
			oldRocks, oldCurrent, oldExit := LuaRocksCfg, CurrentDir, ExitCode
			t.Cleanup(func() { LuaRocksCfg, CurrentDir, ExitCode = oldRocks, oldCurrent, oldExit })
			LuaRocksCfg = &TargetConfig{Name: "luarocks", RootDir: fixture.rocks, VersionFile: fixture.versionFile}
			CurrentDir, ExitCode = fixture.current, 0
			done := make(chan struct{})
			completed := false
			go func() {
				defer close(done)
				CmdUninstall([]string{":3.0.1"})
				completed = true
			}()
			<-done
			if !completed || ExitCode != 0 {
				t.Fatalf("uninstall completed = %v, exit code = %d", completed, ExitCode)
			}
			fixture.assertAfterUninstall(t)
		})
	}
}

func TestUninstallLuaRocksSelectionThroughCLI(t *testing.T) {
	for _, tc := range uninstallSelectionCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newUninstallSelectionFixture(t, tc.selected)
			code, output := startupDiagnosisProcess(t, fixture.project, "-p", "uninstall", ":3.0.1")
			if code != 0 {
				t.Fatalf("uninstall exit code = %d\n%s", code, output)
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
	code, output := startupDiagnosisProcess(t, fixture.project, "-p", "uninstall", ":3.0.1")
	if code != 0 {
		t.Fatalf("uninstall exit code = %d: %s", code, output)
	}
	fixture.assertAfterUninstall(t)
}
