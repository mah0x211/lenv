package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildSourcePreparationReportsFilesystemFailures(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	for _, stage := range []string{"create parent", "inspect parent", "inspect source", "create temporary tree", "remove source"} {
		t.Run(stage, func(t *testing.T) {
			fixture := newInstallFetchFixture(t, "")
			build := filepath.Join(LenvDir, "build")
			parent := filepath.Join(build, "lua")
			source := filepath.Join(parent, "5.5.1")
			require.NoError(t, os.MkdirAll(parent, 0755))
			blocked, mode := parent, os.FileMode(0555)
			switch stage {
			case "create parent":
				require.NoError(t, os.Remove(parent))
				blocked = build
			case "inspect parent":
				blocked = build
				mode = 0
			case "inspect source":
				mode = 0
			case "remove source":
				require.NoError(t, os.Mkdir(source, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(source, "marker"), []byte("keep"), 0600))
				blocked = source
			}
			require.NoError(t, os.Chmod(blocked, mode))
			t.Cleanup(func() { require.NoError(t, os.Chmod(blocked, 0755)) })
			result, err := prepareBuildSource(LuaCfg, &VerItem{Ver: "5.5.1"}, &installOptions{noRebuild: true})
			require.Error(t, err)
			require.Empty(t, result)
			require.Zero(t, fixture.requests, "a local preparation failure must stop before downloading")
			require.NoError(t, os.Chmod(blocked, 0755))
			if stage == "remove source" {
				assertArchiveFile(t, filepath.Join(source, "marker"), "keep")
			}
		})
	}
}

func TestCachedArchiveReadFailureIsReported(t *testing.T) {
	newInstallFetchFixture(t, "")
	require.NoError(t, os.Mkdir(filepath.Join(SrcDir, "lua.tar.gz"), 0755))
	source, err := extractCachedFile(t.TempDir(), "https://example.invalid/lua.tar.gz", LuaCfg, &VerItem{}, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "open cached file error")
	require.Empty(t, source)
	require.DirExists(t, filepath.Join(SrcDir, "lua.tar.gz"))
}

func TestArchivePathReportsInaccessibleParent(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	root := t.TempDir()
	blocked := filepath.Join(root, "private")
	require.NoError(t, os.Mkdir(blocked, 0))
	t.Cleanup(func() { require.NoError(t, os.Chmod(blocked, 0755)) })
	_, err := archivePath(root, "private/child")
	require.Error(t, err)
	require.Contains(t, err.Error(), "archive entry")
}
