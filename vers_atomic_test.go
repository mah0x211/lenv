package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func atomicTestVersions(ver string) *Versions {
	versions := NewVersions()
	versions.Add("lua-"+ver+".tar.gz", ver, strings.Repeat("a", 65536), ".tar.gz")
	return versions
}

func assertNoVersionTemps(t *testing.T, dir string) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, ".lenv-versions-*"))
	require.NoError(t, err)
	require.Empty(t, files)
}

func TestVersionsWriteFileAtomicCreateAndReplace(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "nested", "versions.json")
	old := atomicTestVersions("5.4.8")
	next := atomicTestVersions("5.5.1")
	require.NoError(t, old.WriteFile(filename))
	loaded, err := NewVersionsFromFile(filename)
	require.NoError(t, err)
	require.Equal(t, old, loaded)
	require.NoError(t, os.Chmod(filename, 0640))
	require.NoError(t, next.WriteFile(filename))
	loaded, err = NewVersionsFromFile(filename)
	require.NoError(t, err)
	require.Equal(t, next, loaded)
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filename)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0640), info.Mode().Perm())
	}
	assertNoVersionTemps(t, filepath.Dir(filename))
}

func TestVersionsWriteFileFailedReplacePreservesDestination(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "versions.json")
	require.NoError(t, os.Mkdir(filename, 0755))
	marker := filepath.Join(filename, "existing")
	require.NoError(t, os.WriteFile(marker, []byte("old data"), 0600))
	require.Error(t, atomicTestVersions("5.5.1").WriteFile(filename))
	data, err := os.ReadFile(marker)
	require.NoError(t, err)
	require.Equal(t, "old data", string(data))
	assertNoVersionTemps(t, dir)
}

func TestVersionsWriteFileFailedCreatePreservesList(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix directory permissions required")
	}
	dir := t.TempDir()
	filename := filepath.Join(dir, "versions.json")
	old := atomicTestVersions("5.4.8")
	require.NoError(t, old.WriteFile(filename))
	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.NoError(t, os.Chmod(dir, 0555))
	defer os.Chmod(dir, 0755)
	// Verify that this process cannot bypass directory permissions (for example, root).
	probe, err := os.CreateTemp(dir, ".permission-probe-*")
	if err == nil {
		probe.Close()
		os.Remove(probe.Name())
		t.Skip("process can bypass directory permissions")
	}
	require.Error(t, atomicTestVersions("5.5.1").WriteFile(filename))
	got, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, data, got)
	assertNoVersionTemps(t, dir)
}

func TestVersionsWriteFileInvalidParent(t *testing.T) {
	dir := t.TempDir()
	parent := filepath.Join(dir, "file")
	require.NoError(t, os.WriteFile(parent, []byte("existing"), 0600))
	require.Error(t, atomicTestVersions("5.5.1").WriteFile(filepath.Join(parent, "versions.json")))
	data, err := os.ReadFile(parent)
	require.NoError(t, err)
	require.Equal(t, "existing", string(data))
	assertNoVersionTemps(t, dir)
}

func TestVersionsWriteFileReadersSeeCompleteList(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix atomic rename semantics required")
	}
	dir := t.TempDir()
	filename := filepath.Join(dir, "versions.json")
	old := atomicTestVersions("5.4.8")
	next := atomicTestVersions("5.5.1")
	require.NoError(t, old.WriteFile(filename))
	a, err := json.MarshalIndent(old.dict, "", "  ")
	require.NoError(t, err)
	b, err := json.MarshalIndent(next.dict, "", "  ")
	require.NoError(t, err)
	started := make(chan struct{})
	stop := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		first := true
		for {
			data, err := os.ReadFile(filename)
			if err == nil && !bytes.Equal(data, a) && !bytes.Equal(data, b) {
				err = fmt.Errorf("reader observed an incomplete list (%d bytes)", len(data))
			}
			if first {
				close(started)
				first = false
			}
			if err != nil {
				result <- err
				return
			}
			select {
			case <-stop:
				result <- nil
				return
			default:
			}
		}
	}()
	<-started
	defer func() { close(stop); require.NoError(t, <-result) }()
	for i := 0; i < 30; i++ {
		versions := next
		if i%2 == 0 {
			versions = old
		}
		require.NoError(t, versions.WriteFile(filename))
	}
	assertNoVersionTemps(t, dir)
}
