package main

import (
	"archive/tar"
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestVerifiedLuaCacheAndDownload(t *testing.T) {
	old := SrcDir
	SrcDir = t.TempDir()
	t.Cleanup(func() { SrcDir = old })
	valid := archiveData(t, archiveEntry{name: "pkg/file", kind: tar.TypeReg, body: "verified"})
	cfg := &TargetConfig{Name: "lua"}
	item := &VerItem{Name: "lua.tar.gz", Sum: archiveSHA256(valid)}
	var requests int32
	server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		_, _ = w.Write(valid)
	}))
	address := server.URL + "/lua.tar.gz"
	file := filepath.Join(SrcDir, item.Name)
	if err := os.WriteFile(file, []byte("altered cache"), 0600); err != nil {
		t.Fatal(err)
	}
	parent := t.TempDir()
	if _, err := extractCachedFile(parent, address, cfg, item, false); err == nil || !strings.Contains(err.Error(), "SHA-256 mismatch") {
		t.Fatalf("bad cache = %v", err)
	}
	assertEmptyExtraction(t, parent)
	root, err := extractDownloadedFile(parent, address, cfg, item, false)
	if err != nil {
		t.Fatal(err)
	}
	assertArchiveFile(t, filepath.Join(root, "pkg/file"), "verified")
	assertArchiveFile(t, file, string(valid))
	server.Close()
	if _, err := extractCachedFile(t.TempDir(), address, cfg, item, false); err != nil {
		t.Fatalf("valid cache should work offline: %v", err)
	}
	if atomic.LoadInt32(&requests) != 1 {
		t.Fatal("cached verification unexpectedly made a network request")
	}
}

func TestRejectedLuaDownloadPreservesInstallation(t *testing.T) {
	newInstallFetchFixture(t, "")
	valid := archiveData(t, archiveEntry{name: "pkg/file", kind: tar.TypeReg, body: "expected"})
	actual := archiveData(t, archiveEntry{name: "pkg/file", kind: tar.TypeReg, body: "altered"})
	server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(actual)
	}))
	LuaCfg.DownloadURL = server.URL + "/"
	versions := NewVersions()
	versions.Add("lua-5.4.8.tar.gz", "5.4.8", archiveSHA256(valid), ".tar.gz")
	if err := versions.WriteFile(LuaCfg.VersionFile); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(SrcDir, "lua-5.4.8.tar.gz")
	if err := os.WriteFile(file, []byte("previous cache"), 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Readlink(CurrentDir)
	if err != nil {
		t.Fatal(err)
	}
	if runInstallForTest("5.4.8") || ExitCode != 1 {
		t.Fatal("invalid download reached installation")
	}
	after, err := os.Readlink(CurrentDir)
	if err != nil || after != before {
		t.Fatalf("selection changed after failed verification: %q, %v", after, err)
	}
	assertArchiveFile(t, file, "previous cache")
}

func TestVerifiedLuaRocksCacheAndDownload(t *testing.T) {
	old := SrcDir
	SrcDir = t.TempDir()
	t.Cleanup(func() { SrcDir = old })
	data, signature := officialLuaRocksArchive(t)
	cfg := &TargetConfig{Name: "luarocks", SigningKeyURL: publisherKeyURL(t)}
	item := &VerItem{Name: "luarocks-3.13.0.tar.gz", Sum: "pgp:luarocks-3.13.0.tar.gz.asc"}
	var requests int32
	server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		if strings.HasSuffix(r.URL.Path, ".asc") {
			_, _ = w.Write(signature)
		} else {
			_, _ = w.Write(data)
		}
	}))
	address := server.URL + "/" + item.Name
	file := filepath.Join(SrcDir, item.Name)
	if _, err := extractDownloadedFile(t.TempDir(), address, cfg, item, false); err != nil {
		t.Fatal(err)
	}
	assertArchiveFile(t, file, string(data))
	assertArchiveFile(t, file+".asc", string(signature))
	server.Close()
	if _, err := extractCachedFile(t.TempDir(), address, cfg, item, false); err != nil {
		t.Fatalf("valid signed cache should reuse the archive: %v", err)
	}
	for _, corrupt := range []string{"signature", "archive", "missing signature"} {
		if err := os.WriteFile(file, data, 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file+".asc", signature, 0600); err != nil {
			t.Fatal(err)
		}
		switch corrupt {
		case "signature":
			if err := os.WriteFile(file+".asc", []byte("invalid"), 0600); err != nil {
				t.Fatal(err)
			}
		case "archive":
			if err := os.WriteFile(file, []byte("altered"), 0600); err != nil {
				t.Fatal(err)
			}
		case "missing signature":
			if err := os.Remove(file + ".asc"); err != nil {
				t.Fatal(err)
			}
		}
		parent := t.TempDir()
		if _, err := extractCachedFile(parent, address, cfg, item, false); err == nil {
			t.Fatalf("accepted cache with %s", corrupt)
		}
		assertEmptyExtraction(t, parent)
	}
	if atomic.LoadInt32(&requests) != 2 {
		t.Fatal("cached verification downloaded the archive")
	}
}

