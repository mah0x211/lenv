package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilesystemHelpersReportInvalidPaths(t *testing.T) {
	dir := t.TempDir()
	loop := filepath.Join(dir, "loop")
	require.NoError(t, os.Symlink("loop", loop))
	path := filepath.Join(loop, "child")
	_, err := Lstat(path)
	require.Error(t, err)
	_, err = IsDir(path)
	require.Error(t, err)
	_, err = getFileMode(path)
	require.Error(t, err)
	_, err = ReadSymlink(path)
	require.Error(t, err)
	_, err = EvalSymlink(path)
	require.Error(t, err)
	require.Error(t, Mkdir(path))
	_, err = createFile(path, 0600)
	require.Error(t, err)
	_, err = OpenFile(path)
	require.Error(t, err)
	_, err = createFile(dir, 0600)
	require.Error(t, err)
	require.NoError(t, os.Remove(loop))
	require.NoError(t, os.Mkdir(loop, 0755))
	require.NoError(t, os.MkdirAll(path, 0755), "failure must not leave an obstructing child")
}

func TestWriteFileRemovesPartialOutputAfterReadError(t *testing.T) {
	root := t.TempDir()
	source, err := os.Open(root)
	require.NoError(t, err)
	defer source.Close()
	destination := filepath.Join(root, "output")
	require.Error(t, WriteFile(destination, 0600, source), "reading directory data must fail")
	_, err = os.Lstat(destination)
	require.True(t, os.IsNotExist(err), "partial output remains: %v", err)
}
