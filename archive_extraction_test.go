package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type archiveEntry struct {
	name, body, link string
	kind             byte
	pax              map[string]string
}

func archiveData(t *testing.T, entries ...archiveEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, entry := range entries {
		h := &tar.Header{Name: entry.name, Typeflag: entry.kind, Mode: 0755, Linkname: entry.link, PAXRecords: entry.pax}
		if entry.kind == tar.TypeXGlobalHeader {
			h.Mode = 0
		}
		if entry.kind == tar.TypeReg || entry.kind == tar.TypeRegA {
			h.Size = int64(len(entry.body))
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(tw, entry.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func assertArchiveFile(t *testing.T, name, want string) {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil || string(data) != want {
		t.Fatalf("read %q = %q, %v; want %q", name, data, err, want)
	}
}

func assertEmptyExtraction(t *testing.T, parent string) {
	t.Helper()
	entries, err := os.ReadDir(parent)
	if err != nil || len(entries) != 0 {
		t.Fatalf("failed extraction left entries in %q: %v, %v", parent, entries, err)
	}
}

func TestArchiveExtractionValid(t *testing.T) {
	for _, first := range []string{"pkg/", "./", "file"} {
		t.Run(first, func(t *testing.T) {
			parent := t.TempDir()
			entries := []archiveEntry{{name: first, kind: tar.TypeDir}}
			wantRoot := filepath.Clean(first)
			if first == "file" {
				entries[0] = archiveEntry{name: "file", kind: tar.TypeReg, body: "first"}
				wantRoot = "."
			}
			entries = append(entries,
				archiveEntry{name: "pkg/sub/run", kind: tar.TypeReg, body: "original"},
				archiveEntry{name: "pkg/sub/run", kind: tar.TypeReg, body: "replacement"},
				archiveEntry{name: "pkg/sub/", kind: tar.TypeDir},
				archiveEntry{name: "pkg/link", kind: tar.TypeSymlink, link: "sub/run"},
			)
			root, err := untarGz(parent, bytes.NewReader(archiveData(t, entries...)))
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
			assertArchiveFile(t, filepath.Join(attempt, "pkg/sub/run"), "replacement")
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
		entries []archiveEntry
	}{
		{"parent path", []archiveEntry{{name: "../../outside/sentinel", kind: tar.TypeReg, body: "changed"}}},
		{"absolute path", []archiveEntry{{name: sentinel, kind: tar.TypeReg, body: "changed"}}},
		{"outside directory", []archiveEntry{{name: "../../new", kind: tar.TypeDir}}},
		{"outside symlink", []archiveEntry{{name: "../../new", kind: tar.TypeSymlink, link: sentinel}}},
		{"link parent", []archiveEntry{
			{name: "pkg/link", kind: tar.TypeSymlink, link: outside},
			{name: "pkg/link/sentinel", kind: tar.TypeReg, body: "changed"},
		}},
		{"link leaf", []archiveEntry{
			{name: "pkg/link", kind: tar.TypeSymlink, link: sentinel},
			{name: "pkg/link", kind: tar.TypeReg, body: "changed"},
		}},
		{"link directory", []archiveEntry{
			{name: "pkg/link", kind: tar.TypeSymlink, link: outside},
			{name: "pkg/link/", kind: tar.TypeDir},
		}},
		{"relative link", []archiveEntry{
			{name: "link", kind: tar.TypeSymlink, link: "../../outside"},
			{name: "link/sentinel", kind: tar.TypeReg, body: "changed"},
		}},
		{"dangling link", []archiveEntry{
			{name: "link", kind: tar.TypeSymlink, link: "missing"},
			{name: "link/new", kind: tar.TypeReg, body: "changed"},
		}},
		{"internal link", []archiveEntry{
			{name: "pkg/real/", kind: tar.TypeDir},
			{name: "pkg/link", kind: tar.TypeSymlink, link: "real"},
			{name: "pkg/link/new", kind: tar.TypeReg, body: "changed"},
		}},
		{"non-directory parent", []archiveEntry{
			{name: "file", kind: tar.TypeReg, body: "original"},
			{name: "file/new", kind: tar.TypeReg, body: "changed"},
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parent := filepath.Join(base, strings.ReplaceAll(tc.name, " ", "-"))
			if err := os.Mkdir(parent, 0755); err != nil {
				t.Fatal(err)
			}
			entries := append([]archiveEntry{{name: "pkg/partial", kind: tar.TypeReg, body: "partial"}}, tc.entries...)
			root, err := untarGz(parent, bytes.NewReader(archiveData(t, entries...)))
			last := tc.entries[len(tc.entries)-1].name
			if err == nil || root != "" || !strings.Contains(err.Error(), strings.TrimSuffix(last, "/")) {
				t.Fatalf("root = %q, error = %v; want rejected entry %q", root, err, last)
			}
			assertEmptyExtraction(t, parent)
			assertArchiveFile(t, sentinel, "preserved")
			if _, err := os.Lstat(filepath.Join(base, "new")); !os.IsNotExist(err) {
				t.Fatalf("outside entry created: %v", err)
			}
		})
	}
}

func TestArchiveExtractionFailureCleanup(t *testing.T) {
	valid := archiveData(t, archiveEntry{name: "pkg/partial", kind: tar.TypeReg, body: "partial"})
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
		"unsupported type": archiveData(t,
			archiveEntry{name: "pkg/partial", kind: tar.TypeReg, body: "partial"},
			archiveEntry{name: "pkg/hardlink", kind: tar.TypeLink, link: "pkg/partial"}),
		"directory over file": archiveData(t,
			archiveEntry{name: "pkg", kind: tar.TypeReg, body: "file"},
			archiveEntry{name: "pkg/", kind: tar.TypeDir}),
		"file over directory": archiveData(t,
			archiveEntry{name: "pkg/", kind: tar.TypeDir},
			archiveEntry{name: "pkg", kind: tar.TypeReg, body: "file"}),
		"symlink over file": archiveData(t,
			archiveEntry{name: "pkg", kind: tar.TypeReg, body: "file"},
			archiveEntry{name: "pkg", kind: tar.TypeSymlink, link: "other"}),
	} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			root, err := untarGz(parent, bytes.NewReader(data))
			if err == nil || root != "" {
				t.Fatalf("root = %q, error = %v; want failure", root, err)
			}
			assertEmptyExtraction(t, parent)
		})
	}
}

func TestArchiveExtractionCacheRetryIsIsolated(t *testing.T) {
	oldSrc := SrcDir
	SrcDir = t.TempDir()
	t.Cleanup(func() { SrcDir = oldSrc })
	parent := t.TempDir()
	valid := archiveData(t,
		archiveEntry{name: "pkg/", kind: tar.TypeDir},
		archiveEntry{name: "pkg/link/file", kind: tar.TypeReg, body: "downloaded"})
	requests := 0
	server := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write(valid)
	}))
	defer server.Close()
	url := server.URL + "/source.tar.gz"
	cache := filepath.Join(SrcDir, "source.tar.gz")
	bad := archiveData(t,
		archiveEntry{name: "pkg/link", kind: tar.TypeSymlink, link: t.TempDir()},
		archiveEntry{name: "pkg/link/file", kind: tar.TypeReg, body: "rejected"})
	if err := os.WriteFile(cache, bad, 0600); err != nil {
		t.Fatal(err)
	}
	cfg := &TargetConfig{Name: "lua"}
	item := &VerItem{Name: "source.tar.gz", Sum: archiveSHA256(bad)}
	if root, err := extractCachedFile(parent, url, cfg, item, false); err == nil || root != "" {
		t.Fatalf("cached extraction = %q, %v; want failure", root, err)
	}
	assertEmptyExtraction(t, parent)
	assertArchiveFile(t, cache, string(bad))
	if requests != 0 {
		t.Fatal("cached extraction unexpectedly downloaded")
	}
	item.Sum = archiveSHA256(valid)
	root, err := extractDownloadedFile(parent, url, cfg, item, false)
	if err != nil {
		t.Fatal(err)
	}
	assertArchiveFile(t, filepath.Join(root, "link/file"), "downloaded")
	assertArchiveFile(t, cache, string(valid))
	if requests != 1 {
		t.Fatalf("download requests = %d; want 1", requests)
	}
	// A successful cached attempt must also get a separate extraction tree.
	cachedRoot, err := extractCachedFile(parent, url, cfg, item, false)
	if err != nil || cachedRoot == root {
		t.Fatalf("cached retry root = %q, error = %v", cachedRoot, err)
	}
	assertArchiveFile(t, filepath.Join(cachedRoot, "link/file"), "downloaded")
	if requests != 1 {
		t.Fatal("successful cached extraction unexpectedly downloaded")
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
	root, err := untarGz(parent, bytes.NewReader(archiveData(t,
		archiveEntry{name: "pkg/", kind: tar.TypeDir},
		archiveEntry{name: "pkg/link", kind: tar.TypeSymlink, link: outside},
		archiveEntry{name: "pkg/LINK/sentinel", kind: tar.TypeReg, body: "separate"})))
	if caseSensitive {
		if err != nil {
			t.Fatal(err)
		}
		assertArchiveFile(t, filepath.Join(root, "LINK/sentinel"), "separate")
	} else {
		if err == nil || root != "" || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("case alias: root = %q, error = %v; want rejected symlink", root, err)
		}
		assertEmptyExtraction(t, parent)
	}
	assertArchiveFile(t, sentinel, "preserved")
}

func TestArchiveExtractionGlobalPAXMetadata(t *testing.T) {
	for _, name := range []string{"pkg/file", "../outside"} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			data := archiveData(t,
				archiveEntry{name: "pax_global_header", kind: tar.TypeXGlobalHeader, pax: map[string]string{"comment": "archive revision"}},
				archiveEntry{name: "pkg/", kind: tar.TypeDir},
				archiveEntry{name: name, kind: tar.TypeReg, body: "content"},
			)
			root, err := untarGz(parent, bytes.NewReader(data))
			if name == "../outside" {
				if err == nil {
					t.Fatal("accepted an entry outside the extraction directory")
				}
				assertEmptyExtraction(t, parent)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Base(root) != "pkg" {
				t.Fatalf("root = %q; want pkg directory", root)
			}
			assertArchiveFile(t, filepath.Join(root, "file"), "content")
			if _, err := os.Lstat(filepath.Join(filepath.Dir(root), "pax_global_header")); !os.IsNotExist(err) {
				t.Fatalf("global metadata became a filesystem entry: %v", err)
			}
		})
	}
}
