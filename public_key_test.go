package main

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func publisherKeyData(t *testing.T) []byte {
	t.Helper()
	for _, dir := range []string{"testdata", "../testdata"} {
		data, err := os.ReadFile(filepath.Join(dir, "luarocks-publisher.asc"))
		if err == nil {
			return data
		}
	}
	t.Fatal("publisher key fixture is missing")
	return nil
}

func publisherKeyURL(t *testing.T) string {
	t.Helper()
	data := publisherKeyData(t)
	return newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	})).URL + "/publisher.gpg"
}

// Use actual installed GnuPG binaries, with gpg2 as an alternate name for gpg
// when the platform only packages that name. No verification command is faked.
func verifierPath(t *testing.T, names ...string) {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		source := name
		if name == "gpg2" {
			source = "gpg"
		}
		binary, err := exec.LookPath(source)
		if err != nil {
			t.Fatal(err)
		}
		binary, err = filepath.Abs(binary)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(binary, filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir)
}

func TestVerifierDiscoveryAndActualVerification(t *testing.T) {
	data, signature := officialLuaRocksArchive(t)
	for _, names := range [][]string{{"gpg", "gpg2", "gpgv"}, {"gpg2", "gpgv"}, {"gpgv"}} {
		t.Run(strings.Join(names, "-"), func(t *testing.T) {
			verifierPath(t, names...)
			verifier, err := findGPG()
			if err != nil || verifier.name != names[0] {
				t.Fatalf("discovery: %v, %v", verifier, err)
			}
			address := publisherKeyURL(t)
			if err := verifyLuaRocksSignature(address, data, signature); err != nil {
				t.Fatal(err)
			}
			err = verifyLuaRocksSignature(address, data, []byte("invalid signature"))
			if err == nil || !strings.Contains(err.Error(), names[0]+" verification failed") {
				t.Fatalf("bad signature: %v", err)
			}
		})
	}
}

func TestVerifierDoesNotSkipUnusableEarlierCommand(t *testing.T) {
	for _, name := range []string{"gpg", "gpg2"} {
		t.Run(name, func(t *testing.T) {
			verifierPath(t, "gpgv")
			if err := os.Mkdir(filepath.Join(os.Getenv("PATH"), name), 0700); err != nil {
				t.Fatal(err)
			}
			if _, err := findGPG(); err == nil || !strings.Contains(err.Error(), "not executable") {
				t.Fatalf("unusable entry skipped: %v", err)
			}
		})
	}
}

func TestPublicKeyFetchedForEveryVerification(t *testing.T) {
	data, signature := officialLuaRocksArchive(t)
	key := publisherKeyData(t)
	var requests int32
	server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&requests, 1) == 1 {
			w.Write(key)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	if err := verifyLuaRocksSignature(server.URL, data, signature); err != nil {
		t.Fatal(err)
	}
	if err := verifyLuaRocksSignature(server.URL, data, signature); err == nil || !strings.Contains(err.Error(), "cannot download") {
		t.Fatalf("stale key reused: %v", err)
	}
	if atomic.LoadInt32(&requests) != 2 {
		t.Fatal("key not fetched per verification")
	}
}

func TestPublicKeyAcquisitionErrors(t *testing.T) {
	data, signature := officialLuaRocksArchive(t)
	for _, failure := range []string{"empty", "HTML", "invalid armor", "invalid key packets", "HTTP", "downgrade"} {
		t.Run(failure, func(t *testing.T) {
			var insecureRequests int32
			server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch failure {
				case "HTML":
					w.Write([]byte("<html>error</html>"))
				case "invalid armor":
					w.Write([]byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n\n!\n-----END PGP PUBLIC KEY BLOCK-----"))
				case "invalid key packets":
					w.Write([]byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nYWJj\n-----END PGP PUBLIC KEY BLOCK-----"))
				case "HTTP":
					atomic.AddInt32(&insecureRequests, 1)
				case "downgrade":
					http.Redirect(w, r, "http://example.invalid/key", http.StatusFound)
				}
			}))
			address := server.URL
			if failure == "HTTP" {
				address = "http://" + strings.TrimPrefix(address, "https://")
			}
			if err := verifyLuaRocksSignature(address, data, signature); err == nil {
				t.Fatal("invalid key source accepted")
			}
			if atomic.LoadInt32(&insecureRequests) != 0 {
				t.Fatal("HTTP key endpoint contacted")
			}
		})
	}
}

