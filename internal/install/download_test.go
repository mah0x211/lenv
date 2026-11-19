package install

import (
	"archive/tar"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/testutil"
	version "github.com/mah0x211/lenv/internal/versions"
)

func TestDownloadPublishesCompleteArchive(t *testing.T) {
	oldSrc := config.SrcDir
	config.SrcDir = t.TempDir()
	t.Cleanup(func() { config.SrcDir = oldSrc })
	file := filepath.Join(config.SrcDir, "archive.tar.gz")
	target := filepath.Join(t.TempDir(), "untouched")
	if err := os.WriteFile(target, []byte("original target"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, file); err != nil {
		t.Fatal(err)
	}
	complete := testutil.ArchiveData(t, testutil.ArchiveEntry{Name: "pkg/file", Kind: tar.TypeReg, Body: "new complete download"})
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(complete)
	}))
	defer server.Close()
	root, err := extractDownloadedFile(t.TempDir(), server.URL+"/archive.tar.gz", &config.TargetConfig{Name: "luajit"}, &version.VerItem{Name: "master"}, false)
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertFileContent(t, filepath.Join(root, "pkg/file"), "new complete download")
	testutil.AssertFileContent(t, file, string(complete))
	testutil.AssertFileContent(t, target, "original target")
	info, err := os.Lstat(file)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatalf("published cache is not a regular file: %v, %v", info, err)
	}
	entries, err := os.ReadDir(config.SrcDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary download remains: %v, %v", entries, err)
	}
}

func TestDownloadFailurePreservesPreviousArchive(t *testing.T) {
	for _, failure := range []string{"status", "truncated"} {
		t.Run(failure, func(t *testing.T) {
			oldSrc := config.SrcDir
			config.SrcDir = t.TempDir()
			t.Cleanup(func() { config.SrcDir = oldSrc })
			file := filepath.Join(config.SrcDir, "archive.tar.gz")
			if err := os.WriteFile(file, []byte("previous download"), 0600); err != nil {
				t.Fatal(err)
			}
			server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if failure == "status" {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				w.Header().Set("Content-Length", "100")
				fmt.Fprint(w, "incomplete")
			}))
			defer server.Close()
			if _, err := extractDownloadedFile(t.TempDir(), server.URL+"/archive.tar.gz", &config.TargetConfig{Name: "luajit"}, &version.VerItem{Name: "master"}, false); err == nil {
				t.Fatal("incomplete/failed download unexpectedly succeeded")
			}
			testutil.AssertFileContent(t, file, "previous download")
			entries, err := os.ReadDir(config.SrcDir)
			if err != nil || len(entries) != 1 {
				t.Fatalf("failed download left temporary files: %v, %v", entries, err)
			}
		})
	}
}
