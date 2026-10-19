package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifierRejectsRelativeExecutableDiscovery(t *testing.T) {
	preserveWorkingDirectory(t)
	verifierPath(t, "gpg")
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
	data, signature := officialLuaRocksArchive(t)
	address := publisherKeyURL(t)
	root := t.TempDir()
	file := filepath.Join(root, "not-a-directory")
	require.NoError(t, os.WriteFile(file, []byte("keep"), 0600))
	t.Setenv("TMPDIR", file)
	err := verifyLuaRocksSignature(address, data, signature)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a directory")
	assertArchiveFile(t, file, "keep")
}
