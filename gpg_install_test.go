package main

import (
	"archive/tar"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func captureInstallError(t *testing.T, args ...string) (bool, string) {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "stderr-")
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stderr
	os.Stderr = file
	defer func() { os.Stderr = old; file.Close() }()
	completed := runInstallForTest(args...)
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return completed, string(data)
}

func TestParseAllowMissingGPG(t *testing.T) {
	for _, args := range [][]string{
		{"--allow-missing-gpg", ":latest"},
		{"--allow-missing-gpg", "--fetch", "--if-missing", "5.5.:latest"},
		{"--if-missing", "--fetch", "--allow-missing-gpg", "5.5.:latest"},
	} {
		opts, err := parseInstallOptions(args)
		if err != nil || !opts.allowMissingGPG {
			t.Fatalf("parse %q: %v, %v", args, opts, err)
		}
		if len(args) == 4 && (!opts.fetch || !opts.ifMissing) {
			t.Fatal("combined flags lost")
		}
	}
	opts, err := parseInstallOptions([]string{"latest", "--allow-missing-gpg"})
	if err != nil || opts.allowMissingGPG || len(opts.build) != 1 {
		t.Fatalf("build argument: %v, %v", opts, err)
	}
	if _, err := parseInstallOptions([]string{"--allow-missing-gpg"}); err == nil {
		t.Fatal("missing version accepted")
	}
}

func TestInstallMissingGPGPolicy(t *testing.T) {
	for _, allow := range []bool{false, true} {
		t.Run(map[bool]string{false: "required", true: "explicit warning"}[allow], func(t *testing.T) {
			fixture := newInstallFetchFixture(t, "")
			addIfMissingVersionLists(t)
			t.Setenv("PATH", t.TempDir())
			// This archive must still be rejected by extraction even when PGP is skipped.
			data := archiveData(t, archiveEntry{name: "../outside", kind: tar.TypeReg, body: "rejected"})
			var downloads, signatures int32
			server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&downloads, 1)
				if strings.HasSuffix(r.URL.Path, ".asc") {
					atomic.AddInt32(&signatures, 1)
				}
				w.Write(data)
			}))
			LuaRocksCfg.DownloadURL = server.URL + "/"
			args := []string{":latest"}
			if allow {
				args = append([]string{"--allow-missing-gpg"}, args...)
			}
			completed, message := captureInstallError(t, args...)
			if completed || ExitCode != 1 {
				t.Fatal("installation unexpectedly succeeded")
			}
			if allow {
				if !strings.Contains(message, "Warning: no PGP verification command was found") || !strings.Contains(message, "not a local path") {
					t.Fatalf("diagnostic: %s", message)
				}
				if atomic.LoadInt32(&downloads) != 1 || atomic.LoadInt32(&signatures) != 0 {
					t.Fatal("unexpected signature request")
				}
				assertArchiveFile(t, filepath.Join(SrcDir, "luarocks-3.13.0.tar.gz"), string(data))
			} else {
				if !strings.Contains(message, "--allow-missing-gpg") || strings.Contains(message, "Warning:") {
					t.Fatalf("diagnostic: %s", message)
				}
				if atomic.LoadInt32(&downloads) != 0 {
					t.Fatal("download started without verifier")
				}
			}
			assertArchiveFile(t, filepath.Join(fixture.dir, "lua/5.4.8/marker"), "installed")
		})
	}
}

func TestAllowMissingGPGStillRejectsInvalidSignature(t *testing.T) {
	newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t)
	versions := NewVersions()
	versions.Add("luarocks-3.13.0.tar.gz", "3.13.0", "pgp:luarocks-3.13.0.tar.gz.asc", ".tar.gz")
	if err := versions.WriteFile(LuaRocksCfg.VersionFile); err != nil {
		t.Fatal(err)
	}
	data, _ := officialLuaRocksArchive(t)
	server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".asc") {
			w.Write([]byte("invalid signature"))
		} else {
			w.Write(data)
		}
	}))
	LuaRocksCfg.DownloadURL = server.URL + "/"
	completed, message := captureInstallError(t, "--allow-missing-gpg", ":latest")
	if completed || ExitCode != 1 || !strings.Contains(message, "gpg verification failed") || strings.Contains(message, "Warning:") {
		t.Fatalf("invalid signature was bypassed: %s", message)
	}
	if _, err := os.Stat(filepath.Join(SrcDir, "luarocks-3.13.0.tar.gz")); !os.IsNotExist(err) {
		t.Fatalf("unverified cache published: %v", err)
	}
}

