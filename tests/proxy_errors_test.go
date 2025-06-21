package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FokusInternal/bifrost/pkg/keys"
	"github.com/FokusInternal/bifrost/pkg/rootkeys"
	"github.com/FokusInternal/bifrost/pkg/services"
	routes "github.com/FokusInternal/bifrost/routes"
)

func TestProxyMissingKey(t *testing.T) {
	routes.ServiceStore = services.NewStore()
	routes.KeyStore = keys.NewStore()
	routes.RootKeyStore = rootkeys.NewStore()

	router := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/backend", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if body := strings.TrimSpace(rr.Body.String()); body != "missing key" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestProxyInvalidKey(t *testing.T) {
	routes.ServiceStore = services.NewStore()
	routes.KeyStore = keys.NewStore()
	routes.RootKeyStore = rootkeys.NewStore()

	router := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/backend", nil)
	req.Header.Set("X-Virtual-Key", "bad")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if body := strings.TrimSpace(rr.Body.String()); body != "invalid key" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestProxyExpiredKey(t *testing.T) {
	routes.ServiceStore = services.NewStore()
	routes.KeyStore = keys.NewStore()
	routes.RootKeyStore = rootkeys.NewStore()

	k := keys.VirtualKey{ID: "expired", Scope: keys.ScopeRead, Target: "svc", ExpiresAt: time.Now().Add(-time.Hour), RateLimit: 1}
	if err := routes.KeyStore.Create(k); err != nil {
		t.Fatalf("seed key: %v", err)
	}

	router := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/backend", nil)
	req.Header.Set("X-Virtual-Key", k.ID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if body := strings.TrimSpace(rr.Body.String()); body != "key expired" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestProxyScopeViolation(t *testing.T) {
	routes.ServiceStore = services.NewStore()
	routes.KeyStore = keys.NewStore()
	routes.RootKeyStore = rootkeys.NewStore()

	called := false
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		io.WriteString(w, "ok")
	}))
	defer backend.Close()

	rk := rootkeys.RootKey{ID: "rk", APIKey: "real"}
	if err := routes.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}
	svc := services.Service{ID: "svc", Endpoint: backend.URL, RootKeyID: rk.ID}
	if err := routes.ServiceStore.Create(svc); err != nil {
		t.Fatalf("seed service: %v", err)
	}
	k := keys.VirtualKey{ID: "vkey", Target: svc.ID, Scope: keys.ScopeRead, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
	if err := routes.KeyStore.Create(k); err != nil {
		t.Fatalf("seed key: %v", err)
	}

	router := setupRouter()
	req := httptest.NewRequest(http.MethodPost, "/v1/proxy/backend", nil)
	req.Header.Set("X-Virtual-Key", k.ID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	if body := strings.TrimSpace(rr.Body.String()); body != "insufficient scope" {
		t.Fatalf("unexpected body: %s", body)
	}
	if called {
		t.Fatalf("backend should not be called")
	}
}
