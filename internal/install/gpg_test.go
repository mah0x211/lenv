package install

import (
	"archive/tar"
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"os/exec"
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

func TestGPGMissingToolFails(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := verifyLuaRocksSignature("", []byte("archive"), nil); !errors.Is(err, errGPGNotFound) {
		t.Fatalf("missing GPG = %v", err)
	}
}

func TestGPGStatusPolicy(t *testing.T) {
	fpr := "8460980B2B79786DE0C7FCC83FD8F43C2BB3C478"
	valid := "[GNUPG:] VALIDSIG " + fpr + " 2026-01-01 1767225600 0 4 0 1 8 00 " + fpr
	for name, status := range map[string]string{
		"valid":            valid,
		"unknown metadata": "[GNUPG:] FUTURE_FIELD value\n" + valid,
		"rotated key":      strings.ReplaceAll(valid, fpr, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		"subkey":           strings.Replace(valid, "VALIDSIG "+fpr, "VALIDSIG AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if err := checkGPGStatus(status); err != nil {
				t.Fatal(err)
			}
		})
	}
	for name, status := range map[string]string{
		"missing":           "",
		"only goodsig":      "[GNUPG:] GOODSIG 3FD8F43C2BB3C478 signer",
		"short":             "[GNUPG:] VALIDSIG",
		"sha1":              strings.Replace(valid, " 1 8 00 ", " 1 2 00 ", 1),
		"bad":               valid + "\n[GNUPG:] BADSIG key signer",
		"expired signature": valid + "\n[GNUPG:] EXPSIG key signer",
		"expired key":       valid + "\n[GNUPG:] EXPKEYSIG key signer",
		"revoked":           valid + "\n[GNUPG:] REVKEYSIG key signer",
		"error":             valid + "\n[GNUPG:] ERRSIG key",
	} {
		t.Run(name, func(t *testing.T) {
			if err := checkGPGStatus(status); err == nil {
				t.Fatal("invalid verification status accepted")
			}
		})
	}
}

func TestVerifierRejectsRelativeExecutableDiscovery(t *testing.T) {
	testutil.PreserveWorkingDirectory(t)
	testutil.VerifierPath(t, "gpg")
	require.NoError(t, os.Chdir(os.Getenv("PATH")))
	t.Setenv("PATH", ".")
	verifier, err := findGPG()
	require.Nil(t, verifier)
	require.ErrorIs(t, err, exec.ErrDot)
}

func TestVerifierReportsInaccessibleSearchDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := t.TempDir()
	require.NoError(t, os.Chmod(root, 0))
	t.Cleanup(func() { require.NoError(t, os.Chmod(root, 0755)) })
	t.Setenv("PATH", root)
	verifier, err := findGPG()
	require.Nil(t, verifier)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot inspect gpg entry")
	require.NotErrorIs(t, err, errGPGNotFound)
}

func TestSignatureVerificationReportsUnusableTemporaryDirectory(t *testing.T) {
	data, signature := testutil.OfficialLuaRocksArchive(t)
	address := testutil.PublisherKeyURL(t)
	root := t.TempDir()
	file := filepath.Join(root, "not-a-directory")
	require.NoError(t, os.WriteFile(file, []byte("keep"), 0600))
	t.Setenv("TMPDIR", file)
	err := verifyLuaRocksSignature(address, data, signature)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a directory")
	testutil.AssertFileContent(t, file, "keep")
}