func TestAllowMissingGPGKeepsOtherChecks(t *testing.T) {
	for _, target := range []string{"lua checksum", "HTTPS"} {
		t.Run(target, func(t *testing.T) {
			newInstallFetchFixture(t, "")
			addIfMissingVersionLists(t)
			t.Setenv("PATH", t.TempDir())
			var downloads int32
			server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&downloads, 1)
				w.Write([]byte("invalid checksum"))
			}))
			args := []string{"--allow-missing-gpg", "5.5.1"}
			diagnostic := "invalid SHA-256"
			if target == "HTTPS" {
				LuaRocksCfg.DownloadURL = "http://" + strings.TrimPrefix(server.URL, "https://") + "/"
				args[1] = ":latest"
				diagnostic = "HTTPS"
			} else {
				LuaCfg.DownloadURL = server.URL + "/"
			}
			completed, message := captureInstallError(t, args...)
			if completed || !strings.Contains(message, diagnostic) {
				t.Fatalf("validation bypassed: %s", message)
			}
			if target == "HTTPS" && atomic.LoadInt32(&downloads) != 0 {
				t.Fatal("insecure request reached server")
			}
		})
	}
}

func TestIfMissingLuaRocksDoesNotRequireGPG(t *testing.T) {
	fixture := newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t)
	dir := filepath.Join(CurrentUsed, "luarocks", "3.13.0")
	installedTestDirectory(t, dir)
	if err := os.Mkdir(filepath.Join(dir, "lua_modules"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	completed, message := captureInstallError(t, "--if-missing", ":latest")
	if !completed || ExitCode != 0 || message != "" {
		t.Fatalf("existing install required gpg: %s", message)
	}
	if atomic.LoadInt32(&fixture.requests) != 0 {
		t.Fatal("existing install downloaded source")
	}
}

func TestExplicitPGPSkipAllowsUnsignedCacheAndDownload(t *testing.T) {
	old := SrcDir
	SrcDir = t.TempDir()
	t.Cleanup(func() { SrcDir = old })
	data := archiveData(t, archiveEntry{name: "pkg/file", kind: tar.TypeReg, body: "source"})
	server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write(data) }))
	cfg := &TargetConfig{Name: "luarocks"}
	item := &VerItem{Name: "rocks.tar.gz"}
	root, err := extractDownloadedFile(t.TempDir(), server.URL+"/rocks.tar.gz", cfg, item, true)
	if err != nil {
		t.Fatal(err)
	}
	assertArchiveFile(t, filepath.Join(root, "pkg/file"), "source")
	server.Close()
	root, err = extractCachedFile(t.TempDir(), server.URL+"/rocks.tar.gz", cfg, item, true)
	if err != nil {
		t.Fatal(err)
	}
	assertArchiveFile(t, filepath.Join(root, "pkg/file"), "source")
}

func TestAllowMissingGPGRejectsUnusableEntry(t *testing.T) {
	for _, kind := range []string{"file", "directory", "broken symlink"} {
		t.Run(kind, func(t *testing.T) {
			fixture := newInstallFetchFixture(t, "")
			addIfMissingVersionLists(t)
			dir := t.TempDir()
			file := filepath.Join(dir, "gpg")
			var err error
			switch kind {
			case "file":
				err = os.WriteFile(file, []byte("not executable"), 0600)
			case "directory":
				err = os.Mkdir(file, 0700)
			case "broken symlink":
				err = os.Symlink(filepath.Join(dir, "missing"), file)
			}
			if err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir)
			completed, message := captureInstallError(t, "--allow-missing-gpg", ":latest")
			if completed || !strings.Contains(message, "exists but is not executable") || strings.Contains(message, "Warning:") {
				t.Fatalf("unusable gpg was bypassed: %s", message)
			}
			if atomic.LoadInt32(&fixture.requests) != 0 {
				t.Fatal("download started with unusable gpg")
			}
		})
	}
}
