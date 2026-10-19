package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLuaRocksReleaseParserRejectsUnpairedAndUnsupportedReleases(t *testing.T) {
	for _, html := range []string{
		`<a href="luarocks-3.13.0.tar.gz">archive</a><a href="luarocks-3.12.2.tar.gz.asc">signature</a>`,
		`<a href="luarocks-latest.tar.gz">archive</a><a href="luarocks-latest.tar.gz.asc">signature</a>`,
	} {
		list, _ := parseLuaRocksVers([]byte(html)).GetList()
		require.Empty(t, list, "an invalid release must not be published")
	}
}

func TestPublicKeyArmorRequiresHeaderSeparator(t *testing.T) {
	_, err := decodePublicKeys([]byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: test\nYWJj\n-----END PGP PUBLIC KEY BLOCK-----"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing public key armor separator")
}

func TestPublicKeyArmorRequiresClosingFooter(t *testing.T) {
	_, err := decodePublicKeys([]byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nYWJj\n"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing public key armor footer")
}
