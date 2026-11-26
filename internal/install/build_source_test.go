package install

import (
	"archive/tar"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/env"
	"github.com/mah0x211/lenv/internal/testutil"
	versionmodel "github.com/mah0x211/lenv/internal/versions"
	"github.com/stretchr/testify/require"
)

func TestBuildSourcePreparationReportsFilesystemFailures(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	for _, stage := range []string{"create parent", "inspect parent", "inspect source", "create temporary tree", "remove source"} {
		t.Run(stage, func(t *testing.T) {
			fixture := newInstallFetchFixture(t, "")
			build := filepath.Join(config.LenvDir, "build")
			parent := filepath.Join(build, "lua")
			source := filepath.Join(parent, "5.5.1")
			require.NoError(t, os.MkdirAll(parent, 0755))
			blocked, mode := parent, os.FileMode(0555)
			switch stage {
			case "create parent":
				require.NoError(t, os.Remove(parent))
				blocked = build
			case "inspect parent":
				blocked = build
				mode = 0
			case "inspect source":
				mode = 0
			case "remove source":
				require.NoError(t, os.Mkdir(source, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(source, "marker"), []byte("keep"), 0600))
				blocked = source
			}
			require.NoError(t, os.Chmod(blocked, mode))
			t.Cleanup(func() { require.NoError(t, os.Chmod(blocked, 0755)) })
			result, err := prepareBuildSource(config.LuaCfg, &versionmodel.VerItem{Ver: "5.5.1"}, &options{NoRebuild: true})
			require.Error(t, err)
			require.Empty(t, result)
			require.Zero(t, fixture.requests, "a local preparation failure must stop before downloading")
			require.NoError(t, os.Chmod(blocked, 0755))
			if stage == "remove source" {
				testutil.AssertFileContent(t, filepath.Join(source, "marker"), "keep")
			}
		})
	}
}

func TestCachedArchiveReadFailureIsReported(t *testing.T) {
	oldSrc := config.SrcDir
	config.SrcDir = t.TempDir()
	t.Cleanup(func() { config.SrcDir = oldSrc })
	require.NoError(t, os.Mkdir(filepath.Join(config.SrcDir, "lua.tar.gz"), 0755))
	source, err := extractCachedFile(t.TempDir(), "https://example.invalid/lua.tar.gz", &config.TargetConfig{Name: "lua"}, &versionmodel.VerItem{}, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "open cached file error")
	require.Empty(t, source)
	require.DirExists(t, filepath.Join(config.SrcDir, "lua.tar.gz"))
}

func TestRebuildOptionConflictStopsBeforeFetch(t *testing.T) {
	for _, flags := range [][]string{{"--rebuild", "--no-rebuild"}, {"--no-rebuild", "--rebuild"}} {
		fixture := newInstallFetchFixture(t, "")
		args := append([]string{"--fetch"}, flags...)
		completed, message := captureInstallError(t, append(args, "5.5.1")...)
		require.False(t, completed)
		require.Contains(t, message, "cannot be combined")
		require.Zero(t, atomic.LoadInt32(&fixture.requests))
	}
}

