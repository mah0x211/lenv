package main

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
	_, err := lstat(path)
	require.Error(t, err)
	_, err = isDir(path)
	require.Error(t, err)
	_, err = getFileMode(path)
	require.Error(t, err)
	_, err = readSymlink(path)
	require.Error(t, err)
	_, err = EvalSymlink(path)
	require.Error(t, err)
	require.Error(t, mkdir(path))
	_, err = createFile(path, 0600)
	require.Error(t, err)
	_, err = openFile(path)
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
	require.Error(t, writeFile(destination, 0600, source), "reading directory data must fail")
	_, err = os.Lstat(destination)
	require.True(t, os.IsNotExist(err), "partial output remains: %v", err)
}

func TestVersionFileOperationsRejectInvalidFilesystemEntries(t *testing.T) {
	root := t.TempDir()
	versions := NewVersions()
	versions.Add("lua-5.5.1.tar.gz", "5.5.1", "", ".tar.gz")
	loop := filepath.Join(root, "loop")
	require.NoError(t, os.Symlink("loop", loop))
	require.Error(t, versions.WriteFile(loop), "stat of a link loop must fail before replacement")
	require.Error(t, versions.WriteFile(filepath.Join(loop, "list")))
	require.Error(t, versions.ReadFile(root), "a directory is not a readable version list")
	require.NotNil(t, versions.GetItem("5.5.1"), "failed reads must preserve the existing list")
	link, err := os.Readlink(loop)
	require.NoError(t, err)
	require.Equal(t, "loop", link)
}

func TestVersionFileWriteReportsUnwritableParent(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := t.TempDir()
	require.NoError(t, os.Chmod(root, 0555))
	t.Cleanup(func() { require.NoError(t, os.Chmod(root, 0755)) })
	versions := NewVersions()
	versions.Add("lua-5.5.1.tar.gz", "5.5.1", "", ".tar.gz")
	require.Error(t, versions.WriteFile(filepath.Join(root, "versions")))
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Empty(t, entries, "failed save must not leave partial data")
}

func TestArchiveSaveReportsUnwritableDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := t.TempDir()
	require.NoError(t, os.Chmod(root, 0555))
	t.Cleanup(func() { require.NoError(t, os.Chmod(root, 0755)) })
	require.Error(t, saveDownloadedArchive(filepath.Join(root, "source.tar.gz"), []byte("archive")))
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Empty(t, entries)
}
