package install

import (
	"fmt"
	"net/http"
	"os"
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

func TestParseInstallOptions(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		fetch   bool
		version string
		build   []string
		invalid bool
	}{
		{"existing", []string{"5.4.", "linux", "MYCFLAGS=-O2"}, false, "5.4.", []string{"linux", "MYCFLAGS=-O2"}, false},
		{"refresh", []string{"--fetch", "5.5.:latest", "linux"}, true, "5.5.:latest", []string{"linux"}, false},
		{"LuaJIT", []string{"--fetch", "lj-v2.1"}, true, "lj-v2.1", []string{}, false},
		{"LuaRocks", []string{"--fetch", ":latest"}, true, ":latest", []string{}, false},
		{"build arguments", []string{"latest", "--fetch", "-j4"}, false, "latest", []string{"--fetch", "-j4"}, false},
		{"missing", nil, false, "", nil, true},
		{"missing after flag", []string{"--fetch"}, false, "", nil, true},
		{"empty", []string{""}, false, "", nil, true},
		{"colon", []string{":"}, false, "", nil, true},
		{"unknown", []string{"--unknown", "latest"}, false, "", nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options, err := parseInstallOptions(tc.args)
			if tc.invalid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.fetch, options.Fetch)
			require.Equal(t, tc.version, options.Version)
			require.Equal(t, tc.build, options.Build)
		})
	}
}

type installFetchFixture struct {
	dir       string
	downloads int32
	requests  int32
}

func newInstallFetchFixture(t *testing.T, failure string) *installFetchFixture {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	fixture := &installFetchFixture{dir: dir}
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&fixture.requests, 1)
		switch r.URL.Path {
		case "/lua":
			if failure == "read" {
				conn, _, err := w.(http.Hijacker).Hijack()
				if err != nil {
					t.Error(err)
					return
				}
				defer conn.Close()
				fmt.Fprint(conn, "HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\nshort")
				return
			}
			if failure == "http" {
				w.WriteHeader(503)
				return
			}
			if failure == "empty" {
				fmt.Fprint(w, "<html>no releases</html>")
				return
			}
			fmt.Fprintf(w, "<tr><td class=\"name\"><a href=\"lua-5.5.1.tar.gz\">lua</a></td><td class=\"sum\">%s</td></tr>", strings.Repeat("a", 64))
		case "/jit":
			if failure == "branches" {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			fmt.Fprint(w, `[{"name":"v2.1"},{"name":"main"}]`)
		case "/rocks":
			fmt.Fprint(w, `<a href="luarocks-3.13.0.tar.gz">archive</a><a href="luarocks-3.13.0.tar.gz.asc">signature</a>`)
		default:
			atomic.AddInt32(&fixture.downloads, 1)
			w.WriteHeader(503)
		}
	}))
	t.Cleanup(server.Close)
	if failure == "connection" {
		server.Close()
	}
	oldLua, oldJit, oldRocks := config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg
	oldCWD, oldLenv, oldSrc, oldCurrent, oldUsed, oldExit := config.CWD, config.LenvDir, config.SrcDir, config.CurrentDir, config.CurrentUsed, cli.ExitCode
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg = oldLua, oldJit, oldRocks
		config.CWD, config.LenvDir, config.SrcDir, config.CurrentDir, config.CurrentUsed, cli.ExitCode = oldCWD, oldLenv, oldSrc, oldCurrent, oldUsed, oldExit
		require.NoError(t, os.Chdir(wd))
	})
	config.CWD = wd
	config.LenvDir = fixture.dir
	config.SrcDir = filepath.Join(fixture.dir, "src")
	config.CurrentDir = filepath.Join(fixture.dir, "current")
	cli.ExitCode = 0
	require.NoError(t, os.MkdirAll(config.SrcDir, 0755))
	config.LuaCfg = &config.TargetConfig{Name: "lua", RootDir: filepath.Join(fixture.dir, "lua"), VersionFile: filepath.Join(fixture.dir, "lua.json"), ReleaseURL: server.URL + "/lua", DownloadURL: server.URL + "/download/"}
	config.LuaJitCfg = &config.TargetConfig{Name: "luajit", RootDir: filepath.Join(fixture.dir, "luajit"), VersionFile: filepath.Join(fixture.dir, "jit.json"), ReleaseURL: server.URL + "/jit", DownloadURL: server.URL + "/jit-download/"}
	config.LuaRocksCfg = &config.TargetConfig{Name: "luarocks", SigningKeyURL: testutil.PublisherKeyURL(t), VersionFile: filepath.Join(fixture.dir, "rocks.json"), ReleaseURL: server.URL + "/rocks", DownloadURL: server.URL + "/download/"}
	old := version.NewVersions()
	old.Add("lua-5.4.8.tar.gz", "5.4.8", "sha256:"+strings.Repeat("b", 64), ".tar.gz")
	require.NoError(t, old.WriteFile(config.LuaCfg.VersionFile))
	runtimeDir := filepath.Join(config.LuaCfg.RootDir, "5.4.8")
	require.NoError(t, os.MkdirAll(runtimeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(runtimeDir, "marker"), []byte("installed"), 0600))
	require.NoError(t, os.Symlink(runtimeDir, config.CurrentDir))
	config.CurrentUsed = runtimeDir
	if failure == "save" {
		require.NoError(t, os.Remove(config.LuaCfg.VersionFile))
		require.NoError(t, os.Mkdir(config.LuaCfg.VersionFile, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(config.LuaCfg.VersionFile, "marker"), []byte("old"), 0600))
	}
	return fixture
}

