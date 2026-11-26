package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func AssertFileContent(t testing.TB, name, want string) {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil || string(data) != want {
		t.Fatalf("read %q = %q, %v; want %q", name, data, err, want)
	}
}

func AssertDirectoryEmpty(t testing.TB, directory string) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 0 {
		t.Fatalf("failed extraction left entries in %q: %v, %v", directory, entries, err)
	}
}

func PreserveWorkingDirectory(t testing.TB) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Error(err)
		}
	})
}

func CreateDirectoryWithMarker(t testing.TB, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(path, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(path, "marker"), []byte("keep"), 0600))
}
