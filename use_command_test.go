package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUseCommandSelections(t *testing.T) {
	for _, tc := range []struct{ target, runtime, rocks string }{
		{"5.5.1", "lua/5.5.1", ""},
		{"5.5", "lua/5.5.1", ""},
		{"latest", "lua/5.5.1", ""},
		{"lj-v2.1", "luajit/v2.1", ""},
		{":3.13.0", "lua/5.4.8", "3.13.0"},
		{"5.4.8:3.13", "lua/5.4.8", "3.13.0"},
	} {
		t.Run(tc.target, func(t *testing.T) {
			project, root := newCLIProject(t)
			code, _, stderr := runCLICommand(t, project, "-p", "use", tc.target)
			require.Equal(t, 0, code, stderr)
			link, err := os.Readlink(filepath.Join(root, "current"))
			require.NoError(t, err)
			require.Equal(t, tc.runtime, link)
			info, err := os.Lstat(filepath.Join(root, tc.runtime, "luarocks"))
			require.NoError(t, err)
			require.True(t, info.IsDir())
			if tc.rocks != "" {
				link, err = os.Readlink(filepath.Join(root, tc.runtime, "lua_modules"))
				require.NoError(t, err)
				require.Equal(t, filepath.Join("luarocks", tc.rocks, "lua_modules"), link)
			}
			assertArchiveFile(t, filepath.Join(root, "lua/5.4.8/marker"), "5.4.8")
		})
	}
}

func TestUseCommandRejectsUnavailableTargets(t *testing.T) {
	for _, state := range []string{"missing", "file", "modules directory"} {
		t.Run(state, func(t *testing.T) {
			project, root := newCLIProject(t)
			target := "5.5.1"
			message := "is not installed"
			entry := filepath.Join(root, "lua/5.5.1")
			require.NoError(t, os.RemoveAll(entry))
			switch state {
			case "file":
				require.NoError(t, os.WriteFile(entry, []byte("keep"), 0600))
				message = "not a directory"
			case "modules directory":
				target = ":3.13.0"
				modules := filepath.Join(root, "lua/5.4.8/lua_modules")
				require.NoError(t, os.Remove(modules))
				require.NoError(t, os.Mkdir(modules, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(modules, "marker"), []byte("keep"), 0600))
				message = "failed to create symlink"
			}
			code, _, stderr := runCLICommand(t, project, "-p", "use", target)
			require.Equal(t, 1, code, stderr)
			require.Contains(t, stderr, message)
			current, err := os.Readlink(filepath.Join(root, "current"))
			require.NoError(t, err)
			require.Equal(t, "lua/5.4.8", current)
			if state == "file" {
				assertArchiveFile(t, entry, "keep")
			}
			if state == "modules directory" {
				assertArchiveFile(t, filepath.Join(root, "lua/5.4.8/lua_modules/marker"), "keep")
			}
		})
	}
}

func TestUseCommandRejectsInvalidVersionMetadata(t *testing.T) {
	for _, state := range []string{"missing", "invalid", "unsupported archive"} {
		t.Run(state, func(t *testing.T) {
			project, root := newCLIProject(t)
			file := filepath.Join(root, "lua_vers.txt")
			message := "failed to read version file"
			switch state {
			case "missing":
				require.NoError(t, os.Remove(file))
			case "invalid":
				require.NoError(t, os.WriteFile(file, []byte("{"), 0600))
			default:
				versions := NewVersions()
				versions.Add("lua-5.5.1.tar.xz", "5.5.1", "", ".tar.xz")
				require.NoError(t, versions.WriteFile(file))
				message = "unsupported media-type"
			}
			code, _, stderr := runCLICommand(t, project, "-p", "use", "5.5.1")
			require.Equal(t, 1, code, stderr)
			require.Contains(t, stderr, message)
			link, err := os.Readlink(filepath.Join(root, "current"))
			require.NoError(t, err)
			require.Equal(t, "lua/5.4.8", link)
		})
	}
}

func TestUseCommandReportsConflictingLuaRocksDirectory(t *testing.T) {
	project, root := newCLIProject(t)
	rocks := filepath.Join(root, "lua/5.5.1/luarocks")
	require.NoError(t, os.WriteFile(rocks, []byte("keep"), 0600))
	code, _, stderr := runCLICommand(t, project, "-p", "use", "5.5.1")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, stderr, "Recovery is disabled")
	require.Contains(t, stderr, "setup --recover")
	assertArchiveFile(t, rocks, "keep")
}