func TestPublicKeyDownloadFailureCannotBeBypassed(t *testing.T) {
	newInstallFetchFixture(t, "")
	data, signature := officialLuaRocksArchive(t)
	versions := NewVersions()
	versions.Add("luarocks-3.13.0.tar.gz", "3.13.0", "pgp:luarocks-3.13.0.tar.gz.asc", ".tar.gz")
	if err := versions.WriteFile(LuaRocksCfg.VersionFile); err != nil {
		t.Fatal(err)
	}
	installed := filepath.Join(CurrentUsed, "luarocks", "3.13.0")
	installedTestDirectory(t, installed)
	server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/key":
			w.WriteHeader(http.StatusServiceUnavailable)
		case strings.HasSuffix(r.URL.Path, ".asc"):
			w.Write(signature)
		default:
			w.Write(data)
		}
	}))
	LuaRocksCfg.DownloadURL = server.URL + "/"
	LuaRocksCfg.SigningKeyURL = server.URL + "/key"
	for name, content := range map[string][]byte{"luarocks-3.13.0.tar.gz": data, "luarocks-3.13.0.tar.gz.asc": signature} {
		if err := os.WriteFile(filepath.Join(SrcDir, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	completed, message := captureInstallError(t, "--allow-missing-gpg", ":latest")
	if completed || !strings.Contains(message, "cannot download LuaRocks public key") || strings.Contains(message, "Warning:") {
		t.Fatalf("key failure bypassed: %s", message)
	}
	assertArchiveFile(t, filepath.Join(installed, "marker"), "keep")
	assertArchiveFile(t, filepath.Join(SrcDir, "luarocks-3.13.0.tar.gz"), string(data))
}

func TestDecodePublicKeys(t *testing.T) {
	key := publisherKeyData(t)
	decoded, err := decodePublicKeys(key)
	if err != nil || len(decoded) == 0 {
		t.Fatalf("real public key: %v", err)
	}
	for name, data := range map[string][]byte{
		"multiple": append(append([]byte(nil), key...), key...),
		"CRLF":     bytes.ReplaceAll(key, []byte("\n"), []byte("\r\n")),
	} {
		got, err := decodePublicKeys(data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		expected := decoded
		if name == "multiple" {
			expected = append(append([]byte(nil), decoded...), decoded...)
		}
		if !bytes.Equal(got, expected) {
			t.Fatalf("%s changed key bytes", name)
		}
	}
	body := base64.StdEncoding.EncodeToString(decoded)
	withoutCRC := "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\n" + body + "\n-----END PGP PUBLIC KEY BLOCK-----"
	withHeader := strings.Replace(withoutCRC, "BLOCK-----\n\n", "BLOCK-----\nVersion: fixture\n\n", 1)
	for _, data := range []string{withoutCRC, withHeader} {
		got, err := decodePublicKeys([]byte(data))
		if err != nil || !bytes.Equal(got, decoded) {
			t.Fatalf("armor variation: %v", err)
		}
	}
	for _, data := range []string{"", "garbage", "-----BEGIN PGP PUBLIC KEY BLOCK-----\n", "-----BEGIN PGP PUBLIC KEY BLOCK-----\n-----END PGP PUBLIC KEY BLOCK-----", "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\n!\n-----END PGP PUBLIC KEY BLOCK-----", withoutCRC + "garbage"} {
		if _, err := decodePublicKeys([]byte(data)); err == nil {
			t.Fatalf("invalid armor accepted: %q", data)
		}
	}
}
