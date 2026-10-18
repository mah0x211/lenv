package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newDownloadTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	old := http.DefaultClient
	http.DefaultClient = server.Client()
	t.Cleanup(func() {
		http.DefaultClient = old
		server.Close()
	})
	return server
}

func TestHTTPSRejectsPlaintextAndDowngrade(t *testing.T) {
	hits := 0
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))
	defer plain.Close()
	if _, err := readDownload(plain.URL); err == nil {
		t.Fatal("plaintext download accepted")
	}
	secure := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, plain.URL, http.StatusFound)
	}))
	if _, err := readDownload(secure.URL); err == nil {
		t.Fatal("HTTPS downgrade accepted")
	}
	if hits != 0 {
		t.Fatal("insecure endpoint was contacted")
	}
}

func TestHTTPSRedirectAndCertificateValidation(t *testing.T) {
	secure := newDownloadTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start":
			http.Redirect(w, r, "/end", http.StatusFound)
		case "/loop":
			http.Redirect(w, r, "/loop", http.StatusFound)
		default:
			_, _ = w.Write([]byte("secure"))
		}
	}))
	if data, err := readDownload(secure.URL + "/start"); err != nil || string(data) != "secure" {
		t.Fatalf("HTTPS redirect = %q, %v", data, err)
	}
	if _, err := readDownload(secure.URL + "/loop"); err == nil {
		t.Fatal("redirect loop accepted")
	}
	trusted := http.DefaultClient
	http.DefaultClient = &http.Client{}
	defer func() { http.DefaultClient = trusted }()
	if _, err := readDownload(secure.URL); err == nil {
		t.Fatal("untrusted TLS certificate accepted")
	}
	if _, err := readDownload("%"); err == nil {
		t.Fatal("invalid URL accepted")
	}
}
