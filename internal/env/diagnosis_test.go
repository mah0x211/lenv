package env

import (
	"bytes"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/fs"
	"github.com/mah0x211/lenv/internal/testutil/recovery"
	"github.com/stretchr/testify/require"
)

func TestRecoveryDiagnosisDoesNotCreateDirectories(t *testing.T) {
	project := t.TempDir()
	recovery.ConfigureProject(t, project)
	diagnosis := DiagnoseStartup()
	require.NoError(t, diagnosis.Err)
	require.False(t, diagnosis.Ready())
	require.True(t, diagnosis.PendingCurrent)
	require.Len(t, diagnosis.Steps, 4)
	_, err := os.Lstat(config.LenvDir)
	require.True(t, os.IsNotExist(err))
}

func TestRecoveryDiagnosisSuggestionPreservesDirectoryOption(t *testing.T) {
	for _, tc := range []struct{ mode, want string }{
		{"", "lenv setup"},
		{"project", "lenv -p setup"},
		{"global", "lenv -g setup"},
	} {
		require.Equal(t, tc.want, SetupCommand(tc.mode, false))
		require.Equal(t, tc.want+" --recover", SetupCommand(tc.mode, true))
	}
}

func TestStartupDiagnosisIsReadOnly(t *testing.T) {
	for _, kind := range []string{"missing", "healthy", "file", "directory", "dangling", "loop", "to-file", "intermediate-file"} {
		t.Run(kind, func(t *testing.T) {
			project, err := filepath.EvalSymlinks(t.TempDir())
			require.NoError(t, err)
			recovery.ConfigureProject(t, project)
			for _, path := range []string{config.SrcDir, filepath.Join(config.LuaCfg.RootDir, "5.4"), config.LuaJitCfg.RootDir} {
				require.NoError(t, os.MkdirAll(path, 0755))
			}
			switch kind {
			case "healthy":
				require.NoError(t, os.Symlink("lua/5.4", config.CurrentDir))
			case "file":
				require.NoError(t, os.WriteFile(config.CurrentDir, []byte("keep"), 0600))
			case "directory":
				require.NoError(t, os.Mkdir(config.CurrentDir, 0755))
			case "dangling":
				require.NoError(t, os.Symlink("missing", config.CurrentDir))
			case "loop":
				require.NoError(t, os.Symlink("current", config.CurrentDir))
			case "to-file", "intermediate-file":
				target := filepath.Join(config.LuaCfg.RootDir, "plain")
				require.NoError(t, os.WriteFile(target, []byte("keep"), 0600))
				if kind == "intermediate-file" {
					target = filepath.Join(target, "child")
				}
				require.NoError(t, os.Symlink(target, config.CurrentDir))
			}
			before, err := fs.Lstat(config.CurrentDir)
			require.NoError(t, err)
			diagnosis := DiagnoseStartup()
			require.NoError(t, diagnosis.Err)
			if kind == "missing" {
				require.True(t, diagnosis.Ready())
			} else {
				require.False(t, diagnosis.Ready())
				require.Len(t, diagnosis.Steps, 1)
				if kind == "healthy" {
					require.Equal(t, filepath.Join(config.LuaCfg.RootDir, "5.4"), diagnosis.Current)
					require.Equal(t, filepath.Join(diagnosis.Current, "luarocks"), diagnosis.Steps[0].Path)
					require.Nil(t, diagnosis.Steps[0].Info)
				} else {
					require.Equal(t, config.CurrentDir, diagnosis.Steps[0].Path)
					require.True(t, os.SameFile(before, diagnosis.Steps[0].Info))
				}
			}
			after, err := fs.Lstat(config.CurrentDir)
			require.NoError(t, err)
			if before == nil {
				require.Nil(t, after)
			} else {
				require.True(t, os.SameFile(before, after))
			}
			_, err = os.Lstat(filepath.Join(config.LuaCfg.RootDir, "5.4/luarocks"))
			require.True(t, os.IsNotExist(err))
		})
	}
}

func TestStartupDiagnosisReportsMissingAndConflictingDirectories(t *testing.T) {
	for _, kind := range []string{"missing", "file", "link"} {
		t.Run(kind, func(t *testing.T) {
			project := t.TempDir()
			recovery.ConfigureProject(t, project)
			target := filepath.Join(project, "target")
			require.NoError(t, os.Mkdir(target, 0755))
			require.NoError(t, os.MkdirAll(config.LuaCfg.RootDir, 0755))
			require.NoError(t, os.Mkdir(config.LuaJitCfg.RootDir, 0755))
			if kind == "file" {
				require.NoError(t, os.WriteFile(config.SrcDir, []byte("keep"), 0600))
			}
			if kind == "link" {
				require.NoError(t, os.Symlink(target, config.SrcDir))
			}
			before, err := fs.Lstat(config.SrcDir)
			require.NoError(t, err)
			diagnosis := DiagnoseStartup()
			require.NoError(t, diagnosis.Err)
			require.True(t, diagnosis.PendingCurrent)
			require.Len(t, diagnosis.Steps, 1)
			require.Equal(t, config.SrcDir, diagnosis.Steps[0].Path)
			require.True(t, diagnosis.Steps[0].Directory)
			after, err := fs.Lstat(config.SrcDir)
			require.NoError(t, err)
			if before == nil {
				require.Nil(t, after)
			} else {
				require.True(t, os.SameFile(before, after))
			}
		})
	}
}

func TestRecoveryDiagnosisReportsSpecialFile(t *testing.T) {
	recovery.ConfigureProject(t, t.TempDir())
	for _, dir := range []string{config.SrcDir, config.LuaCfg.RootDir, config.LuaJitCfg.RootDir} {
		require.NoError(t, os.MkdirAll(dir, 0755))
	}
	require.NoError(t, syscall.Mkfifo(config.CurrentDir, 0600))

	diagnosis := DiagnoseStartup()
	require.NoError(t, diagnosis.Err)
	require.False(t, diagnosis.Ready())
	var report bytes.Buffer
	diagnosis.Report(&report, "project")
	require.Contains(t, report.String(), "Type: special file")
	info, err := os.Lstat(config.CurrentDir)
	require.NoError(t, err)
	require.NotZero(t, info.Mode()&os.ModeNamedPipe)
}

func TestRecoveryDiagnosisReportsInaccessibleEnvironment(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged user")
	}
	project, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	recovery.ConfigureProject(t, project)
	require.NoError(t, os.MkdirAll(config.LuaJitCfg.RootDir, 0755))
	require.NoError(t, os.Chmod(config.LenvDir, 0))
	t.Cleanup(func() { require.NoError(t, os.Chmod(config.LenvDir, 0755)) })
	diagnosis := DiagnoseStartup()
	require.Error(t, diagnosis.Err)
	require.Contains(t, diagnosis.Err.Error(), "cannot inspect required directory")
	require.False(t, diagnosis.Ready())
	diagnosis = diagnoseCurrent()
	require.Error(t, diagnosis.Err)
	require.Contains(t, diagnosis.Err.Error(), "cannot inspect current selection")
	require.False(t, diagnosis.Ready())
	err = ResolveCurrentDir()
	require.Error(t, err)
	require.Empty(t, config.CurrentUsed)
	require.Empty(t, config.LuaRocksCfg.RootDir)
}
