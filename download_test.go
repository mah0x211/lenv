package main

import (
	"archive/tar"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadPublishesCompleteArchive(t *testing.T) {
	oldSrc := SrcDir
	SrcDir = t.TempDir()
	t.Cleanup(func() { SrcDir = oldSrc })
	file := filepath.Join(SrcDir, "archive.tar.gz")
	target := filepath.Join(t.TempDir(), "untouched")
	if err := os.WriteFile(target, []byte("original target"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, file); err != nil {
		t.Fatal(err)
	}
	complete := archiveData(t, archiveEntry{name: "pkg/file", kind: tar.TypeReg, body: "new complete download"})
	server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(complete)
	}))
	defer server.Close()
	root, err := extractDownloadedFile(t.TempDir(), server.URL+"/archive.tar.gz", &TargetConfig{Name: "luajit"}, &VerItem{Name: "master"}, false)
	if err != nil {
		t.Fatal(err)
	}
	assertArchiveFile(t, filepath.Join(root, "pkg/file"), "new complete download")
	assertArchiveFile(t, file, string(complete))
	assertArchiveFile(t, target, "original target")
	info, err := os.Lstat(file)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatalf("published cache is not a regular file: %v, %v", info, err)
	}
	entries, err := os.ReadDir(SrcDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary download remains: %v, %v", entries, err)
	}
}

func TestDownloadFailurePreservesPreviousArchive(t *testing.T) {
	for _, failure := range []string{"status", "truncated"} {
		t.Run(failure, func(t *testing.T) {
			oldSrc := SrcDir
			SrcDir = t.TempDir()
			t.Cleanup(func() { SrcDir = oldSrc })
			file := filepath.Join(SrcDir, "archive.tar.gz")
			if err := os.WriteFile(file, []byte("previous download"), 0600); err != nil {
				t.Fatal(err)
			}
			server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if failure == "status" {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				w.Header().Set("Content-Length", "100")
				fmt.Fprint(w, "incomplete")
			}))
			defer server.Close()
			if _, err := extractDownloadedFile(t.TempDir(), server.URL+"/archive.tar.gz", &TargetConfig{Name: "luajit"}, &VerItem{Name: "master"}, false); err == nil {
				t.Fatal("incomplete/failed download unexpectedly succeeded")
			}
			assertArchiveFile(t, file, "previous download")
			entries, err := os.ReadDir(SrcDir)
			if err != nil || len(entries) != 1 {
				t.Fatalf("failed download left temporary files: %v, %v", entries, err)
			}
		})
	}
}

func TestSaveDownloadedArchiveFailureCleansTemporaryFile(t *testing.T) {
	parent := t.TempDir()
	blocker := filepath.Join(parent, "blocker")
	if err := os.WriteFile(blocker, []byte("preserved"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := saveDownloadedArchive(filepath.Join(blocker, "file"), []byte("new")); err == nil {
		t.Fatal("accepted non-directory parent")
	}
	assertArchiveFile(t, blocker, "preserved")
	destination := filepath.Join(parent, "directory")
	if err := os.Mkdir(destination, 0700); err != nil {
		t.Fatal(err)
	}
	if err := saveDownloadedArchive(destination, []byte("new")); err == nil {
		t.Fatal("unexpectedly replaced a directory")
	}
	entries, err := os.ReadDir(parent)
	if err != nil || len(entries) != 2 {
		t.Fatalf("temporary archive left after failed rename: %v, %v", entries, err)
	}
}
