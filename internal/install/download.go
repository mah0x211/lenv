package install

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/download"
	version "github.com/mah0x211/lenv/internal/versions"
)

func extractDownloadedFile(tmpdir, url string, cfg *config.TargetConfig, item *version.VerItem, skipPGP bool) (string, error) {
	data, err := download.ReadDownload(url)
	if err != nil {
		return "", fmt.Errorf("download error: %w", err)
	}
	var signature []byte
	if cfg.Name == "luarocks" && !skipPGP {
		signature, err = download.ReadDownload(url + ".asc")
		if err != nil {
			return "", fmt.Errorf("signature download error: %w", err)
		}
	}
	if cfg.Name != "luarocks" || !skipPGP {
		if err := verify(cfg, item, data, signature); err != nil {
			return "", err
		}
	}
	file := filepath.Join(config.SrcDir, filepath.Base(url))
	if err := download.SaveDownloadedArchive(file, data); err != nil {
		return "", err
	}
	if cfg.Name == "luarocks" && !skipPGP {
		if err := download.SaveDownloadedArchive(file+".asc", signature); err != nil {
			return "", err
		}
	}
	return untarGz(tmpdir, bytes.NewReader(data))
}
