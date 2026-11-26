package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/testutil"
	version "github.com/mah0x211/lenv/internal/versions"
	"github.com/stretchr/testify/require"
)

func TestArchiveExtractionValid(t *testing.T) {
	for _, first := range []string{"pkg/", "./", "file"} {
		t.Run(first, func(t *testing.T) {
			parent := t.TempDir()
			entries := []testutil.ArchiveEntry{{Name: first, Kind: tar.TypeDir}}
			wantRoot := filepath.Clean(first)
			if first == "file" {
				entries[0] = testutil.ArchiveEntry{Name: "file", Kind: tar.TypeReg, Body: "first"}
				wantRoot = "."
			}
			entries = append(entries,
				testutil.ArchiveEntry{Name: "pkg/sub/run", Kind: tar.TypeReg, Body: "original"},
				testutil.ArchiveEntry{Name: "pkg/sub/run", Kind: tar.TypeReg, Body: "replacement"},
				testutil.ArchiveEntry{Name: "pkg/sub/", Kind: tar.TypeDir},
				testutil.ArchiveEntry{Name: "pkg/link", Kind: tar.TypeSymlink, Link: "sub/run"},
			)
			root, err := untarGz(parent, bytes.NewReader(testutil.ArchiveData(t, entries...)))
			if err != nil {
				t.Fatal(err)
			}
			attempts, err := os.ReadDir(parent)
			if err != nil || len(attempts) != 1 {
				t.Fatalf("extraction attempts: %v, %v", attempts, err)
			}
			attempt := filepath.Join(parent, attempts[0].Name())
			if root != filepath.Join(attempt, wantRoot) {
				t.Fatalf("root = %q; want %q", root, filepath.Join(attempt, wantRoot))
			}
			testutil.AssertFileContent(t, filepath.Join(attempt, "pkg/sub/run"), "replacement")
			link, err := os.Readlink(filepath.Join(attempt, "pkg/link"))
			if err != nil || link != "sub/run" {
				t.Fatalf("symlink = %q, %v", link, err)
			}
			info, err := os.Stat(filepath.Join(attempt, "pkg/sub/run"))
			if err != nil || info.Mode().Perm() != 0755 {
				t.Fatalf("executable file mode: %v, %v", info, err)
			}
		})
	}
}

