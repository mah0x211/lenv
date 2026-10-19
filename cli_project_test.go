package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func newCLIProject(t *testing.T) (string, string) {
	t.Helper()
	project, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	root := filepath.Join(project, ".lenv")
	for _, name := range []string{"src", "lua/5.4.8", "lua/5.5.1", "luajit/v2.1"} {
		require.NoError(t, os.MkdirAll(filepath.Join(root, name), 0755))
	}
	lua, jit, rocks := NewVersions(), NewVersions(), NewVersions()
	for _, version := range []string{"5.4.8", "5.5.1"} {
		lua.Add("lua-"+version+".tar.gz", version, "", ".tar.gz")
		require.NoError(t, os.WriteFile(filepath.Join(root, "lua", version, "marker"), []byte(version), 0600))
	}
	jit.AddBranch("v2.1", "origin")
	for _, version := range []string{"3.12.2", "3.13.0"} {
		rocks.Add("luarocks-"+version+".tar.gz", version, "", ".tar.gz")
		dir := filepath.Join(root, "lua/5.4.8/luarocks", version, "lua_modules")
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "marker"), []byte(version), 0600))
	}
	for name, versions := range map[string]*Versions{"lua": lua, "luajit": jit, "luarocks": rocks} {
		require.NoError(t, versions.WriteFile(filepath.Join(root, name+"_vers.txt")))
	}
	require.NoError(t, os.Symlink("lua/5.4.8", filepath.Join(root, "current")))
	require.NoError(t, os.Symlink("luarocks/3.12.2/lua_modules", filepath.Join(root, "lua/5.4.8/lua_modules")))
	return project, root
}

func runCLICommand(t *testing.T, project string, args ...string) (int, string, string) {
	t.Helper()
	data, err := json.Marshal(args)
	require.NoError(t, err)
	cmd := commandForLenvTest("TestStartupDiagnosisSubprocess", args...)
	cmd.Dir = project
	cmd.Env = append(os.Environ(), "LENV_DIAG_PROCESS=1", "LENV_DIAG_ARGS="+string(data))
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	code := 0
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		require.ErrorAs(t, err, &exit, "%s", stderr.String())
		code = exit.ExitCode()
	}
	return code, stdout.String(), stderr.String()
}
