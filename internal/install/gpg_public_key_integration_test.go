package install

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/testutil"
	version "github.com/mah0x211/lenv/internal/versions"
)

func TestPublicKeyDownloadFailureCannotBeBypassed(t *testing.T) {
	newInstallFetchFixture(t, "")
	data, signature := testutil.OfficialLuaRocksArchive(t)
	versions := version.NewVersions()
	versions.Add("luarocks-3.13.0.tar.gz", "3.13.0", "pgp:luarocks-3.13.0.tar.gz.asc", ".tar.gz")
	if err := versions.WriteFile(config.LuaRocksCfg.VersionFile); err != nil {
		t.Fatal(err)
	}
	installed := filepath.Join(config.CurrentUsed, "luarocks", "3.13.0")
	installedTestDirectory(t, installed)
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/key":
			w.WriteHeader(http.StatusServiceUnavailable)
		case strings.HasSuffix(r.URL.Path, ".asc"):
			w.Write(signature)
		default:
			w.Write(data)
		}
	}))
	config.LuaRocksCfg.DownloadURL = server.URL + "/"
	config.LuaRocksCfg.SigningKeyURL = server.URL + "/key"
	for name, content := range map[string][]byte{"luarocks-3.13.0.tar.gz": data, "luarocks-3.13.0.tar.gz.asc": signature} {
		if err := os.WriteFile(filepath.Join(config.SrcDir, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	completed, message := captureInstallError(t, "--allow-missing-gpg", ":latest")
	if completed || !strings.Contains(message, "cannot download LuaRocks public key") || strings.Contains(message, "Warning:") {
		t.Fatalf("key failure bypassed: %s", message)
	}
	testutil.AssertFileContent(t, filepath.Join(installed, "marker"), "keep")
	testutil.AssertFileContent(t, filepath.Join(config.SrcDir, "luarocks-3.13.0.tar.gz"), string(data))
}