func TestArchiveExtractionRejectsBadGzipTrailer(t *testing.T) {
	data := archiveData(t, archiveEntry{name: "pkg/file", kind: tar.TypeReg, body: "data"})
	data[len(data)-8] ^= 1
	parent := t.TempDir()
	if _, err := untarGz(parent, bytes.NewReader(data)); err == nil {
		t.Fatal("accepted corrupted gzip trailer")
	}
	assertEmptyExtraction(t, parent)
}

func TestLuaRocksDownloadVerificationFailurePreservesCache(t *testing.T) {
	data, signature := officialLuaRocksArchive(t)
	for _, failure := range []string{"missing signature", "bad signature", "bad archive"} {
		t.Run(failure, func(t *testing.T) {
			old := SrcDir
			SrcDir = t.TempDir()
			t.Cleanup(func() { SrcDir = old })
			item := &VerItem{Name: "luarocks-3.13.0.tar.gz", Sum: "pgp:luarocks-3.13.0.tar.gz.asc"}
			file := filepath.Join(SrcDir, item.Name)
			if err := os.WriteFile(file, []byte("previous archive"), 0600); err != nil {
				t.Fatal(err)
			}
			server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, ".asc") {
					switch failure {
					case "missing signature":
						w.WriteHeader(http.StatusNotFound)
					case "bad signature":
						_, _ = w.Write([]byte("invalid"))
					default:
						_, _ = w.Write(signature)
					}
				} else if failure == "bad archive" {
					_, _ = w.Write([]byte("altered"))
				} else {
					_, _ = w.Write(data)
				}
			}))
			parent := t.TempDir()
			if _, err := extractDownloadedFile(parent, server.URL+"/"+item.Name, &TargetConfig{Name: "luarocks", SigningKeyURL: publisherKeyURL(t)}, item, false); err == nil {
				t.Fatal("unverified LuaRocks download accepted")
			}
			assertArchiveFile(t, file, "previous archive")
			assertEmptyExtraction(t, parent)
		})
	}
}

func TestVerifiedDownloadCannotPublishOverDirectory(t *testing.T) {
	for _, target := range []string{"archive", "signature"} {
		t.Run(target, func(t *testing.T) {
			old := SrcDir
			SrcDir = t.TempDir()
			t.Cleanup(func() { SrcDir = old })
			data, signature := officialLuaRocksArchive(t)
			item := &VerItem{Name: "luarocks-3.13.0.tar.gz", Sum: "pgp:luarocks-3.13.0.tar.gz.asc"}
			dest := filepath.Join(SrcDir, item.Name)
			if target == "signature" {
				dest += ".asc"
			}
			if err := os.Mkdir(dest, 0700); err != nil {
				t.Fatal(err)
			}
			server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, ".asc") {
					_, _ = w.Write(signature)
				} else {
					_, _ = w.Write(data)
				}
			}))
			parent := t.TempDir()
			if _, err := extractDownloadedFile(parent, server.URL+"/"+item.Name, &TargetConfig{Name: "luarocks", SigningKeyURL: publisherKeyURL(t)}, item, false); err == nil {
				t.Fatal("unexpectedly replaced cache directory")
			}
			assertEmptyExtraction(t, parent)
			info, err := os.Stat(dest)
			if err != nil || !info.IsDir() {
				t.Fatalf("cache blocker changed: %v, %v", info, err)
			}
		})
	}
}
