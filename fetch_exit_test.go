package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

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
	LuaCfg.ReleaseURL = address + "/lua"
	LuaJitCfg.ReleaseURL = address + "/jit"
	LuaRocksCfg.ReleaseURL = address + "/rocks"
	os.Args = []string{"lenv", "--project", "fetch"}
	main()
}

func TestFetchExitStatus(t *testing.T) {
	for _, failure := range []string{"", "lua", "jit", "rocks"} {
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
			var hits [3]int32
			server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/lua":
					atomic.AddInt32(&hits[0], 1)
				case "/jit":
					atomic.AddInt32(&hits[1], 1)
				case "/rocks":
					atomic.AddInt32(&hits[2], 1)
				default:
					t.Errorf("unexpected request %q", r.URL.Path)
					w.WriteHeader(404)
					return
				}
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
			for _, target := range []struct{ name, file, old string }{
				{"lua", "lua_vers.txt", "5.4.8"},
				{"jit", "luajit_vers.txt", "v2.0"},
				{"rocks", "luarocks_vers.txt", "3.12.2"},
			} {
				versions := NewVersions()
				if target.name == "jit" {
					versions.AddBranch(target.old, "origin")
				} else {
					versions.Add("old.tar.gz", target.old, "", ".tar.gz")
				}
				require.NoError(t, versions.WriteFile(filepath.Join(root, target.file)))
			}
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
			for i := range hits {
				require.Equal(t, int32(1), atomic.LoadInt32(&hits[i]))
			}
			for _, target := range []struct{ name, file, old, current string }{
				{"lua", "lua_vers.txt", "5.4.8", "5.5.1"},
				{"jit", "luajit_vers.txt", "v2.0", "v2.1"},
				{"rocks", "luarocks_vers.txt", "3.12.2", "3.13.0"},
			} {
				versions, err := NewVersionsFromFile(filepath.Join(root, target.file))
				require.NoError(t, err)
				expected := target.current
				if failure == target.name {
					expected = target.old
				}
				require.NotNil(t, versions.GetItem(expected), "%s", output)
			}
		})
	}
}

func TestFetchDispatchExitCode(t *testing.T) {
	for _, failure := range []string{"", "http"} {
		t.Run("failure="+failure, func(t *testing.T) {
			fixture := newInstallFetchFixture(t, failure)
			project := filepath.Join(fixture.dir, "project")
			for _, dir := range []string{"src", "lua", "luajit"} {
				require.NoError(t, os.MkdirAll(filepath.Join(project, ".lenv", dir), 0755))
			}
			oldArgs := os.Args
			t.Cleanup(func() { os.Args = oldArgs })
			os.Args = []string{"lenv", "--project", "fetch"}
			require.NoError(t, os.Chdir(project))
			start()
			expected := 0
			if failure != "" {
				expected = 1
			}
			require.Equal(t, expected, ExitCode)
		})
	}
}
