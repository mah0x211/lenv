package install

import (
	"archive/tar"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/testutil"
	version "github.com/mah0x211/lenv/internal/versions"
)

func TestArchiveExtractionCacheRetryIsIsolated(t *testing.T) {
	oldSrc := config.SrcDir
	config.SrcDir = t.TempDir()
	t.Cleanup(func() { config.SrcDir = oldSrc })
	parent := t.TempDir()
	valid := testutil.ArchiveData(t,
		testutil.ArchiveEntry{Name: "pkg/", Kind: tar.TypeDir},
		testutil.ArchiveEntry{Name: "pkg/link/file", Kind: tar.TypeReg, Body: "downloaded"})
	requests := 0
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write(valid)
	}))
	defer server.Close()
	url := server.URL + "/source.tar.gz"
	cache := filepath.Join(config.SrcDir, "source.tar.gz")
	bad := testutil.ArchiveData(t,
		testutil.ArchiveEntry{Name: "pkg/link", Kind: tar.TypeSymlink, Link: t.TempDir()},
		testutil.ArchiveEntry{Name: "pkg/link/file", Kind: tar.TypeReg, Body: "rejected"})
	if err := os.WriteFile(cache, bad, 0600); err != nil {
		t.Fatal(err)
	}
	cfg := &config.TargetConfig{Name: "lua"}
	item := &version.VerItem{Name: "source.tar.gz", Sum: testutil.ArchiveSHA256(bad)}
	if root, err := extractCachedFile(parent, url, cfg, item, false); err == nil || root != "" {
		t.Fatalf("cached extraction = %q, %v; want failure", root, err)
	}
	testutil.AssertDirectoryEmpty(t, parent)
	testutil.AssertFileContent(t, cache, string(bad))
	if requests != 0 {
		t.Fatal("cached extraction unexpectedly downloaded")
	}
	item.Sum = testutil.ArchiveSHA256(valid)
	root, err := extractDownloadedFile(parent, url, cfg, item, false)
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertFileContent(t, filepath.Join(root, "link/file"), "downloaded")
	testutil.AssertFileContent(t, cache, string(valid))
	if requests != 1 {
		t.Fatalf("download requests = %d; want 1", requests)
	}
	// A successful cached attempt must also get a separate extraction tree.
	cachedRoot, err := extractCachedFile(parent, url, cfg, item, false)
	if err != nil || cachedRoot == root {
		t.Fatalf("cached retry root = %q, error = %v", cachedRoot, err)
	}
	testutil.AssertFileContent(t, filepath.Join(cachedRoot, "link/file"), "downloaded")
	if requests != 1 {
		t.Fatal("successful cached extraction unexpectedly downloaded")
	}
}
