package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func buildSourceProcess(t *testing.T, project string, answers []string, mutation string, args ...string) recoveryProcessResult {
	t.Helper()
	script := strings.ReplaceAll(recoveryPTY, "Delete and recover? [y/N] ", "Delete existing build sources? [y/N] ")
	script = strings.ReplaceAll(script, `os.path.join(directory, ".lenv", "current")`, `os.path.join(directory, ".lenv", "build", "lua", "5.5.1")`)
	answerJSON, err := json.Marshal(answers)
	require.NoError(t, err)
	argJSON, err := json.Marshal(args)
	require.NoError(t, err)
	cmd := exec.Command("python3", "-c", script, os.Args[0], project, string(answerJSON), mutation, string(argJSON))
	data, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", data)
	var result recoveryProcessResult
	require.NoError(t, json.Unmarshal(data, &result), "%s", data)
	return result
}

func buildSourceProject(t *testing.T) (string, string, string) {
	t.Helper()
	project, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	root := filepath.Join(project, ".lenv")
	source := filepath.Join(root, "build/lua/5.5.1")
	installed := filepath.Join(root, "lua/5.5.1")
	for _, dir := range []string{source, installed, filepath.Join(root, "src"), filepath.Join(root, "luajit"), filepath.Join(installed, "luarocks")} {
		require.NoError(t, os.MkdirAll(dir, 0755))
	}
	require.NoError(t, os.Symlink("lua/5.5.1", filepath.Join(root, "current")))
	require.NoError(t, os.WriteFile(filepath.Join(installed, "marker"), []byte("installed"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(source, "input"), []byte("edited"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(source, "Makefile"), []byte(rebuildMakefile), 0600))
	data := rebuildLuaArchive(t, false)
	require.NoError(t, os.WriteFile(filepath.Join(root, "src/lua-5.5.1.tar.gz"), data, 0600))
	versions := NewVersions()
	versions.Add("lua-5.5.1.tar.gz", "5.5.1", archiveSHA256(data), ".tar.gz")
	require.NoError(t, versions.WriteFile(filepath.Join(root, "lua_vers.txt")))
	return project, source, installed
}

func TestBuildSourcePromptCancellation(t *testing.T) {
	for _, answer := range []string{"n", "", "EOF", "not-yes"} {
		t.Run(answer, func(t *testing.T) {
			project, source, installed := buildSourceProject(t)
			result := buildSourceProcess(t, project, []string{answer}, "", "-p", "install", "5.5.1")
			require.Equal(t, 1, result.Exit, result.Output)
			require.Contains(t, result.Output, source)
			require.Contains(t, result.Output, "ALL source changes and build output")
			require.Contains(t, result.Output, "source replacement cancelled")
			require.Contains(t, result.Output, "--rebuild")
			require.Contains(t, result.Output, "--no-rebuild")
			assertArchiveFile(t, filepath.Join(source, "input"), "edited")
			assertArchiveFile(t, filepath.Join(installed, "marker"), "installed")
		})
	}
}

func TestBuildSourcePromptApprovalAndExplicitModes(t *testing.T) {
	for _, mode := range []string{"yes", "y", "--rebuild", "--no-rebuild"} {
		t.Run(mode, func(t *testing.T) {
			project, source, installed := buildSourceProject(t)
			args := []string{"-p", "install"}
			var answers []string
			if strings.HasPrefix(mode, "--") {
				args = append(args, mode)
			} else {
				answers = []string{mode}
			}
			args = append(args, "5.5.1")
			result := buildSourceProcess(t, project, answers, "", args...)
			require.Equal(t, 0, result.Exit, result.Output)
			want := "upstream"
			if mode == "--rebuild" {
				want = "edited"
				assertArchiveFile(t, filepath.Join(source, "input"), want)
			} else {
				_, err := os.Lstat(source)
				require.True(t, os.IsNotExist(err))
			}
			assertArchiveFile(t, filepath.Join(installed, "marker"), want)
			if strings.HasPrefix(mode, "--") {
				require.NotContains(t, result.Output, "Delete existing build sources?")
			}
		})
	}
}

func TestBuildSourcePromptRechecksTarget(t *testing.T) {
	for _, mutation := range []string{"replace", "contents"} {
		t.Run(mutation, func(t *testing.T) {
			project, source, installed := buildSourceProject(t)
			result := buildSourceProcess(t, project, []string{"yes"}, mutation, "-p", "install", "5.5.1")
			if mutation == "replace" {
				require.Equal(t, 1, result.Exit, result.Output)
				require.Contains(t, result.Output, "changed while awaiting confirmation")
				assertArchiveFile(t, source, "replacement")
				assertArchiveFile(t, filepath.Join(source+".saved", "input"), "edited")
				assertArchiveFile(t, filepath.Join(installed, "marker"), "installed")
			} else {
				require.Equal(t, 0, result.Exit, result.Output)
				assertArchiveFile(t, filepath.Join(installed, "marker"), "upstream")
			}
		})
	}
}

func TestBuildSourcePromptRejectsPipedApproval(t *testing.T) {
	project, source, installed := buildSourceProject(t)
	cmd := commandForLenvTest("TestRecoverySubprocess", "-p", "install", "5.5.1")
	cmd.Dir = project
	cmd.Env = append(os.Environ(), "LENV_RECOVERY_PROCESS=1", `LENV_RECOVERY_ARGS=["-p","install","5.5.1"]`)
	cmd.Stdin = strings.NewReader("yes\n")
	data, err := cmd.CombinedOutput()
	require.Error(t, err)
	require.Contains(t, string(data), "interactive terminal")
	assertArchiveFile(t, filepath.Join(source, "input"), "edited")
	assertArchiveFile(t, filepath.Join(installed, "marker"), "installed")
}
