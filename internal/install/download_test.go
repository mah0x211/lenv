package install

import (
	"archive/tar"
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

func TestVerifiedLuaCacheAndDownload(t *testing.T) {
	old := config.SrcDir
	config.SrcDir = t.TempDir()
	t.Cleanup(func() { config.SrcDir = old })
	valid := testutil.ArchiveData(t, testutil.ArchiveEntry{Name: "pkg/file", Kind: tar.TypeReg, Body: "verified"})
	cfg := &config.TargetConfig{Name: "lua"}
	item := &version.VerItem{Name: "lua.tar.gz", Sum: testutil.ArchiveSHA256(valid)}
	var requests int32
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		_, _ = w.Write(valid)
	}))
	address := server.URL + "/lua.tar.gz"
	file := filepath.Join(config.SrcDir, item.Name)
	if err := os.WriteFile(file, []byte("altered cache"), 0600); err != nil {
		t.Fatal(err)
	}
	parent := t.TempDir()
	if _, err := extractCachedFile(parent, address, cfg, item, false); err == nil || !strings.Contains(err.Error(), "SHA-256 mismatch") {
		t.Fatalf("bad cache = %v", err)
	}
	testutil.AssertDirectoryEmpty(t, parent)
	root, err := extractDownloadedFile(parent, address, cfg, item, false)
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertFileContent(t, filepath.Join(root, "pkg/file"), "verified")
	testutil.AssertFileContent(t, file, string(valid))
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
	valid := testutil.ArchiveData(t, testutil.ArchiveEntry{Name: "pkg/file", Kind: tar.TypeReg, Body: "expected"})
	actual := testutil.ArchiveData(t, testutil.ArchiveEntry{Name: "pkg/file", Kind: tar.TypeReg, Body: "altered"})
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(actual)
	}))
	config.LuaCfg.DownloadURL = server.URL + "/"
	versions := version.NewVersions()
	versions.Add("lua-5.4.8.tar.gz", "5.4.8", testutil.ArchiveSHA256(valid), ".tar.gz")
	if err := versions.WriteFile(config.LuaCfg.VersionFile); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(config.SrcDir, "lua-5.4.8.tar.gz")
	if err := os.WriteFile(file, []byte("previous cache"), 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Readlink(config.CurrentDir)
	if err != nil {
		t.Fatal(err)
	}
	if runInstallForTest("5.4.8") || cli.ExitCode != 1 {
		t.Fatal("invalid download reached installation")
	}
	after, err := os.Readlink(config.CurrentDir)
	if err != nil || after != before {
		t.Fatalf("selection changed after failed verification: %q, %v", after, err)
	}
	testutil.AssertFileContent(t, file, "previous cache")
}

func TestVerifiedLuaRocksCacheAndDownload(t *testing.T) {
	old := config.SrcDir
	config.SrcDir = t.TempDir()
	t.Cleanup(func() { config.SrcDir = old })
	data, signature := testutil.OfficialLuaRocksArchive(t)
	cfg := &config.TargetConfig{Name: "luarocks", SigningKeyURL: testutil.PublisherKeyURL(t)}
	item := &version.VerItem{Name: "luarocks-3.13.0.tar.gz", Sum: "pgp:luarocks-3.13.0.tar.gz.asc"}
	var requests int32
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		if strings.HasSuffix(r.URL.Path, ".asc") {
			_, _ = w.Write(signature)
		} else {
			_, _ = w.Write(data)
		}
	}))
	address := server.URL + "/" + item.Name
	file := filepath.Join(config.SrcDir, item.Name)
	if _, err := extractDownloadedFile(t.TempDir(), address, cfg, item, false); err != nil {
		t.Fatal(err)
	}
	testutil.AssertFileContent(t, file, string(data))
	testutil.AssertFileContent(t, file+".asc", string(signature))
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
		testutil.AssertDirectoryEmpty(t, parent)
	}
	if atomic.LoadInt32(&requests) != 2 {
		t.Fatal("cached verification downloaded the archive")
	}
}

func TestLuaRocksDownloadVerificationFailurePreservesCache(t *testing.T) {
	data, signature := testutil.OfficialLuaRocksArchive(t)
	for _, failure := range []string{"missing signature", "bad signature", "bad archive"} {
		t.Run(failure, func(t *testing.T) {
			old := config.SrcDir
			config.SrcDir = t.TempDir()
			t.Cleanup(func() { config.SrcDir = old })
			item := &version.VerItem{Name: "luarocks-3.13.0.tar.gz", Sum: "pgp:luarocks-3.13.0.tar.gz.asc"}
			file := filepath.Join(config.SrcDir, item.Name)
			if err := os.WriteFile(file, []byte("previous archive"), 0600); err != nil {
				t.Fatal(err)
			}
			server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			if _, err := extractDownloadedFile(parent, server.URL+"/"+item.Name, &config.TargetConfig{Name: "luarocks", SigningKeyURL: testutil.PublisherKeyURL(t)}, item, false); err == nil {
				t.Fatal("unverified LuaRocks download accepted")
			}
			testutil.AssertFileContent(t, file, "previous archive")
			testutil.AssertDirectoryEmpty(t, parent)
		})
	}
}

func TestVerifiedDownloadCannotPublishOverDirectory(t *testing.T) {
	for _, target := range []string{"archive", "signature"} {
		t.Run(target, func(t *testing.T) {
			old := config.SrcDir
			config.SrcDir = t.TempDir()
			t.Cleanup(func() { config.SrcDir = old })
			data, signature := testutil.OfficialLuaRocksArchive(t)
			item := &version.VerItem{Name: "luarocks-3.13.0.tar.gz", Sum: "pgp:luarocks-3.13.0.tar.gz.asc"}
			dest := filepath.Join(config.SrcDir, item.Name)
			if target == "signature" {
				dest += ".asc"
			}
			if err := os.Mkdir(dest, 0700); err != nil {
				t.Fatal(err)
			}
			server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, ".asc") {
					_, _ = w.Write(signature)
				} else {
					_, _ = w.Write(data)
				}
			}))
			parent := t.TempDir()
			_, err := extractDownloadedFile(parent, server.URL+"/"+item.Name, &config.TargetConfig{Name: "luarocks", SigningKeyURL: testutil.PublisherKeyURL(t)}, item, false)
			var renameErr *os.LinkError
			require.ErrorAs(t, err, &renameErr)
			require.Equal(t, "rename", renameErr.Op)
			require.Equal(t, dest, renameErr.New)
			testutil.AssertDirectoryEmpty(t, parent)
			info, err := os.Stat(dest)
			if err != nil || !info.IsDir() {
				t.Fatalf("cache blocker changed: %v, %v", info, err)
			}
		})
	}
}