func TestBuildSourcePaths(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	oldRoot := config.LenvDir
	config.LenvDir = root
	t.Cleanup(func() { config.LenvDir = oldRoot })
	for _, target := range []string{"lua", "luajit"} {
		path, err := buildSourcePath(&config.TargetConfig{Name: target}, &versionmodel.VerItem{Ver: "5.5.1"})
		require.NoError(t, err)
		require.Equal(t, filepath.Join(root, "build", target, "5.5.1"), path)
	}
	for _, runtime := range []string{"lua/5.5.1", "luajit/v2.1"} {
		dir := filepath.Join(root, filepath.FromSlash(runtime))
		require.NoError(t, os.MkdirAll(dir, 0755))
		cfg := &config.TargetConfig{Name: "luarocks", RootDir: filepath.Join(dir, "luarocks")}
		path, err := buildSourcePath(cfg, &versionmodel.VerItem{Ver: "3.13.0"})
		require.NoError(t, err)
		require.Equal(t, filepath.Join(root, "build", "luarocks", filepath.FromSlash(runtime), "3.13.0"), path)
	}
	external := t.TempDir()
	cfg := &config.TargetConfig{Name: "luarocks", RootDir: filepath.Join(external, "luarocks")}
	first, err := buildSourcePath(cfg, &versionmodel.VerItem{Ver: "3.13.0"})
	require.NoError(t, err)
	require.Contains(t, first, filepath.Join("build", "luarocks", "external"))
	cfg.RootDir = filepath.Join(t.TempDir(), "luarocks")
	second, err := buildSourcePath(cfg, &versionmodel.VerItem{Ver: "3.13.0"})
	require.NoError(t, err)
	require.NotEqual(t, first, second)
	lua := &config.TargetConfig{Name: "lua"}
	for _, version := range []string{"", ".", "../outside", "/outside", "a/b", "a\\b"} {
		_, err := buildSourcePath(lua, &versionmodel.VerItem{Ver: version})
		require.Error(t, err)
	}
	_, err = buildSourcePath(&config.TargetConfig{Name: "unknown"}, &versionmodel.VerItem{Ver: "1.0"})
	require.Error(t, err)
}

func TestBuildSourceRejectsConflicts(t *testing.T) {
	for _, conflict := range []string{"parent link", "parent file", "source link", "source file"} {
		t.Run(conflict, func(t *testing.T) {
			newInstallFetchFixture(t, "")
			item := &versionmodel.VerItem{Ver: "5.5.1"}
			path, err := buildSourcePath(config.LuaCfg, item)
			require.NoError(t, err)
			outside := t.TempDir()
			marker := filepath.Join(outside, "marker")
			require.NoError(t, os.WriteFile(marker, []byte("keep"), 0600))
			at := filepath.Join(config.LenvDir, "build")
			if strings.HasPrefix(conflict, "source") {
				at = path
				require.NoError(t, os.MkdirAll(filepath.Dir(at), 0755))
			}
			if strings.HasSuffix(conflict, "link") {
				require.NoError(t, os.Symlink(outside, at))
			} else {
				require.NoError(t, os.WriteFile(at, []byte("keep"), 0600))
			}
			for _, opts := range []*options{{Rebuild: true}, {NoRebuild: true}} {
				_, err := prepareBuildSource(config.LuaCfg, item, opts)
				require.Error(t, err)
			}
			testutil.AssertFileContent(t, marker, "keep")
			info, err := os.Lstat(at)
			require.NoError(t, err)
			if strings.HasSuffix(conflict, "link") {
				require.NotZero(t, info.Mode()&os.ModeSymlink)
			} else {
				testutil.AssertFileContent(t, at, "keep")
			}
		})
	}
}

func TestBuildSourceRocksReuseWithoutVerifier(t *testing.T) {
	newInstallFetchFixture(t, "")
	cfg := &config.TargetConfig{Name: "luarocks", RootDir: filepath.Join(config.CurrentUsed, "luarocks")}
	item := &versionmodel.VerItem{Ver: "3.13.0"}
	path, err := buildSourcePath(cfg, item)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(path, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(path, "edited"), []byte("debug"), 0600))
	t.Setenv("PATH", t.TempDir())
	got, err := prepareBuildSource(cfg, item, &options{Rebuild: true})
	require.NoError(t, err)
	require.Equal(t, path, got)
	testutil.AssertFileContent(t, filepath.Join(path, "edited"), "debug")
}

