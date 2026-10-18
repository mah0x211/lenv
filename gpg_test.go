package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGPGMissingToolFails(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := verifyLuaRocksSignature("", []byte("archive"), nil); !errors.Is(err, errGPGNotFound) {
		t.Fatalf("missing GPG = %v", err)
	}
}

func TestGPGStatusPolicy(t *testing.T) {
	fpr := "8460980B2B79786DE0C7FCC83FD8F43C2BB3C478"
	valid := "[GNUPG:] VALIDSIG " + fpr + " 2026-01-01 1767225600 0 4 0 1 8 00 " + fpr
	for name, status := range map[string]string{
		"valid":            valid,
		"unknown metadata": "[GNUPG:] FUTURE_FIELD value\n" + valid,
		"rotated key":      strings.ReplaceAll(valid, fpr, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		"subkey":           strings.Replace(valid, "VALIDSIG "+fpr, "VALIDSIG AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if err := checkGPGStatus(status); err != nil {
				t.Fatal(err)
			}
		})
	}
	for name, status := range map[string]string{
		"missing":           "",
		"only goodsig":      "[GNUPG:] GOODSIG 3FD8F43C2BB3C478 signer",
		"short":             "[GNUPG:] VALIDSIG",
		"sha1":              strings.Replace(valid, " 1 8 00 ", " 1 2 00 ", 1),
		"bad":               valid + "\n[GNUPG:] BADSIG key signer",
		"expired signature": valid + "\n[GNUPG:] EXPSIG key signer",
		"expired key":       valid + "\n[GNUPG:] EXPKEYSIG key signer",
		"revoked":           valid + "\n[GNUPG:] REVKEYSIG key signer",
		"error":             valid + "\n[GNUPG:] ERRSIG key",
	} {
		t.Run(name, func(t *testing.T) {
			if err := checkGPGStatus(status); err == nil {
				t.Fatal("invalid verification status accepted")
			}
		})
	}
}

func TestGPGDoesNotUseOrChangeUserKeyring(t *testing.T) {
	data, signature := officialLuaRocksArchive(t)
	home := t.TempDir()
	configuration := filepath.Join(home, "gpg.conf")
	if err := os.WriteFile(configuration, []byte("this-option-is-invalid\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GNUPGHOME", home)
	temp := t.TempDir()
	t.Setenv("TMPDIR", temp)
	if err := verifyLuaRocksSignature(publisherKeyURL(t), data, signature); err != nil {
		t.Fatal(err)
	}
	assertArchiveFile(t, configuration, "this-option-is-invalid\n")
	entries, err := os.ReadDir(home)
	if err != nil || len(entries) != 1 {
		t.Fatalf("user GPG directory changed: %v, %v", entries, err)
	}
	assertEmptyExtraction(t, temp)
}
