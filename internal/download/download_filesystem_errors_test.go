package download_test

import (
	"os"
	"path/filepath"
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
