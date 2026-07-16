package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthRejectsNon2xxWithoutReturningRawBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "secret upstream body", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	_, err := New(server.URL, "").Health()
	if err == nil || strings.Contains(err.Error(), "secret upstream body") {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestHealthHonorsHTTPTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()
	c := New(server.URL, "")
	c.http.Timeout = 10 * time.Millisecond
	if _, err := c.Health(); err == nil {
		t.Fatal("expected timeout")
	}
}

func TestResolveNetworkEscapesEveryPathValue(t *testing.T) {
	var path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"wallet_address":"ok"}`))
	}))
	defer server.Close()
	if _, err := New(server.URL, "").ResolveNetwork("alice/bob", "e/../../admin"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(path, "/admin") || !strings.Contains(path, "%2F") {
		t.Fatalf("unsafe path %q", path)
	}
}
