package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func NewDownloadTestServer(t testing.TB, handler http.Handler) *httptest.Server {
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
