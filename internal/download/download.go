package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mah0x211/lenv/internal/fs"
)

// Check both the initial request and every redirect before sending it.
func HttpsRequest(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" {
		return nil, fmt.Errorf("refusing non-HTTPS download %q", req.URL.String())
	}
	client := *http.DefaultClient
	previous := client.CheckRedirect
	client.CheckRedirect = func(next *http.Request, via []*http.Request) error {
		if next.URL.Scheme != "https" {
			return fmt.Errorf("refusing non-HTTPS redirect to %q", next.URL.String())
		}
		if previous != nil {
			return previous(next, via)
		}
		if len(via) >= 10 {
			return fmt.Errorf("stopped after 10 redirects")
		}
		return nil
	}
	return client.Do(req)
}

func ReadDownload(address string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	rsp, err := HttpsRequest(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get %q: %s", address, rsp.Status)
	}
	return io.ReadAll(rsp.Body)
}

// Publish a complete archive without truncating the previous download.
func SaveDownloadedArchive(file string, data []byte) error {
	if err := fs.Mkdir(filepath.Dir(file)); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(file), ".lenv-download-*")
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), file)
}
