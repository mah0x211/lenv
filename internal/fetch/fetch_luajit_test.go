package fetch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/testutil"
	version "github.com/mah0x211/lenv/internal/versions"
)

func githubBranchPage(t *testing.T, count int) []byte {
	t.Helper()
	branches := make([]map[string]string, count)
	for i := range branches {
		branches[i] = map[string]string{"name": fmt.Sprintf("v2.test%d", i)}
	}
	data, err := json.Marshal(branches)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func newFetchFixture(t *testing.T, intercept func(http.ResponseWriter, *http.Request) bool) {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if intercept != nil && intercept(w, r) {
			return
		}
		switch r.URL.Path {
		case "/lua":
			fmt.Fprintf(w, `<tr><td class="name"><a href="lua-5.5.1.tar.gz">lua</a></td><td class="sum">%s</td></tr>`, strings.Repeat("a", 64))
		case "/jit":
			fmt.Fprint(w, `[{"name":"v2.1"}]`)
		case "/rocks":
			fmt.Fprint(w, `<a href="luarocks-3.13.0.tar.gz">archive</a><a href="luarocks-3.13.0.tar.gz.asc">signature</a>`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	oldLua, oldJit, oldRocks := config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg
	t.Cleanup(func() { config.LuaCfg, config.LuaJitCfg, config.LuaRocksCfg = oldLua, oldJit, oldRocks })
	config.LuaCfg = &config.TargetConfig{Name: "lua", VersionFile: filepath.Join(dir, "lua.json"), ReleaseURL: server.URL + "/lua"}
	config.LuaJitCfg = &config.TargetConfig{Name: "luajit", VersionFile: filepath.Join(dir, "jit.json"), ReleaseURL: server.URL + "/jit"}
	config.LuaRocksCfg = &config.TargetConfig{Name: "luarocks", VersionFile: filepath.Join(dir, "rocks.json"), ReleaseURL: server.URL + "/rocks"}
}

func TestFetchLuaJitArchiveBranches(t *testing.T) {
	page := githubBranchPage(t, 100)
	for _, exactPage := range []bool{false, true} {
		t.Run(strconv.FormatBool(exactPage), func(t *testing.T) {
			requests := 0
			server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.URL.Query().Get("per_page") != "100" || r.Header.Get("Accept") != "application/vnd.github+json" {
					t.Error("missing GitHub page parameters or representation")
				}
				switch r.URL.Query().Get("page") {
				case "1":
					_, _ = w.Write(page)
				case "2":
					if exactPage {
						fmt.Fprint(w, "[]")
					} else {
						fmt.Fprint(w, `[{"name":"v2.1"},{"name":"feature/test"}]`)
					}
				default:
					t.Error("unexpected extra page")
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			vers, err := fetchLuaJitVers(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			want := 101
			if exactPage {
				want = 100
			}
			if len(vers.Dict) != want || requests != 2 {
				t.Fatalf("branches = %d, requests = %d; want %d and 2", len(vers.Dict), requests, want)
			}
			if !exactPage {
				item := vers.GetItem("v2.1")
				if item == nil || item.Remote != "origin" || item.Ver != "v2.1" {
					t.Fatalf("incompatible branch item: %#v", item)
				}
				if vers.GetItem("feature/test") != nil {
					t.Fatal("slash-containing version unexpectedly supported")
				}
			}
		})
	}
}

func TestFetchLuaJitFailuresPreserveList(t *testing.T) {
	fullPage := githubBranchPage(t, 100)
	for _, failure := range []string{"http", "limit", "json", "empty", "name", "read", "second page"} {
		t.Run(failure, func(t *testing.T) {
			newFetchFixture(t, nil)
			old := version.NewVersions()
			old.AddBranch("old-branch", "origin")
			if err := old.WriteFile(config.LuaJitCfg.VersionFile); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(config.LuaJitCfg.VersionFile)
			if err != nil {
				t.Fatal(err)
			}
			server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch failure {
				case "http":
					w.WriteHeader(503)
				case "limit":
					w.Header().Set("X-RateLimit-Remaining", "0")
					w.WriteHeader(403)
				case "json":
					fmt.Fprint(w, "{not-json")
				case "empty":
					fmt.Fprint(w, "[]")
				case "name":
					fmt.Fprint(w, `[{"other":"missing-name"}]`)
				case "read":
					conn, _, err := w.(http.Hijacker).Hijack()
					if err != nil {
						t.Error(err)
						return
					}
					defer conn.Close()
					fmt.Fprint(conn, "HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\n[")
				case "second page":
					if r.URL.Query().Get("page") == "1" {
						_, _ = w.Write(fullPage)
					} else {
						w.WriteHeader(503)
					}
				}
			}))
			defer server.Close()
			config.LuaJitCfg.ReleaseURL = server.URL
			if CmdFetch() {
				t.Fatal("fetch unexpectedly succeeded")
			}
			after, err := os.ReadFile(config.LuaJitCfg.VersionFile)
			if err != nil || string(after) != string(before) {
				t.Fatalf("old list changed: %q, %v", after, err)
			}
		})
	}
}

func TestFetchLuaJitArchiveErrors(t *testing.T) {
	for _, address := range []string{"%", "file:///unavailable"} {
		if _, err := fetchLuaJitVers(address); err == nil {
			t.Fatalf("accepted invalid request %q", address)
		}
	}
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(403)
	}))
	_, err := fetchLuaJitVers(server.URL)
	server.Close()
	if err == nil || !strings.Contains(err.Error(), "rate limit") {
		t.Fatalf("missing rate-limit diagnosis: %v", err)
	}
	if _, err := fetchLuaJitVers(server.URL); err == nil {
		t.Fatal("closed endpoint unexpectedly succeeded")
	}
}

func TestFetchLuaJitFiltersUnsafeBranchNames(t *testing.T) {
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"name":"v2.1"},{"name":"."},{"name":".."},{"name":"../escape"},{"name":"a b"},{"name":"a\\b"}]`)
	}))
	defer server.Close()
	vers, err := fetchLuaJitVers(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(vers.Dict) != 1 || vers.GetItem("v2.1") == nil {
		t.Fatalf("unexpected branches: %#v", vers.Dict)
	}
}
