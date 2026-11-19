package current

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/testutil"
	"github.com/stretchr/testify/require"
)

func currentSelection(t *testing.T) (string, string) {
	t.Helper()
	root := filepath.Join(t.TempDir(), ".lenv")
	runtime := filepath.Join(root, "lua/5.4.8")
	modules := filepath.Join(runtime, "lua_modules")
	require.NoError(t, os.MkdirAll(filepath.Join(runtime, "luarocks/3.12.2/lua_modules"), 0755))
	require.NoError(t, os.Symlink("lua/5.4.8", filepath.Join(root, "current")))
	require.NoError(t, os.Symlink("luarocks/3.12.2/lua_modules", modules))
	return filepath.Join(root, "current"), modules
}

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
			currentDir, modules := currentSelection(t)
			if tc.state == "no runtime" {
				require.NoError(t, os.Remove(currentDir))
			} else if tc.state != "" {
				require.NoError(t, os.Remove(modules))
				if tc.state == "file" {
					require.NoError(t, os.WriteFile(modules, []byte("keep"), 0600))
				} else if tc.state == "dangling" {
					require.NoError(t, os.Symlink("luarocks/missing/lua_modules", modules))
				}
			}
			opts := []string{}
			if tc.target != "" {
				opts = append(opts, tc.target)
			}
			cli.ExitCode = 0
			out, stderr, completed := testutil.CaptureCommandOutput(t, func() { CmdCurrent(currentDir, opts) })
			require.True(t, completed)
			require.Equal(t, 0, cli.ExitCode, stderr)
			require.Equal(t, tc.want, out)
			require.NotContains(t, stderr, "Error:")
		})
	}
}

func TestCurrentCommandRejectsModulesLinkLoop(t *testing.T) {
	currentDir, modules := currentSelection(t)
	require.NoError(t, os.Remove(modules))
	require.NoError(t, os.Symlink("lua_modules", modules))
	cli.ExitCode = 0
	_, stderr, completed := testutil.CaptureCommandOutput(t, func() { CmdCurrent(currentDir, []string{"luarocks"}) })
	require.False(t, completed)
	require.Equal(t, 1, cli.ExitCode, stderr)
	require.Contains(t, stderr, "failed to EvalSymlink()")
	link, err := os.Readlink(modules)
	require.NoError(t, err)
	require.Equal(t, "lua_modules", link)
}