func TestArchiveExtractionRejectsUnsafePaths(t *testing.T) {
	base := t.TempDir()
	outside := filepath.Join(base, "outside")
	if err := os.Mkdir(outside, 0755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(outside, "sentinel")
	if err := os.WriteFile(sentinel, []byte("preserved"), 0600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		entries []testutil.ArchiveEntry
	}{
		{"parent path", []testutil.ArchiveEntry{{Name: "../../outside/sentinel", Kind: tar.TypeReg, Body: "changed"}}},
		{"absolute path", []testutil.ArchiveEntry{{Name: sentinel, Kind: tar.TypeReg, Body: "changed"}}},
		{"outside directory", []testutil.ArchiveEntry{{Name: "../../new", Kind: tar.TypeDir}}},
		{"outside symlink", []testutil.ArchiveEntry{{Name: "../../new", Kind: tar.TypeSymlink, Link: sentinel}}},
		{"link parent", []testutil.ArchiveEntry{
			{Name: "pkg/link", Kind: tar.TypeSymlink, Link: outside},
			{Name: "pkg/link/sentinel", Kind: tar.TypeReg, Body: "changed"},
		}},
		{"link leaf", []testutil.ArchiveEntry{
			{Name: "pkg/link", Kind: tar.TypeSymlink, Link: sentinel},
			{Name: "pkg/link", Kind: tar.TypeReg, Body: "changed"},
		}},
		{"link directory", []testutil.ArchiveEntry{
			{Name: "pkg/link", Kind: tar.TypeSymlink, Link: outside},
			{Name: "pkg/link/", Kind: tar.TypeDir},
		}},
		{"relative link", []testutil.ArchiveEntry{
			{Name: "link", Kind: tar.TypeSymlink, Link: "../../outside"},
			{Name: "link/sentinel", Kind: tar.TypeReg, Body: "changed"},
		}},
		{"dangling link", []testutil.ArchiveEntry{
			{Name: "link", Kind: tar.TypeSymlink, Link: "missing"},
			{Name: "link/new", Kind: tar.TypeReg, Body: "changed"},
		}},
		{"internal link", []testutil.ArchiveEntry{
			{Name: "pkg/real/", Kind: tar.TypeDir},
			{Name: "pkg/link", Kind: tar.TypeSymlink, Link: "real"},
			{Name: "pkg/link/new", Kind: tar.TypeReg, Body: "changed"},
		}},
		{"non-directory parent", []testutil.ArchiveEntry{
			{Name: "file", Kind: tar.TypeReg, Body: "original"},
			{Name: "file/new", Kind: tar.TypeReg, Body: "changed"},
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parent := filepath.Join(base, strings.ReplaceAll(tc.name, " ", "-"))
			if err := os.Mkdir(parent, 0755); err != nil {
				t.Fatal(err)
			}
			entries := append([]testutil.ArchiveEntry{{Name: "pkg/partial", Kind: tar.TypeReg, Body: "partial"}}, tc.entries...)
			root, err := untarGz(parent, bytes.NewReader(testutil.ArchiveData(t, entries...)))
			last := tc.entries[len(tc.entries)-1].Name
			if err == nil || root != "" || !strings.Contains(err.Error(), strings.TrimSuffix(last, "/")) {
				t.Fatalf("root = %q, error = %v; want rejected entry %q", root, err, last)
			}
			testutil.AssertDirectoryEmpty(t, parent)
			testutil.AssertFileContent(t, sentinel, "preserved")
			if _, err := os.Lstat(filepath.Join(base, "new")); !os.IsNotExist(err) {
				t.Fatalf("outside entry created: %v", err)
			}
		})
	}
}

func TestArchiveExtractionFailureCleanup(t *testing.T) {
	valid := testutil.ArchiveData(t, testutil.ArchiveEntry{Name: "pkg/partial", Kind: tar.TypeReg, Body: "partial"})
	// Keep a complete first entry and truncate the next tar header.
	raw := make([]byte, 0)
	gz, err := gzip.NewReader(bytes.NewReader(valid))
	if err != nil {
		t.Fatal(err)
	}
	raw, err = io.ReadAll(gz)
	gz.Close()
	if err != nil {
		t.Fatal(err)
	}
	var truncated bytes.Buffer
	writer := gzip.NewWriter(&truncated)
	if _, err := writer.Write(raw[:1025]); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"invalid gzip":  []byte("not gzip"),
		"truncated tar": truncated.Bytes(),
		"unsupported type": testutil.ArchiveData(t,
			testutil.ArchiveEntry{Name: "pkg/partial", Kind: tar.TypeReg, Body: "partial"},
			testutil.ArchiveEntry{Name: "pkg/hardlink", Kind: tar.TypeLink, Link: "pkg/partial"}),
		"directory over file": testutil.ArchiveData(t,
			testutil.ArchiveEntry{Name: "pkg", Kind: tar.TypeReg, Body: "file"},
			testutil.ArchiveEntry{Name: "pkg/", Kind: tar.TypeDir}),
		"file over directory": testutil.ArchiveData(t,
			testutil.ArchiveEntry{Name: "pkg/", Kind: tar.TypeDir},
			testutil.ArchiveEntry{Name: "pkg", Kind: tar.TypeReg, Body: "file"}),
		"symlink over file": testutil.ArchiveData(t,
			testutil.ArchiveEntry{Name: "pkg", Kind: tar.TypeReg, Body: "file"},
			testutil.ArchiveEntry{Name: "pkg", Kind: tar.TypeSymlink, Link: "other"}),
	} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			root, err := untarGz(parent, bytes.NewReader(data))
			if err == nil || root != "" {
				t.Fatalf("root = %q, error = %v; want failure", root, err)
			}
			testutil.AssertDirectoryEmpty(t, parent)
		})
	}
}

// Case-insensitive filesystems can resolve differently spelled archive paths to
// the same symlink. Check the filesystem, rather than tracking header names.
func TestArchiveExtractionCaseAlias(t *testing.T) {
	base := t.TempDir()
	outside := filepath.Join(base, "outside")
	if err := os.Mkdir(outside, 0755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(outside, "sentinel")
	if err := os.WriteFile(sentinel, []byte("preserved"), 0600); err != nil {
		t.Fatal(err)
	}
	probe := filepath.Join(base, "probe")
	if err := os.WriteFile(probe, nil, 0600); err != nil {
		t.Fatal(err)
	}
	_, err := os.Stat(filepath.Join(base, "PROBE"))
	caseSensitive := os.IsNotExist(err)
	if err != nil && !caseSensitive {
		t.Fatal(err)
	}
	parent := filepath.Join(base, "extract")
	if err := os.Mkdir(parent, 0755); err != nil {
		t.Fatal(err)
	}
	root, err := untarGz(parent, bytes.NewReader(testutil.ArchiveData(t,
		testutil.ArchiveEntry{Name: "pkg/", Kind: tar.TypeDir},
		testutil.ArchiveEntry{Name: "pkg/link", Kind: tar.TypeSymlink, Link: outside},
		testutil.ArchiveEntry{Name: "pkg/LINK/sentinel", Kind: tar.TypeReg, Body: "separate"})))
	if caseSensitive {
		if err != nil {
			t.Fatal(err)
		}
		testutil.AssertFileContent(t, filepath.Join(root, "LINK/sentinel"), "separate")
	} else {
		if err == nil || root != "" || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("case alias: root = %q, error = %v; want rejected symlink", root, err)
		}
		testutil.AssertDirectoryEmpty(t, parent)
	}
	testutil.AssertFileContent(t, sentinel, "preserved")
}

