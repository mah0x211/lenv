package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCurrentCommandSelections(t *testing.T) {
	for _, tc := range []struct {
		name, target, state, want string
	}{
		{"both", "", "", "lua/5.4.8\nluarocks/3.12.2\n"},
		{"lua", "lua", "", "lua/5.4.8\n"},
		{"rocks", "luarocks", "", "luarocks/3.12.2\n"},
		{"no rocks selected", "", "no rocks", "lua/5.4.8\n"},
		{"missing rocks", "luarocks", "no rocks", "\n"},
		{"no runtime selected", "", "no runtime", "\n"},
		{"no runtime target", "lua", "no runtime", "\n"},
		{"no rocks runtime", "luarocks", "no runtime", "\n"},
		{"regular modules file", "luarocks", "file", "\n"},
		{"dangling modules link", "luarocks", "dangling", "\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			project, root := newCLIProject(t)
			modules := filepath.Join(root, "lua/5.4.8/lua_modules")
			if tc.state == "no runtime" {
				require.NoError(t, os.Remove(filepath.Join(root, "current")))
			} else if tc.state != "" {
				require.NoError(t, os.Remove(modules))
				if tc.state == "file" {
					require.NoError(t, os.WriteFile(modules, []byte("keep"), 0600))
				} else if tc.state == "dangling" {
					require.NoError(t, os.Symlink("luarocks/missing/lua_modules", modules))
				}
			}
			args := []string{"-p", "current"}
			if tc.target != "" {
				args = append(args, tc.target)
			}
			code, out, stderr := runCLICommand(t, project, args...)
			require.Equal(t, 0, code, stderr)
			require.Equal(t, tc.want, out)
			require.NotContains(t, stderr, "Error:")
		})
	}
}

func TestCurrentCommandRejectsModulesLinkLoop(t *testing.T) {
	project, root := newCLIProject(t)
	modules := filepath.Join(root, "lua/5.4.8/lua_modules")
	require.NoError(t, os.Remove(modules))
	require.NoError(t, os.Symlink("lua_modules", modules))
	code, _, stderr := runCLICommand(t, project, "-p", "current", "luarocks")
	require.Equal(t, 1, code, stderr)
	require.Contains(t, stderr, "failed to EvalSymlink()")
	link, err := os.Readlink(modules)
	require.NoError(t, err)
	require.Equal(t, "lua_modules", link)
}
