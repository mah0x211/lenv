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

	version "github.com/mah0x211/lenv/internal/versions"
	"github.com/stretchr/testify/require"
)

func NewInstalledProject(t testing.TB) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), ".lenv")
	for _, name := range []string{"src", "lua/5.4.8", "lua/5.5.1", "luajit/v2.1"} {
		require.NoError(t, os.MkdirAll(filepath.Join(root, name), 0755))
	}
	lua, jit, rocks := version.NewVersions(), version.NewVersions(), version.NewVersions()
	for _, item := range []string{"5.4.8", "5.5.1"} {
		lua.Add("lua-"+item+".tar.gz", item, "", ".tar.gz")
		require.NoError(t, os.WriteFile(filepath.Join(root, "lua", item, "marker"), []byte(item), 0600))
	}
	jit.AddBranch("v2.1", "origin")
	for _, item := range []string{"3.12.2", "3.13.0"} {
		rocks.Add("luarocks-"+item+".tar.gz", item, "", ".tar.gz")
		dir := filepath.Join(root, "lua/5.4.8/luarocks", item, "lua_modules")
		require.NoError(t, os.MkdirAll(dir, 0755))
	}
	for name, versions := range map[string]*version.Versions{"lua": lua, "luajit": jit, "luarocks": rocks} {
		require.NoError(t, versions.WriteFile(filepath.Join(root, name+"_vers.txt")))
	}
	require.NoError(t, os.Symlink("lua/5.4.8", filepath.Join(root, "current")))
	require.NoError(t, os.Symlink("luarocks/3.12.2/lua_modules", filepath.Join(root, "lua/5.4.8/lua_modules")))
	return root
}

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

func UntrustedSignature(t testing.TB) []byte {
	t.Helper()
	return readFixture(t, "untrusted-signature.asc")
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