func TestArchiveExtractionGlobalPAXMetadata(t *testing.T) {
	for _, name := range []string{"pkg/file", "../outside"} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			data := testutil.ArchiveData(t,
				testutil.ArchiveEntry{Name: "pax_global_header", Kind: tar.TypeXGlobalHeader, PAX: map[string]string{"comment": "archive revision"}},
				testutil.ArchiveEntry{Name: "pkg/", Kind: tar.TypeDir},
				testutil.ArchiveEntry{Name: name, Kind: tar.TypeReg, Body: "content"},
			)
			root, err := untarGz(parent, bytes.NewReader(data))
			if name == "../outside" {
				if err == nil {
					t.Fatal("accepted an entry outside the extraction directory")
				}
				testutil.AssertDirectoryEmpty(t, parent)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Base(root) != "pkg" {
				t.Fatalf("root = %q; want pkg directory", root)
			}
			testutil.AssertFileContent(t, filepath.Join(root, "file"), "content")
			if _, err := os.Lstat(filepath.Join(filepath.Dir(root), "pax_global_header")); !os.IsNotExist(err) {
				t.Fatalf("global metadata became a filesystem entry: %v", err)
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
	data := testutil.ArchiveData(t)
	for _, parent := range []string{file, filepath.Join(base, "missing")} {
		if root, err := untarGz(parent, bytes.NewReader(data)); err == nil || root != "" {
			t.Fatalf("unavailable parent: root = %q, error = %v", root, err)
		}
	}
	testutil.AssertFileContent(t, file, "preserved")
}

func TestArchiveExtractionRejectsBadGzipTrailer(t *testing.T) {
	data := testutil.ArchiveData(t, testutil.ArchiveEntry{Name: "pkg/file", Kind: tar.TypeReg, Body: "data"})
	data[len(data)-8] ^= 1
	parent := t.TempDir()
	if _, err := untarGz(parent, bytes.NewReader(data)); err == nil {
		t.Fatal("accepted corrupted gzip trailer")
	}
	testutil.AssertDirectoryEmpty(t, parent)
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

func TestVerifyLuaArchive(t *testing.T) {
	data := []byte("archive")
	cfg := &config.TargetConfig{Name: "lua"}
	for _, sum := range []string{testutil.ArchiveSHA256(data), strings.ToUpper(strings.TrimPrefix(testutil.ArchiveSHA256(data), "sha256:"))} {
		if !strings.HasPrefix(sum, "sha256:") {
			sum = "sha256:" + sum
		}
		if err := verify(cfg, &version.VerItem{Name: "lua.tar.gz", Sum: sum}, data, nil); err != nil {
			t.Fatal(err)
		}
	}
	for _, sum := range []string{"", "pgp:lua.tar.gz.asc", "sha256:invalid", "sha256:00", testutil.ArchiveSHA256([]byte("altered"))} {
		if err := verify(cfg, &version.VerItem{Name: "lua.tar.gz", Sum: sum}, data, nil); err == nil {
			t.Fatalf("accepted invalid checksum %q", sum)
		}
	}
}

func TestVerifyLuaRocksOfficialArchive(t *testing.T) {
	data, signature := testutil.OfficialLuaRocksArchive(t)
	cfg := &config.TargetConfig{Name: "luarocks", SigningKeyURL: testutil.PublisherKeyURL(t)}
	item := &version.VerItem{Name: "luarocks-3.13.0.tar.gz", Sum: "pgp:luarocks-3.13.0.tar.gz.asc"}
	if err := verify(cfg, item, data, signature); err != nil {
		t.Fatal(err)
	}
	altered := append([]byte(nil), data...)
	altered[len(altered)/2] ^= 1
	for name, inputs := range map[string][][]byte{
		"altered archive":   {altered, signature},
		"missing signature": {data, nil},
		"invalid signature": {data, []byte("invalid")},
		"wrong archive":     {[]byte("another archive"), signature},
	} {
		t.Run(name, func(t *testing.T) {
			if err := verify(cfg, item, inputs[0], inputs[1]); err == nil {
				t.Fatal("unverified archive accepted")
			}
		})
	}
	item.Sum = "pgp:../unrelated.asc"
	if err := verify(cfg, item, data, signature); err == nil {
		t.Fatal("accepted unrelated signature reference")
	}
}

func TestVerifyLuaRocksRejectsOtherSigner(t *testing.T) {
	signature := testutil.UntrustedSignature(t)
	item := &version.VerItem{Name: "rocks.tar.gz", Sum: "pgp:rocks.tar.gz.asc"}
	if err := verify(&config.TargetConfig{Name: "luarocks", SigningKeyURL: testutil.PublisherKeyURL(t)}, item, []byte("archive"), signature); err == nil {
		t.Fatal("accepted an untrusted signing key")
	}
}

func TestVerifyArchiveTargetPolicy(t *testing.T) {
	if err := verify(&config.TargetConfig{Name: "luajit"}, &version.VerItem{}, nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := verify(&config.TargetConfig{Name: "unknown"}, &version.VerItem{}, nil, nil); err == nil {
		t.Fatal("accepted unknown verification policy")
	}
}

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
			oldCWD, oldLenv, oldSrc := config.CWD, config.LenvDir, config.SrcDir
			oldCurrent, oldUsed, oldRocks, oldExit := config.CurrentDir, config.CurrentUsed, config.LuaRocksCfg, cli.ExitCode
			t.Cleanup(func() {
				config.CWD, config.LenvDir, config.SrcDir = oldCWD, oldLenv, oldSrc
				config.CurrentDir, config.CurrentUsed, config.LuaRocksCfg, cli.ExitCode = oldCurrent, oldUsed, oldRocks, oldExit
				if err := os.Chdir(wd); err != nil {
					t.Error(err)
				}
			})
			config.CWD, config.LenvDir, config.SrcDir = wd, filepath.Join(base, ".lenv"), filepath.Join(base, ".lenv/src")
			config.CurrentDir, config.CurrentUsed = filepath.Join(config.LenvDir, "current"), ""
			config.LuaRocksCfg = &config.TargetConfig{Name: "luarocks"}
			if err := os.MkdirAll(config.SrcDir, 0755); err != nil {
				t.Fatal(err)
			}
			cfg := &config.TargetConfig{Name: "lua", RootDir: filepath.Join(config.LenvDir, "lua"), DownloadURL: "http://127.0.0.1:1/"}
			body := "all:\n\tprintf built > built\ninstall:\n\tmkdir -p \"$(INSTALL_TOP)\"\n\tcp built \"$(INSTALL_TOP)/marker\"\n"
			if failBuild {
				body = "all:\n\tfalse\n"
			}
			data := testutil.ArchiveData(t,
				testutil.ArchiveEntry{Name: "source/", Kind: tar.TypeDir},
				testutil.ArchiveEntry{Name: "source/Makefile", Kind: tar.TypeReg, Body: body})
			cache := filepath.Join(config.SrcDir, "source.tar.gz")
			if err := os.WriteFile(cache, data, 0600); err != nil {
				t.Fatal(err)
			}
			cli.ExitCode = 0
			done := make(chan struct{})
			completed := false
			go func() {
				defer close(done)
				doInstall(cfg, &version.VerItem{Name: "source.tar.gz", Ver: "0.0.1", Sum: testutil.ArchiveSHA256(data)}, &options{})
				completed = true
			}()
			<-done
			if completed == failBuild {
				t.Fatalf("completed = %v; failBuild = %v", completed, failBuild)
			}
			testutil.AssertDirectoryEmpty(t, parent)
			testutil.AssertFileContent(t, cache, string(data))
			if !failBuild {
				if cli.ExitCode != 0 {
					t.Fatalf("successful install exit code = %d", cli.ExitCode)
				}
				installed := filepath.Join(cfg.RootDir, "0.0.1")
				testutil.AssertFileContent(t, filepath.Join(installed, "marker"), "built")
				selected, err := os.Stat(config.CurrentDir)
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
			} else if cli.ExitCode != 1 {
				t.Fatalf("failed build exit code = %d", cli.ExitCode)
			}
		})
	}
}
