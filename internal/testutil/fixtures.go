package testutil

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func ArchiveSHA256(data []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(data))
}

func readFixture(t testing.TB, name string) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test fixture directory")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func OfficialLuaRocksArchive(t testing.TB) ([]byte, []byte) {
	t.Helper()
	return readFixture(t, "luarocks-3.13.0.tar.gz"), readFixture(t, "luarocks-3.13.0.tar.gz.asc")
}

func PublisherKeyData(t testing.TB) []byte {
	t.Helper()
	return readFixture(t, "luarocks-publisher.asc")
}

func PublisherKeyURL(t testing.TB) string {
	t.Helper()
	data := PublisherKeyData(t)
	return NewDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(data)
	})).URL + "/publisher.gpg"
}

// Use actual installed GnuPG binaries, with gpg2 as an alternate name for gpg
// when the platform only packages that name. No verification command is faked.
func VerifierPath(t testing.TB, names ...string) {
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