func TestPrepareInstallFetchBeforeVersionSelection(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	target, options := prepareInstall([]string{"--fetch", "5.5.:latest", "linux"})
	require.Equal(t, "5.5.1", target.Lua.Version.Ver)
	require.Equal(t, "3.13.0", target.LuaRocks.Version.Ver)
	require.Equal(t, []string{"linux"}, options.Build)
	require.Equal(t, int32(3), atomic.LoadInt32(&fixture.requests))
	require.Zero(t, atomic.LoadInt32(&fixture.downloads))
	jit, err := version.NewVersionsFromFile(config.LuaJitCfg.VersionFile)
	require.NoError(t, err)
	require.NotNil(t, jit.GetItem("v2.1"))
}

func TestPrepareInstallWithoutFetchUsesCachedList(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	target, options := prepareInstall([]string{"latest", "linux"})
	require.Equal(t, "5.4.8", target.Lua.Version.Ver)
	require.False(t, options.Fetch)
	require.Zero(t, atomic.LoadInt32(&fixture.requests))
	require.Zero(t, atomic.LoadInt32(&fixture.downloads))
}

func TestInstallFetchFailureStopsBeforeInstalling(t *testing.T) {
	for _, failure := range []string{"http", "empty", "save", "connection", "read", "branches"} {
		t.Run(failure, func(t *testing.T) {
			fixture := newInstallFetchFixture(t, failure)
			before, err := os.Readlink(config.CurrentDir)
			require.NoError(t, err)
			done := make(chan struct{})
			completed := false
			go func() { defer close(done); CmdInstall([]string{"--fetch", "latest"}); completed = true }()
			<-done
			require.False(t, completed)
			require.Equal(t, 1, cli.ExitCode)
			require.Zero(t, atomic.LoadInt32(&fixture.downloads))
			after, err := os.Readlink(config.CurrentDir)
			require.NoError(t, err)
			require.Equal(t, before, after)
			data, err := os.ReadFile(filepath.Join(before, "marker"))
			require.NoError(t, err)
			require.Equal(t, "installed", string(data))
		})
	}
}

func TestPrepareInstallInvalidArguments(t *testing.T) {
	oldExit := cli.ExitCode
	defer func() { cli.ExitCode = oldExit }()
	for _, args := range [][]string{nil, {"--unknown", "latest"}} {
		cli.ExitCode = 0
		done := make(chan struct{})
		completed := false
		go func() { defer close(done); prepareInstall(args); completed = true }()
		<-done
		require.False(t, completed)
		require.Equal(t, 1, cli.ExitCode)
	}
}
