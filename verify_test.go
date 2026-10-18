package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

func archiveSHA256(data []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(data))
}

func officialLuaRocksArchive(t *testing.T) ([]byte, []byte) {
	t.Helper()
	// Existing tests run from .codex to isolate legacy relative-path output.
	var data, signature []byte
	var err error
	for _, dir := range []string{"testdata/", "../testdata/"} {
		data, err = os.ReadFile(dir + "luarocks-3.13.0.tar.gz")
		if err == nil {
			signature, err = os.ReadFile(dir + "luarocks-3.13.0.tar.gz.asc")
			break
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	return data, signature
}

func TestVerifyLuaArchive(t *testing.T) {
	data := []byte("archive")
	cfg := &TargetConfig{Name: "lua"}
	for _, sum := range []string{archiveSHA256(data), strings.ToUpper(strings.TrimPrefix(archiveSHA256(data), "sha256:"))} {
		if !strings.HasPrefix(sum, "sha256:") {
			sum = "sha256:" + sum
		}
		if err := verifyArchive(cfg, &VerItem{Name: "lua.tar.gz", Sum: sum}, data, nil); err != nil {
			t.Fatal(err)
		}
	}
	for _, sum := range []string{"", "pgp:lua.tar.gz.asc", "sha256:invalid", "sha256:00", archiveSHA256([]byte("altered"))} {
		if err := verifyArchive(cfg, &VerItem{Name: "lua.tar.gz", Sum: sum}, data, nil); err == nil {
			t.Fatalf("accepted invalid checksum %q", sum)
		}
	}
}

func TestVerifyLuaRocksOfficialArchive(t *testing.T) {
	data, signature := officialLuaRocksArchive(t)
	cfg := &TargetConfig{Name: "luarocks", SigningKeyURL: publisherKeyURL(t)}
	item := &VerItem{Name: "luarocks-3.13.0.tar.gz", Sum: "pgp:luarocks-3.13.0.tar.gz.asc"}
	if err := verifyArchive(cfg, item, data, signature); err != nil {
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
			if err := verifyArchive(cfg, item, inputs[0], inputs[1]); err == nil {
				t.Fatal("unverified archive accepted")
			}
		})
	}
	item.Sum = "pgp:../unrelated.asc"
	if err := verifyArchive(cfg, item, data, signature); err == nil {
		t.Fatal("accepted unrelated signature reference")
	}
}

func TestVerifyLuaRocksRejectsOtherSigner(t *testing.T) {
	var signature []byte
	var err error
	for _, name := range []string{"testdata/untrusted-signature.asc", "../testdata/untrusted-signature.asc"} {
		signature, err = os.ReadFile(name)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	item := &VerItem{Name: "rocks.tar.gz", Sum: "pgp:rocks.tar.gz.asc"}
	if err := verifyArchive(&TargetConfig{Name: "luarocks", SigningKeyURL: publisherKeyURL(t)}, item, []byte("archive"), signature); err == nil {
		t.Fatal("accepted an untrusted signing key")
	}
}

func TestVerifyArchiveTargetPolicy(t *testing.T) {
	if err := verifyArchive(&TargetConfig{Name: "luajit"}, &VerItem{}, nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := verifyArchive(&TargetConfig{Name: "unknown"}, &VerItem{}, nil, nil); err == nil {
		t.Fatal("accepted unknown verification policy")
	}
}
