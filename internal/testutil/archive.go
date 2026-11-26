package testutil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"testing"
)

type ArchiveEntry struct {
	Name, Body, Link string
	Kind             byte
	PAX              map[string]string
}

func ArchiveData(t testing.TB, entries ...ArchiveEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, entry := range entries {
		h := &tar.Header{Name: entry.Name, Typeflag: entry.Kind, Mode: 0755, Linkname: entry.Link, PAXRecords: entry.PAX}
		if entry.Kind == tar.TypeXGlobalHeader {
			h.Mode = 0
		}
		if entry.Kind == tar.TypeReg || entry.Kind == tar.TypeRegA {
			h.Size = int64(len(entry.Body))
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(tw, entry.Body); err != nil {
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
