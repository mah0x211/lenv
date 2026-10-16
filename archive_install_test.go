package main

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveExtractionInstallFlow(t *testing.T) {
	for _, failBuild := range []bool{false, true} {
		name := "success"
		if failBuild {
			name = "build failure"
		}
		t.Run(name, func(t *testing.T) {
			base := t.TempDir()
			parent := filepath.Join(base, "tmp")
			if err := os.Mkdir(parent, 0755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("TMPDIR", parent)
			wd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			oldCWD, oldLenv, oldSrc := CWD, LenvDir, SrcDir
			oldCurrent, oldUsed, oldRocks, oldExit := CurrentDir, CurrentUsed, LuaRocksCfg, ExitCode
			t.Cleanup(func() {
				CWD, LenvDir, SrcDir = oldCWD, oldLenv, oldSrc
				CurrentDir, CurrentUsed, LuaRocksCfg, ExitCode = oldCurrent, oldUsed, oldRocks, oldExit
				if err := os.Chdir(wd); err != nil {
					t.Error(err)
				}
			})
			CWD, LenvDir, SrcDir = wd, filepath.Join(base, ".lenv"), filepath.Join(base, ".lenv/src")
			CurrentDir, CurrentUsed = filepath.Join(LenvDir, "current"), ""
			LuaRocksCfg = &TargetConfig{Name: "luarocks"}
			if err := os.MkdirAll(SrcDir, 0755); err != nil {
				t.Fatal(err)
			}
			cfg := &TargetConfig{Name: "lua", RootDir: filepath.Join(LenvDir, "lua"), DownloadURL: "http://127.0.0.1:1/"}
			body := "all:\n\tprintf built > built\ninstall:\n\tmkdir -p \"$(INSTALL_TOP)\"\n\tcp built \"$(INSTALL_TOP)/marker\"\n"
			if failBuild {
				body = "all:\n\tfalse\n"
			}
			data := archiveData(t,
				archiveEntry{name: "source/", kind: tar.TypeDir},
				archiveEntry{name: "source/Makefile", kind: tar.TypeReg, body: body})
			cache := filepath.Join(SrcDir, "source.tar.gz")
			if err := os.WriteFile(cache, data, 0600); err != nil {
				t.Fatal(err)
			}
			ExitCode = 0
			done := make(chan struct{})
			completed := false
			go func() {
				defer close(done)
				doInstall(cfg, &VerItem{Name: "source.tar.gz", Ver: "0.0.1"}, nil, false)
				completed = true
			}()
			<-done
			if completed == failBuild {
				t.Fatalf("completed = %v; failBuild = %v", completed, failBuild)
			}
			assertEmptyExtraction(t, parent)
			assertArchiveFile(t, cache, string(data))
			if !failBuild {
				if ExitCode != 0 {
					t.Fatalf("successful install exit code = %d", ExitCode)
				}
				installed := filepath.Join(cfg.RootDir, "0.0.1")
				assertArchiveFile(t, filepath.Join(installed, "marker"), "built")
				selected, err := os.Stat(CurrentDir)
				if err != nil {
					t.Fatal(err)
				}
				expected, err := os.Stat(installed)
				if err != nil {
					t.Fatal(err)
				}
				// Temporary paths may contain aliases such as /var -> /private/var.
				if !os.SameFile(selected, expected) {
					t.Fatalf("current does not select installed directory %q", installed)
				}
			} else if ExitCode != 1 {
				t.Fatalf("failed build exit code = %d", ExitCode)
			}
		})
	}
}

func TestArchiveExtractionUnavailableParent(t *testing.T) {
	base := t.TempDir()
	file := filepath.Join(base, "file")
	if err := os.WriteFile(file, []byte("preserved"), 0600); err != nil {
		t.Fatal(err)
	}
	data := archiveData(t)
	for _, parent := range []string{file, filepath.Join(base, "missing")} {
		if root, err := untarGz(parent, bytes.NewReader(data)); err == nil || root != "" {
			t.Fatalf("unavailable parent: root = %q, error = %v", root, err)
		}
	}
	assertArchiveFile(t, file, "preserved")
}
