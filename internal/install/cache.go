package install

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/fs"
	version "github.com/mah0x211/lenv/internal/versions"
)

func openCachedFile(url string) ([]byte, error) {
	file := filepath.Join(config.SrcDir, filepath.Base(url))
	f, err := fs.OpenFile(file)
	if err != nil || f == nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func extractCachedFile(tmpdir, url string, cfg *config.TargetConfig, item *version.VerItem, skipPGP bool) (string, error) {
	data, err := openCachedFile(url)
	if err != nil {
		return "", fmt.Errorf("open cached file error: %w", err)
	} else if data == nil {
		return "", nil

	}
	var signature []byte
	if cfg.Name == "luarocks" && !skipPGP {
		signature, err = os.ReadFile(filepath.Join(config.SrcDir, filepath.Base(url)) + ".asc")
		if err != nil {
			return "", fmt.Errorf("read cached signature: %w", err)
		}
	}
	if cfg.Name != "luarocks" || !skipPGP {
		if err := verify(cfg, item, data, signature); err != nil {
			return "", err
		}
	}
	return untarGz(tmpdir, bytes.NewReader(data))
}