func TestBuildSourceRejectsUnsafeExtraction(t *testing.T) {
	newInstallFetchFixture(t, "")
	data := testutil.ArchiveData(t, testutil.ArchiveEntry{Name: "../outside", Kind: tar.TypeReg, Body: "rejected"})
	item := &versionmodel.VerItem{Name: "lua.tar.gz", Ver: "5.5.1", Sum: testutil.ArchiveSHA256(data)}
	require.NoError(t, os.WriteFile(filepath.Join(config.SrcDir, item.Name), data, 0600))
	config.LuaCfg.DownloadURL = "http://example.invalid/"
	path, err := buildSourcePath(config.LuaCfg, item)
	require.NoError(t, err)
	_, err = prepareBuildSource(config.LuaCfg, item, &options{Rebuild: true})
	require.Error(t, err)
	_, err = os.Lstat(path)
	require.True(t, os.IsNotExist(err))
	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestVerifiedBuildSourceLuaRocks(t *testing.T) {
	newInstallFetchFixture(t, "")
	data, signature := testutil.OfficialLuaRocksArchive(t)
	cfg := &config.TargetConfig{Name: "luarocks", RootDir: filepath.Join(config.CurrentUsed, "luarocks"), SigningKeyURL: testutil.PublisherKeyURL(t)}
	item := &versionmodel.VerItem{Name: "luarocks-3.13.0.tar.gz", Ver: "3.13.0", Sum: "pgp:luarocks-3.13.0.tar.gz.asc"}
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".asc") {
			w.Write(signature)
		} else {
			w.Write(data)
		}
	}))
	cfg.DownloadURL = server.URL + "/"
	path, err := prepareBuildSource(cfg, item, &options{Rebuild: true})
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(path, "configure"))
	require.NoError(t, err)
	server.Close()
	// Reuse does not fetch the signing key or archive after sources are edited.
	cfg.SigningKeyURL = "http://example.invalid/"
	got, err := prepareBuildSource(cfg, item, &options{Rebuild: true})
	require.NoError(t, err)
	require.Equal(t, path, got)
}

func TestBuildSourceRejectsInvalidEnvironmentPaths(t *testing.T) {
	newInstallFetchFixture(t, "")
	_, err := prepareBuildSource(config.LuaCfg, &versionmodel.VerItem{Ver: "../outside"}, &options{Rebuild: true})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid build source version")
	err = ensureBuildParent(filepath.Dir(config.LenvDir))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid build directory")
	cfg := &config.TargetConfig{Name: "luarocks", RootDir: filepath.Join(config.LenvDir, "lua/missing/luarocks")}
	_, err = prepareBuildSource(cfg, &versionmodel.VerItem{Ver: "3.13.0"}, &options{Rebuild: true})
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot resolve Lua environment")

	cfg.RootDir = filepath.Join(config.CurrentUsed, "luarocks")
	config.LenvDir = filepath.Join(config.LenvDir, "missing")
	_, err = prepareBuildSource(cfg, &versionmodel.VerItem{Ver: "3.13.0"}, &options{Rebuild: true})
	require.ErrorIs(t, err, os.ErrNotExist)
	var pathErr *os.PathError
	require.ErrorAs(t, err, &pathErr)
	require.Equal(t, config.LenvDir, pathErr.Path)
}

func TestBuildSourcePublicationKeepsNewlyAppearedEntry(t *testing.T) {
	newInstallFetchFixture(t, "")
	data := rebuildLuaArchive(t, false)
	item := &versionmodel.VerItem{Name: "lua-5.5.1.tar.gz", Ver: "5.5.1", Sum: testutil.ArchiveSHA256(data)}
	path, err := buildSourcePath(config.LuaCfg, item)
	require.NoError(t, err)
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Another operation creates the destination while this acquisition is running.
		if err := os.Mkdir(path, 0755); err != nil {
			t.Error(err)
			w.WriteHeader(500)
			return
		}
		if err := os.WriteFile(filepath.Join(path, "marker"), []byte("new owner"), 0600); err != nil {
			t.Error(err)
			w.WriteHeader(500)
			return
		}
		w.Write(data)
	}))
	config.LuaCfg.DownloadURL = server.URL + "/"
	_, err = prepareBuildSource(config.LuaCfg, item, &options{Rebuild: true})
	require.Error(t, err)
	require.Contains(t, err.Error(), "appeared during extraction")
	testutil.AssertFileContent(t, filepath.Join(path, "marker"), "new owner")
	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)
	require.Len(t, entries, 1)
}

