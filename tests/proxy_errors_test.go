package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
)

func TestProxyMissingKey(t *testing.T) {
	router := setupRouter(newTestServer(t))
	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/backend", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "missing key" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestProxyInvalidKey(t *testing.T) {
	router := setupRouter(newTestServer(t))
	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/backend", nil)
	req.Header.Set("X-Virtual-Key", "bad")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "invalid key" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestProxyExpiredKey(t *testing.T) {
	s := newTestServer(t)
	k := keys.VirtualKey{ID: "expired", Scope: keys.ScopeRead, Target: "svc", ExpiresAt: time.Now().Add(-time.Hour), RateLimit: 100}
	if err := s.KeyStore.Create(k); err != nil {
		t.Fatalf("seed key: %v", err)
	}

	router := setupRouter(s)
	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/backend", nil)
	req.Header.Set("X-Virtual-Key", k.ID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "key expired" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestProxyScopeViolation(t *testing.T) {
	s := newTestServer(t)

	called := false
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		io.WriteString(w, "ok")
	}))
	defer backend.Close()

	rk := rootkeys.RootKey{ID: "rk", APIKey: "real"}
	if err := s.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}
	svc := services.Service{ID: "svc", Endpoint: backend.URL, RootKeyID: rk.ID}
	if err := s.ServiceStore.Create(svc); err != nil {
		t.Fatalf("seed service: %v", err)
	}
	k := keys.VirtualKey{ID: "vkey", Target: svc.ID, Scope: keys.ScopeRead, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 100}
	if err := s.KeyStore.Create(k); err != nil {
		t.Fatalf("seed key: %v", err)
	}

	router := setupRouter(s)
	req := httptest.NewRequest(http.MethodPost, "/v1/proxy/backend", nil)
	req.Header.Set("X-Virtual-Key", k.ID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "insufficient scope" {
		t.Fatalf("unexpected error: %s", msg)
	}
	if called {
		t.Fatalf("backend should not be called")
	}
}
