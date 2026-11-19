package target

import (
	"testing"

	"github.com/mah0x211/lenv/internal/config"
	version "github.com/mah0x211/lenv/internal/versions"
	"github.com/stretchr/testify/require"
)

func TestLuaJitBranchURLPreservesName(t *testing.T) {
	cfg := &config.TargetConfig{Name: "luajit", DownloadURL: "https://example.invalid/archive/refs/heads/"}
	item := &version.VerItem{Name: "v2.1", Remote: "origin"}
	require.Equal(t, cfg.DownloadURL+"v2.1.tar.gz", SourceURL(cfg, item))
}