func TestInstallPromptChild(t *testing.T) {
	opts, ok := testutil.TestChildArgs(t)
	if !ok {
		return
	}
	config.InitGlobalVars("project")
	env.DiagnoseStartup().SelectCurrent()
	done := make(chan struct{})
	go func() {
		defer close(done)
		CmdInstall(opts)
	}()
	<-done
	os.Exit(cli.ExitCode)
}

func buildSourceProcess(t *testing.T, project string, answers []string, mutation string, opts ...string) testutil.PTYResult {
	t.Helper()
	cmd := testutil.TestChild(t, "TestInstallPromptChild", project, opts...)
	source := filepath.Join(project, ".lenv", "build", "lua", "5.5.1")
	return testutil.RunPTY(t, cmd, "Delete existing build sources? [y/N] ", answers, source, mutation)
}

func buildSourceProject(t *testing.T) (string, string, string) {
	t.Helper()
	project, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	root := filepath.Join(project, ".lenv")
	source := filepath.Join(root, "build/lua/5.5.1")
	installed := filepath.Join(root, "lua/5.5.1")
	for _, dir := range []string{source, installed, filepath.Join(root, "src"), filepath.Join(root, "luajit"), filepath.Join(installed, "luarocks")} {
		require.NoError(t, os.MkdirAll(dir, 0755))
	}
	require.NoError(t, os.Symlink("lua/5.5.1", filepath.Join(root, "current")))
	require.NoError(t, os.WriteFile(filepath.Join(installed, "marker"), []byte("installed"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(source, "input"), []byte("edited"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(source, "Makefile"), []byte(rebuildMakefile), 0600))
	data := rebuildLuaArchive(t, false)
	require.NoError(t, os.WriteFile(filepath.Join(root, "src/lua-5.5.1.tar.gz"), data, 0600))
	versions := versionmodel.NewVersions()
	versions.Add("lua-5.5.1.tar.gz", "5.5.1", testutil.ArchiveSHA256(data), ".tar.gz")
	require.NoError(t, versions.WriteFile(filepath.Join(root, "lua_vers.txt")))
	return project, source, installed
}

func TestBuildSourcePromptCancellation(t *testing.T) {
	for _, answer := range []string{"n", "", "EOF", "not-yes"} {
		t.Run(answer, func(t *testing.T) {
			project, source, installed := buildSourceProject(t)
			result := buildSourceProcess(t, project, []string{answer}, "", "5.5.1")
			require.Equal(t, 1, result.Exit, result.Output)
			require.Contains(t, result.Output, source)
			require.Contains(t, result.Output, "ALL source changes and build output")
			require.Contains(t, result.Output, "source replacement cancelled")
			require.Contains(t, result.Output, "--rebuild")
			require.Contains(t, result.Output, "--no-rebuild")
			testutil.AssertFileContent(t, filepath.Join(source, "input"), "edited")
			testutil.AssertFileContent(t, filepath.Join(installed, "marker"), "installed")
		})
	}
}

func TestBuildSourcePromptApprovalAndExplicitModes(t *testing.T) {
	for _, mode := range []string{"yes", "y", "--rebuild", "--no-rebuild"} {
		t.Run(mode, func(t *testing.T) {
			project, source, installed := buildSourceProject(t)
			var args []string
			var answers []string
			if strings.HasPrefix(mode, "--") {
				args = append(args, mode)
			} else {
				answers = []string{mode}
			}
			args = append(args, "5.5.1")
			result := buildSourceProcess(t, project, answers, "", args...)
			require.Equal(t, 0, result.Exit, result.Output)
			want := "upstream"
			if mode == "--rebuild" {
				want = "edited"
				testutil.AssertFileContent(t, filepath.Join(source, "input"), want)
			} else {
				_, err := os.Lstat(source)
				require.True(t, os.IsNotExist(err))
			}
			testutil.AssertFileContent(t, filepath.Join(installed, "marker"), want)
			if strings.HasPrefix(mode, "--") {
				require.NotContains(t, result.Output, "Delete existing build sources?")
			}
		})
	}
}

func TestBuildSourcePromptRechecksTarget(t *testing.T) {
	for _, mutation := range []string{"replace", "contents"} {
		t.Run(mutation, func(t *testing.T) {
			project, source, installed := buildSourceProject(t)
			result := buildSourceProcess(t, project, []string{"yes"}, mutation, "5.5.1")
			if mutation == "replace" {
				require.Equal(t, 1, result.Exit, result.Output)
				require.Contains(t, result.Output, "changed while awaiting confirmation")
				testutil.AssertFileContent(t, source, "replacement")
				testutil.AssertFileContent(t, filepath.Join(source+".saved", "input"), "edited")
				testutil.AssertFileContent(t, filepath.Join(installed, "marker"), "installed")
			} else {
				require.Equal(t, 0, result.Exit, result.Output)
				testutil.AssertFileContent(t, filepath.Join(installed, "marker"), "upstream")
			}
		})
	}
}

func TestBuildSourcePromptRejectsPipedApproval(t *testing.T) {
	project, source, installed := buildSourceProject(t)
	cmd := testutil.TestChild(t, "TestInstallPromptChild", project, "5.5.1")
	cmd.Stdin = strings.NewReader("yes\n")
	data, err := cmd.CombinedOutput()
	require.Error(t, err)
	require.Contains(t, string(data), "interactive terminal")
	testutil.AssertFileContent(t, filepath.Join(source, "input"), "edited")
	testutil.AssertFileContent(t, filepath.Join(installed, "marker"), "installed")
}

func TestBuildSourcePromptRejectsMovedTarget(t *testing.T) {
	project, source, installed := buildSourceProject(t)
	result := buildSourceProcess(t, project, []string{"yes"}, "move", "5.5.1")
	require.Equal(t, 1, result.Exit, result.Output)
	require.Contains(t, result.Output, "cannot recheck build sources")
	_, err := os.Lstat(source)
	require.True(t, os.IsNotExist(err))
	testutil.AssertFileContent(t, filepath.Join(source+".saved", "input"), "edited")
	testutil.AssertFileContent(t, filepath.Join(installed, "marker"), "installed")
}

func TestLuaJitBuildFailureDoesNotInstall(t *testing.T) {
	testutil.PreserveWorkingDirectory(t)
	source := t.TempDir()
	prefix := filepath.Join(t.TempDir(), "installed")
	require.NoError(t, os.Chdir(source))
	// An incomplete source tree makes the real build command fail.
	require.Error(t, installLuaJit(prefix, nil))
	_, err := os.Lstat(prefix)
	require.True(t, os.IsNotExist(err), "build failure created an installation: %v", err)
}

func TestLuaJitInstallFailureIsReported(t *testing.T) {
	testutil.PreserveWorkingDirectory(t)
	source := t.TempDir()
	prefix := filepath.Join(t.TempDir(), "installed")
	require.NoError(t, os.WriteFile(filepath.Join(source, "Makefile"), []byte("all: built\nbuilt:\n\tprintf built > built\n"), 0600))
	require.NoError(t, os.Chdir(source))
	// The source builds, but has no install target.
	require.Error(t, installLuaJit(prefix, []string{"MACOSX_DEPLOYMENT_TARGET=10.13"}))
	testutil.AssertFileContent(t, filepath.Join(source, "built"), "built")
	_, err := os.Lstat(prefix)
	require.True(t, os.IsNotExist(err), "failed install created an installation: %v", err)
}
