package fetch

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
