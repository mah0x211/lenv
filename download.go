package main

import (
	"fmt"
	"io"
	"net/http"
)

// Check both the initial request and every redirect before sending it.
func httpsRequest(req *http.Request) (*http.Response, error) {
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

func readDownload(address string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	rsp, err := httpsRequest(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get %q: %s", address, rsp.Status)
	}
	return io.ReadAll(rsp.Body)
}
