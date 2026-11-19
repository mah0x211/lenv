package fetch

import (
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/testutil"
	version "github.com/mah0x211/lenv/internal/versions"
	"github.com/stretchr/testify/require"
)

func TestFetchUpdatesVersionFiles(t *testing.T) {
	for _, failure := range []string{"", "/lua", "/jit", "/rocks"} {
		name := failure
		if name == "" {
			name = "success"
		}
		t.Run(name, func(t *testing.T) {
			hits := map[string]*int32{"/lua": new(int32), "/jit": new(int32), "/rocks": new(int32)}
			newFetchFixture(t, func(w http.ResponseWriter, r *http.Request) bool {
				count := hits[r.URL.Path]
				if count == nil {
					t.Errorf("unexpected request %q", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
					return true
				}
				atomic.AddInt32(count, 1)
				if r.URL.Path == failure {
					w.WriteHeader(http.StatusServiceUnavailable)
					return true
				}
				return false
			})
			targets := []struct {
				path, file, old, current string
			}{
				{"/lua", config.LuaCfg.VersionFile, "5.4.8", "5.5.1"},
				{"/jit", config.LuaJitCfg.VersionFile, "v2.0", "v2.1"},
				{"/rocks", config.LuaRocksCfg.VersionFile, "3.12.2", "3.13.0"},
			}
			for _, target := range targets {
				versions := version.NewVersions()
				if target.path == "/jit" {
					versions.AddBranch(target.old, "origin")
				} else {
					versions.Add("old.tar.gz", target.old, "", ".tar.gz")
				}
				require.NoError(t, versions.WriteFile(target.file))
			}
			var success bool
			output, stderr, completed := testutil.CaptureCommandOutput(t, func() { success = CmdFetch() })
			require.True(t, completed)
			require.Equal(t, failure == "", success, stderr)
			require.NotContains(t, output+stderr, "panic:")
			if failure != "" {
				require.Contains(t, stderr, "503 Service Unavailable")
			}
			for _, target := range targets {
				require.Equal(t, int32(1), atomic.LoadInt32(hits[target.path]), target.path)
				versions, err := version.NewVersionsFromFile(target.file)
				require.NoError(t, err)
				expected := target.current
				if failure == target.path {
					expected = target.old
				}
				require.NotNil(t, versions.GetItem(expected), target.path)
			}
		})
	}
}
