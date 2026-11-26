package download_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/mah0x211/lenv/internal/download"
	"github.com/mah0x211/lenv/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestArchiveSaveReportsUnwritableDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := t.TempDir()
	require.NoError(t, os.Chmod(root, 0555))
	t.Cleanup(func() { require.NoError(t, os.Chmod(root, 0755)) })
	require.Error(t, download.SaveDownloadedArchive(filepath.Join(root, "source.tar.gz"), []byte("archive")))
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestSaveDownloadedArchiveFailureCleansTemporaryFile(t *testing.T) {
	parent := t.TempDir()
	blocker := filepath.Join(parent, "blocker")
	if err := os.WriteFile(blocker, []byte("preserved"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := download.SaveDownloadedArchive(filepath.Join(blocker, "file"), []byte("new")); err == nil {
		t.Fatal("accepted non-directory parent")
	}
	testutil.AssertFileContent(t, blocker, "preserved")
	destination := filepath.Join(parent, "directory")
	if err := os.Mkdir(destination, 0700); err != nil {
		t.Fatal(err)
	}
	if err := download.SaveDownloadedArchive(destination, []byte("new")); err == nil {
		t.Fatal("unexpectedly replaced a directory")
	}
	entries, err := os.ReadDir(parent)
	if err != nil || len(entries) != 2 {
		t.Fatalf("temporary archive left after failed rename: %v, %v", entries, err)
	}
}

func TestHTTPSRejectsPlaintextAndDowngrade(t *testing.T) {
	hits := 0
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))
	defer plain.Close()
	if _, err := download.ReadDownload(plain.URL); err == nil {
		t.Fatal("plaintext download accepted")
	}
	secure := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, plain.URL, http.StatusFound)
	}))
	if _, err := download.ReadDownload(secure.URL); err == nil {
		t.Fatal("HTTPS downgrade accepted")
	}
	if hits != 0 {
		t.Fatal("insecure endpoint was contacted")
	}
}

func TestHTTPSRedirectAndCertificateValidation(t *testing.T) {
	secure := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start":
			http.Redirect(w, r, "/end", http.StatusFound)
		case "/loop":
			http.Redirect(w, r, "/loop", http.StatusFound)
		default:
			_, _ = w.Write([]byte("secure"))
		}
	}))
	if data, err := download.ReadDownload(secure.URL + "/start"); err != nil || string(data) != "secure" {
		t.Fatalf("HTTPS redirect = %q, %v", data, err)
	}
	if _, err := download.ReadDownload(secure.URL + "/loop"); err == nil {
		t.Fatal("redirect loop accepted")
	}
	trusted := http.DefaultClient
	http.DefaultClient = &http.Client{}
	defer func() { http.DefaultClient = trusted }()
	if _, err := download.ReadDownload(secure.URL); err == nil {
		t.Fatal("untrusted TLS certificate accepted")
	}
	if _, err := download.ReadDownload("%"); err == nil {
		t.Fatal("invalid URL accepted")
	}
}

func TestHTTPSRespectsExistingRedirectPolicy(t *testing.T) {
	var destinationHits int32
	secure := testutil.NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/destination", http.StatusFound)
			return
		}
		atomic.AddInt32(&destinationHits, 1)
		_, _ = w.Write([]byte("must not reach"))
	}))
	denied := errors.New("redirect denied by caller policy")
	http.DefaultClient.CheckRedirect = func(*http.Request, []*http.Request) error { return denied }
	_, err := download.ReadDownload(secure.URL + "/start")
	if !errors.Is(err, denied) {
		t.Fatalf("caller redirect policy was not retained: %v", err)
	}
	if atomic.LoadInt32(&destinationHits) != 0 {
		t.Fatal("policy-denied destination was contacted")
	}
}
