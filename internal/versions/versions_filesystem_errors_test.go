package versions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionFileOperationsRejectInvalidFilesystemEntries(t *testing.T) {
	root := t.TempDir()
	versions := NewVersions()
	versions.Add("lua-5.5.1.tar.gz", "5.5.1", "", ".tar.gz")
	loop := filepath.Join(root, "loop")
	require.NoError(t, os.Symlink("loop", loop))
	require.Error(t, versions.WriteFile(loop), "stat of a link loop must fail before replacement")
	require.Error(t, versions.WriteFile(filepath.Join(loop, "list")))
	require.Error(t, versions.readFile(root), "a directory is not a readable version list")
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
