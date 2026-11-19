package install

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mah0x211/lenv/internal/cli"
	"github.com/mah0x211/lenv/internal/download"
)

var errGPGNotFound = errors.New("gpg, gpg2 and gpgv were not found; install GnuPG to verify LuaRocks signatures")

type signatureVerifier struct {
	Name   string
	Binary string
}

// LookPath also reports ErrNotFound for unusable entries. Only try the next
// command, or allow the missing-tool exception, when the entry is absent.
func findGPG() (*signatureVerifier, error) {
	for _, name := range []string{"gpg", "gpg2", "gpgv"} {
		binary, err := exec.LookPath(name)
		if err == nil {
			return &signatureVerifier{Name: name, Binary: binary}, nil
		}
		if !errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("cannot locate %s: %w", name, err)
		}
		for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
			file := filepath.Join(dir, name)
			if _, statErr := os.Lstat(file); statErr == nil {
				return nil, fmt.Errorf("%s entry %q exists but is not executable; repair or remove it", name, file)
			} else if !os.IsNotExist(statErr) {
				return nil, fmt.Errorf("cannot inspect %s entry %q: %w", name, file, statErr)
			}
		}
	}
	return nil, errGPGNotFound
}

// GitHub returns ASCII armor. Convert only its transport encoding to the
// binary keyring format accepted by both gpg and gpgv; GnuPG validates the keys.
func decodePublicKeys(data []byte) ([]byte, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	var keys []byte
	for strings.TrimSpace(text) != "" {
		text = strings.TrimSpace(text)
		const begin = "-----BEGIN PGP PUBLIC KEY BLOCK-----"
		const end = "-----END PGP PUBLIC KEY BLOCK-----"
		if !strings.HasPrefix(text, begin+"\n") {
			return nil, fmt.Errorf("invalid public key armor")
		}
		block, rest, ok := strings.Cut(strings.TrimPrefix(text, begin+"\n"), end)
		if !ok {
			return nil, fmt.Errorf("missing public key armor footer")
		}
		var body string
		// Armor headers, if present, end at the first empty line.
		if len(block) > 0 && block[0] == '\n' {
			body, ok = block[1:], true
		} else {
			_, body, ok = strings.Cut(block, "\n\n")
		}
		if !ok {
			return nil, fmt.Errorf("missing public key armor separator")
		}
		lines := strings.Fields(body)
		if len(lines) > 0 && strings.HasPrefix(lines[len(lines)-1], "=") {
			// The optional armor CRC is not an authentication mechanism.
			lines = lines[:len(lines)-1]
		}
		decoded, err := base64.StdEncoding.DecodeString(strings.Join(lines, ""))
		if err != nil || len(decoded) == 0 {
			return nil, fmt.Errorf("invalid public key armor body")
		}
		keys = append(keys, decoded...)
		text = rest
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("no public keys received")
	}
	return keys, nil
}

func verifyLuaRocksSignature(keyURL string, data, signature []byte) error {
	verifier, err := findGPG()
	if err != nil {
		return err
	}
	cli.Printf("download LuaRocks signing key from %s", keyURL)
	keyData, err := download.ReadDownload(keyURL)
	if err != nil {
		return fmt.Errorf("cannot download LuaRocks public key: %w", err)
	}
	key, err := decodePublicKeys(keyData)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "lenv-gpg-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	keyfile := filepath.Join(dir, "trustedkeys.gpg")
	sigfile := filepath.Join(dir, "archive.asc")
	datafile := filepath.Join(dir, "archive.tar.gz")
	for file, content := range map[string][]byte{keyfile: key, sigfile: signature, datafile: data} {
		if err := os.WriteFile(file, content, 0600); err != nil {
			return err
		}
	}
	args := []string{"--homedir", dir, "--keyring", keyfile, "--status-fd", "1"}
	if verifier.Name == "gpgv" {
		args = append(args, sigfile, datafile)
	} else {
		args = append([]string{"--no-options", "--batch", "--no-tty", "--no-default-keyring",
			"--no-auto-key-retrieve", "--no-autostart"}, args...)
		args = append(args, "--verify", sigfile, datafile)
	}
	cmd := exec.Command(verifier.Binary, args...)
	var status, diagnostic bytes.Buffer
	cmd.Stdout, cmd.Stderr = &status, &diagnostic
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s verification failed: %w: %s", verifier.Name, err, strings.TrimSpace(diagnostic.String()))
	}
	return checkGPGStatus(status.String())
}

// Require GnuPG's machine-readable valid-signature status and SHA-2, and reject
// reported validity errors. Only downloaded keys are in the dedicated keyring.
func checkGPGStatus(status string) error {
	valid := false
	for _, line := range strings.Split(status, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "[GNUPG:]" {
			continue
		}
		switch fields[1] {
		case "BADSIG", "ERRSIG", "EXPSIG", "EXPKEYSIG", "REVKEYSIG", "FAILURE", "NODATA":
			return fmt.Errorf("gpg rejected signature: %s", fields[1])
		case "VALIDSIG":
			if len(fields) < 11 {
				return fmt.Errorf("incomplete gpg VALIDSIG status")
			}
			switch fields[9] {
			case "8", "9", "10": // SHA-256, SHA-384, SHA-512.
				valid = true
			default:
				return fmt.Errorf("unsupported signature hash algorithm %q", fields[9])
			}
		}
	}
	if !valid {
		return fmt.Errorf("gpg did not report a valid LuaRocks signature")
	}
	return nil
}