func TestGPGDoesNotUseOrChangeUserKeyring(t *testing.T) {
	data, signature := testutil.OfficialLuaRocksArchive(t)
	home := t.TempDir()
	configuration := filepath.Join(home, "gpg.conf")
	if err := os.WriteFile(configuration, []byte("this-option-is-invalid\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GNUPGHOME", home)
	temp := t.TempDir()
	t.Setenv("TMPDIR", temp)
	if err := verifyLuaRocksSignature(testutil.PublisherKeyURL(t), data, signature); err != nil {
		t.Fatal(err)
	}
	testutil.AssertFileContent(t, configuration, "this-option-is-invalid\n")
	entries, err := os.ReadDir(home)
	if err != nil || len(entries) != 1 {
		t.Fatalf("user GPG directory changed: %v, %v", entries, err)
	}
	testutil.AssertDirectoryEmpty(t, temp)
}

func TestVerifierDiscoveryAndActualVerification(t *testing.T) {
	data, signature := testutil.OfficialLuaRocksArchive(t)
	for _, names := range [][]string{{"gpg", "gpg2", "gpgv"}, {"gpg2", "gpgv"}, {"gpgv"}} {
		t.Run(strings.Join(names, "-"), func(t *testing.T) {
			testutil.VerifierPath(t, names...)
			verifier, err := findGPG()
			if err != nil || verifier.Name != names[0] {
				t.Fatalf("discovery: %v, %v", verifier, err)
			}
			address := testutil.PublisherKeyURL(t)
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
			testutil.VerifierPath(t, "gpgv")
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
	data, signature := testutil.OfficialLuaRocksArchive(t)
	key := testutil.PublisherKeyData(t)
	var requests int32
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	data, signature := testutil.OfficialLuaRocksArchive(t)
	for _, failure := range []string{"empty", "HTML", "invalid armor", "invalid key packets", "HTTP", "downgrade"} {
		t.Run(failure, func(t *testing.T) {
			var insecureRequests int32
			server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestDecodePublicKeys(t *testing.T) {
	key := testutil.PublisherKeyData(t)
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

func TestPublicKeyArmorRequiresHeaderSeparator(t *testing.T) {
	_, err := decodePublicKeys([]byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: test\nYWJj\n-----END PGP PUBLIC KEY BLOCK-----"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing public key armor separator")
}

func TestPublicKeyArmorRequiresClosingFooter(t *testing.T) {
	_, err := decodePublicKeys([]byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nYWJj\n"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing public key armor footer")
}

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

func TestInstallMissingGPGPolicy(t *testing.T) {
	for _, allow := range []bool{false, true} {
		t.Run(map[bool]string{false: "required", true: "explicit warning"}[allow], func(t *testing.T) {
			fixture := newInstallFetchFixture(t, "")
			addIfMissingVersionLists(t)
			t.Setenv("PATH", t.TempDir())
			// This archive must still be rejected by extraction even when PGP is skipped.
			data := testutil.ArchiveData(t, testutil.ArchiveEntry{Name: "../outside", Kind: tar.TypeReg, Body: "rejected"})
			var downloads, signatures int32
			server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&downloads, 1)
				if strings.HasSuffix(r.URL.Path, ".asc") {
					atomic.AddInt32(&signatures, 1)
				}
				w.Write(data)
			}))
			config.LuaRocksCfg.DownloadURL = server.URL + "/"
			args := []string{":latest"}
			if allow {
				args = append([]string{"--allow-missing-gpg"}, args...)
			}
			completed, message := captureInstallError(t, args...)
			if completed || cli.ExitCode != 1 {
				t.Fatal("installation unexpectedly succeeded")
			}
			if allow {
				if !strings.Contains(message, "Warning: no PGP verification command was found") || !strings.Contains(message, "not a local path") {
					t.Fatalf("diagnostic: %s", message)
				}
				if atomic.LoadInt32(&downloads) != 1 || atomic.LoadInt32(&signatures) != 0 {
					t.Fatal("unexpected signature request")
				}
				testutil.AssertFileContent(t, filepath.Join(config.SrcDir, "luarocks-3.13.0.tar.gz"), string(data))
			} else {
				if !strings.Contains(message, "--allow-missing-gpg") || strings.Contains(message, "Warning:") {
					t.Fatalf("diagnostic: %s", message)
				}
				if atomic.LoadInt32(&downloads) != 0 {
					t.Fatal("download started without verifier")
				}
			}
			testutil.AssertFileContent(t, filepath.Join(fixture.dir, "lua/5.4.8/marker"), "installed")
		})
	}
}

func TestAllowMissingGPGStillRejectsInvalidSignature(t *testing.T) {
	newInstallFetchFixture(t, "")
	addIfMissingVersionLists(t)
	versions := version.NewVersions()
	versions.Add("luarocks-3.13.0.tar.gz", "3.13.0", "pgp:luarocks-3.13.0.tar.gz.asc", ".tar.gz")
	if err := versions.WriteFile(config.LuaRocksCfg.VersionFile); err != nil {
		t.Fatal(err)
	}
	data, _ := testutil.OfficialLuaRocksArchive(t)
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".asc") {
			w.Write([]byte("invalid signature"))
		} else {
			w.Write(data)
		}
	}))
	config.LuaRocksCfg.DownloadURL = server.URL + "/"
	completed, message := captureInstallError(t, "--allow-missing-gpg", ":latest")
	if completed || cli.ExitCode != 1 || !strings.Contains(message, "gpg verification failed") || strings.Contains(message, "Warning:") {
		t.Fatalf("invalid signature was bypassed: %s", message)
	}
	if _, err := os.Stat(filepath.Join(config.SrcDir, "luarocks-3.13.0.tar.gz")); !os.IsNotExist(err) {
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
			server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&downloads, 1)
				w.Write([]byte("invalid checksum"))
			}))
			args := []string{"--allow-missing-gpg", "5.5.1"}
			diagnostic := "invalid SHA-256"
			if target == "HTTPS" {
				config.LuaRocksCfg.DownloadURL = "http://" + strings.TrimPrefix(server.URL, "https://") + "/"
				args[1] = ":latest"
				diagnostic = "HTTPS"
			} else {
				config.LuaCfg.DownloadURL = server.URL + "/"
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
	dir := filepath.Join(config.CurrentUsed, "luarocks", "3.13.0")
	installedTestDirectory(t, dir)
	if err := os.Mkdir(filepath.Join(dir, "lua_modules"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	completed, message := captureInstallError(t, "--if-missing", ":latest")
	if !completed || cli.ExitCode != 0 || message != "" {
		t.Fatalf("existing install required gpg: %s", message)
	}
	if atomic.LoadInt32(&fixture.requests) != 0 {
		t.Fatal("existing install downloaded source")
	}
}

func TestExplicitPGPSkipAllowsUnsignedCacheAndDownload(t *testing.T) {
	old := config.SrcDir
	config.SrcDir = t.TempDir()
	t.Cleanup(func() { config.SrcDir = old })
	data := testutil.ArchiveData(t, testutil.ArchiveEntry{Name: "pkg/file", Kind: tar.TypeReg, Body: "source"})
	server := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write(data) }))
	cfg := &config.TargetConfig{Name: "luarocks"}
	item := &version.VerItem{Name: "rocks.tar.gz"}
	root, err := extractDownloadedFile(t.TempDir(), server.URL+"/rocks.tar.gz", cfg, item, true)
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertFileContent(t, filepath.Join(root, "pkg/file"), "source")
	server.Close()
	root, err = extractCachedFile(t.TempDir(), server.URL+"/rocks.tar.gz", cfg, item, true)
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertFileContent(t, filepath.Join(root, "pkg/file"), "source")
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
