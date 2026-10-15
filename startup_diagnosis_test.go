package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func recoveryDiagnosisProject(t *testing.T, project string) {
	t.Helper()
	oldRoot, oldSrc, oldCurrent, oldUsed := LenvDir, SrcDir, CurrentDir, CurrentUsed
	oldLua, oldJit, oldRocks, oldExit := LuaCfg, LuaJitCfg, LuaRocksCfg, ExitCode
	t.Cleanup(func() {
		LenvDir, SrcDir, CurrentDir, CurrentUsed = oldRoot, oldSrc, oldCurrent, oldUsed
		LuaCfg, LuaJitCfg, LuaRocksCfg, ExitCode = oldLua, oldJit, oldRocks, oldExit
	})
	LenvDir = filepath.Join(project, ".lenv")
	SrcDir, CurrentDir = filepath.Join(LenvDir, "src"), filepath.Join(LenvDir, "current")
	LuaCfg = &TargetConfig{RootDir: filepath.Join(LenvDir, "lua")}
	LuaJitCfg = &TargetConfig{RootDir: filepath.Join(LenvDir, "luajit")}
	LuaRocksCfg = &TargetConfig{}
}

func TestRecoveryDiagnosisDoesNotCreateDirectories(t *testing.T) {
	project := t.TempDir()
	recoveryDiagnosisProject(t, project)
	diagnosis := diagnoseStartup()
	require.NoError(t, diagnosis.err)
	require.False(t, diagnosis.ready())
	require.True(t, diagnosis.pendingCurrent)
	require.Len(t, diagnosis.steps, 4)
	_, err := os.Lstat(LenvDir)
	require.True(t, os.IsNotExist(err))
}

func TestRecoveryDiagnosisSuggestionPreservesDirectoryOption(t *testing.T) {
	for _, tc := range []struct{ mode, want string }{
		{"", "lenv setup"},
		{"project", "lenv -p setup"},
		{"global", "lenv -g setup"},
	} {
		require.Equal(t, tc.want, setupCommand(tc.mode, false))
		require.Equal(t, tc.want+" --recover", setupCommand(tc.mode, true))
	}
}

func TestStartupDiagnosisIsReadOnly(t *testing.T) {
	for _, kind := range []string{"missing", "healthy", "file", "directory", "dangling", "loop", "to-file", "intermediate-file"} {
		t.Run(kind, func(t *testing.T) {
			project, err := filepath.EvalSymlinks(t.TempDir())
			require.NoError(t, err)
			recoveryDiagnosisProject(t, project)
			for _, path := range []string{SrcDir, filepath.Join(LuaCfg.RootDir, "5.4"), LuaJitCfg.RootDir} {
				require.NoError(t, os.MkdirAll(path, 0755))
			}
			switch kind {
			case "healthy":
				require.NoError(t, os.Symlink("lua/5.4", CurrentDir))
			case "file":
				require.NoError(t, os.WriteFile(CurrentDir, []byte("keep"), 0600))
			case "directory":
				require.NoError(t, os.Mkdir(CurrentDir, 0755))
			case "dangling":
				require.NoError(t, os.Symlink("missing", CurrentDir))
			case "loop":
				require.NoError(t, os.Symlink("current", CurrentDir))
			case "to-file", "intermediate-file":
				target := filepath.Join(LuaCfg.RootDir, "plain")
				require.NoError(t, os.WriteFile(target, []byte("keep"), 0600))
				if kind == "intermediate-file" {
					target = filepath.Join(target, "child")
				}
				require.NoError(t, os.Symlink(target, CurrentDir))
			}
			before, err := lstat(CurrentDir)
			require.NoError(t, err)
			diagnosis := diagnoseStartup()
			require.NoError(t, diagnosis.err)
			if kind == "missing" {
				require.True(t, diagnosis.ready())
			} else {
				require.False(t, diagnosis.ready())
				require.Len(t, diagnosis.steps, 1)
				if kind == "healthy" {
					require.Equal(t, filepath.Join(LuaCfg.RootDir, "5.4"), diagnosis.current)
					require.Equal(t, filepath.Join(diagnosis.current, "luarocks"), diagnosis.steps[0].path)
					require.Nil(t, diagnosis.steps[0].info)
				} else {
					require.Equal(t, CurrentDir, diagnosis.steps[0].path)
					require.True(t, os.SameFile(before, diagnosis.steps[0].info))
				}
			}
			after, err := lstat(CurrentDir)
			require.NoError(t, err)
			if before == nil {
				require.Nil(t, after)
			} else {
				require.True(t, os.SameFile(before, after))
			}
			_, err = os.Lstat(filepath.Join(LuaCfg.RootDir, "5.4/luarocks"))
			require.True(t, os.IsNotExist(err))
		})
	}
}

func TestStartupDiagnosisReportsMissingAndConflictingDirectories(t *testing.T) {
	for _, kind := range []string{"missing", "file", "link"} {
		t.Run(kind, func(t *testing.T) {
			project := t.TempDir()
			recoveryDiagnosisProject(t, project)
			target := filepath.Join(project, "target")
			require.NoError(t, os.Mkdir(target, 0755))
			require.NoError(t, os.MkdirAll(LuaCfg.RootDir, 0755))
			require.NoError(t, os.Mkdir(LuaJitCfg.RootDir, 0755))
			if kind == "file" {
				require.NoError(t, os.WriteFile(SrcDir, []byte("keep"), 0600))
			}
			if kind == "link" {
				require.NoError(t, os.Symlink(target, SrcDir))
			}
			before, err := lstat(SrcDir)
			require.NoError(t, err)
			diagnosis := diagnoseStartup()
			require.NoError(t, diagnosis.err)
			require.True(t, diagnosis.pendingCurrent)
			require.Len(t, diagnosis.steps, 1)
			require.Equal(t, SrcDir, diagnosis.steps[0].path)
			require.True(t, diagnosis.steps[0].directory)
			after, err := lstat(SrcDir)
			require.NoError(t, err)
			if before == nil {
				require.Nil(t, after)
			} else {
				require.True(t, os.SameFile(before, after))
			}
		})
	}
}
