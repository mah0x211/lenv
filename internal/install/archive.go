package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/config"
	"github.com/mah0x211/lenv/internal/fs"
	version "github.com/mah0x211/lenv/internal/versions"
)

func verify(cfg *config.TargetConfig, item *version.VerItem, data, signature []byte) error {
	switch cfg.Name {
	case "lua":
		if !strings.HasPrefix(item.Sum, "sha256:") {
			return fmt.Errorf("missing SHA-256 for %q; run lenv fetch", item.Name)
		}
		expected, err := hex.DecodeString(strings.TrimPrefix(item.Sum, "sha256:"))
		if err != nil || len(expected) != sha256.Size {
			return fmt.Errorf("invalid SHA-256 for %q; run lenv fetch", item.Name)
		}
		actual := sha256.Sum256(data)
		if !bytes.Equal(actual[:], expected) {
			return fmt.Errorf("SHA-256 mismatch for %q: got %x, want %x", item.Name, actual, expected)
		}
	case "luarocks":
		if item.Sum != "pgp:"+item.Name+".asc" {
			return fmt.Errorf("invalid PGP signature reference for %q; run lenv fetch", item.Name)
		}
		if err := verifyLuaRocksSignature(cfg.SigningKeyURL, data, signature); err != nil {
			return fmt.Errorf("PGP verification failed for %q: %w", item.Name, err)
		}
	case "luajit":
		// GitHub HTTPS is the trust boundary; there is no publisher checksum.
	default:
		return fmt.Errorf("unsupported archive target %q", cfg.Name)
	}
	return nil
}

// archivePath checks the actual extraction tree as well as the entry name.
// The tree is private to this extraction attempt; archive-created symlinks must
// never be followed, including when an entry replaces an existing leaf.
func archivePath(dir, name string) (string, error) {
	if !filepath.IsLocal(name) {
		return "", fmt.Errorf("archive entry %q is not a local path", name)
	}

	clean := filepath.Clean(name)
	dest := filepath.Join(dir, clean)
	current := dir
	for _, part := range strings.Split(clean, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return dest, nil
		} else if err != nil {
			return "", fmt.Errorf("archive entry %q: %w", name, err)
		} else if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("archive entry %q uses symlink %q", name, current)
		} else if current != dest && !info.IsDir() {
			return "", fmt.Errorf("archive entry %q uses non-directory %q", name, current)
		}
	}
	return dest, nil
}

func untarGz(dir string, data io.Reader) (rootdir string, err error) {
	gz, err := gzip.NewReader(data)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	// A failed cached extraction must not contaminate a subsequent download.
	dir, err = os.MkdirTemp(dir, "extract-")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			if cleanupErr := os.RemoveAll(dir); cleanupErr != nil {
				err = fmt.Errorf("%w; failed to remove extraction directory %q: %v", err, dir, cleanupErr)
			}
		}
	}()

	rootdir = dir
	checked := false
	r := tar.NewReader(gz)
	for {
		h, readErr := r.Next()
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				if _, err := io.Copy(io.Discard, gz); err != nil {
					return "", fmt.Errorf("invalid gzip trailer: %w", err)
				}
				break
			}
			return "", readErr
		}

		// Global PAX headers describe metadata, not filesystem entries.
		if h.Typeflag == tar.TypeXGlobalHeader {
			continue
		}

		dest, pathErr := archivePath(dir, h.Name)
		if pathErr != nil {
			return "", pathErr
		}
		cli.Println(h.Name)
		switch h.Typeflag {
		case tar.TypeReg, tar.TypeRegA:
			err = fs.WriteFile(dest, h.FileInfo().Mode().Perm(), r)
		case tar.TypeSymlink:
			if err = fs.Mkdir(filepath.Dir(dest)); err == nil {
				err = os.Symlink(h.Linkname, dest)
			}
		case tar.TypeDir:
			if !checked {
				rootdir = dest
			}
			err = fs.Mkdir(dest)
		default:
			err = fmt.Errorf("unknown type flag %v", h.Typeflag)
		}
		checked = true

		if err != nil {
			return "", fmt.Errorf("archive entry %q: %w", h.Name, err)
		}
	}
	return rootdir, nil
}
