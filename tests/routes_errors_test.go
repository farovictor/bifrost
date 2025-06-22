package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/users"
	routes "github.com/farovictor/bifrost/routes"
)

func TestCreateKeyInvalidJSON(t *testing.T) {
	routes.KeyStore = keys.NewMemoryStore()
	routes.UserStore = users.NewMemoryStore()
	u := users.User{ID: "u", APIKey: "secret"}
	routes.UserStore.Create(u)
	router := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/v1/keys", strings.NewReader("{bad"))
	req.Header.Set("X-API-Key", u.APIKey)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	if body := strings.TrimSpace(rr.Body.String()); body != "invalid request" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestCreateKeyDuplicate(t *testing.T) {
	routes.KeyStore = keys.NewMemoryStore()
	routes.ServiceStore = services.NewMemoryStore()
	routes.UserStore = users.NewMemoryStore()
	u := users.User{ID: "u", APIKey: "secret"}
	routes.UserStore.Create(u)
	svc := services.Service{ID: "svc", Endpoint: "http://example.com", RootKeyID: "rk"}
	if err := routes.ServiceStore.Create(svc); err != nil {
		t.Fatalf("failed to seed service: %v", err)
	}
	k := keys.VirtualKey{ID: "dup", Scope: "read", Target: "svc", ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
	if err := routes.KeyStore.Create(k); err != nil {
		t.Fatalf("failed to seed store: %v", err)
	}

	body, _ := json.Marshal(k)
	router := setupRouter()
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader(body))
	req.Header.Set("X-API-Key", u.APIKey)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rr.Code)
	}

	if body := strings.TrimSpace(rr.Body.String()); body != "key already exists" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestDeleteKeyNotFound(t *testing.T) {
	routes.KeyStore = keys.NewMemoryStore()
	routes.UserStore = users.NewMemoryStore()
	u := users.User{ID: "u", APIKey: "secret"}
	routes.UserStore.Create(u)
	router := setupRouter()

	req := httptest.NewRequest(http.MethodDelete, "/v1/keys/unknown", nil)
	req.Header.Set("X-API-Key", u.APIKey)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}

	if body := strings.TrimSpace(rr.Body.String()); body != "not found" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestCreateKeyMissingService(t *testing.T) {
	routes.KeyStore = keys.NewMemoryStore()
	routes.ServiceStore = services.NewMemoryStore()
	routes.UserStore = users.NewMemoryStore()
	u := users.User{ID: "u", APIKey: "secret"}
	routes.UserStore.Create(u)
	router := setupRouter()

	k := keys.VirtualKey{ID: "nosvc", Scope: "read", Target: "missing", ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
	body, _ := json.Marshal(k)
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader(body))
	req.Header.Set("X-API-Key", u.APIKey)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}

	if body := strings.TrimSpace(rr.Body.String()); body != "service not found" {
		t.Fatalf("unexpected body: %s", body)
	}

	if _, err := routes.KeyStore.Get(k.ID); err != keys.ErrKeyNotFound {
		t.Fatalf("key should not have been created")
	}
}

func TestCreateKeyInvalidScope(t *testing.T) {
	routes.KeyStore = keys.NewMemoryStore()
	routes.ServiceStore = services.NewMemoryStore()
	routes.RootKeyStore = rootkeys.NewMemoryStore()
	routes.UserStore = users.NewMemoryStore()
	u := users.User{ID: "u", APIKey: "secret"}
	routes.UserStore.Create(u)
	rk := rootkeys.RootKey{ID: "rk-scope", APIKey: "k"}
	if err := routes.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}
	svc := services.Service{ID: "svc-scope", Endpoint: "http://example.com", RootKeyID: rk.ID}
	if err := routes.ServiceStore.Create(svc); err != nil {
		t.Fatalf("failed to seed service: %v", err)
	}
	router := setupRouter()

	k := keys.VirtualKey{ID: "badscope", Scope: "unknown", Target: svc.ID, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
	body, _ := json.Marshal(k)
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader(body))
	req.Header.Set("X-API-Key", u.APIKey)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if body := strings.TrimSpace(rr.Body.String()); body != "invalid scope" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestCreateKeyEmptyScope(t *testing.T) {
	routes.KeyStore = keys.NewMemoryStore()
	routes.ServiceStore = services.NewMemoryStore()
	routes.RootKeyStore = rootkeys.NewMemoryStore()
	routes.UserStore = users.NewMemoryStore()
	u := users.User{ID: "u", APIKey: "secret"}
	routes.UserStore.Create(u)
	rk := rootkeys.RootKey{ID: "rk-empty", APIKey: "k"}
	if err := routes.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}
	svc := services.Service{ID: "svc-empty", Endpoint: "http://example.com", RootKeyID: rk.ID}
	if err := routes.ServiceStore.Create(svc); err != nil {
		t.Fatalf("failed to seed service: %v", err)
	}
	router := setupRouter()

	k := keys.VirtualKey{ID: "emptyscope", Scope: "", Target: svc.ID, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
	body, _ := json.Marshal(k)
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader(body))
	req.Header.Set("X-API-Key", u.APIKey)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if body := strings.TrimSpace(rr.Body.String()); body != "invalid scope" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestCreateKeyPastExpiration(t *testing.T) {
	routes.KeyStore = keys.NewMemoryStore()
	routes.ServiceStore = services.NewMemoryStore()
	routes.RootKeyStore = rootkeys.NewMemoryStore()
	routes.UserStore = users.NewMemoryStore()
	u := users.User{ID: "u", APIKey: "secret"}
	routes.UserStore.Create(u)
	rk := rootkeys.RootKey{ID: "rk-exp", APIKey: "k"}
	if err := routes.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}
	svc := services.Service{ID: "svc-exp", Endpoint: "http://example.com", RootKeyID: rk.ID}
	if err := routes.ServiceStore.Create(svc); err != nil {
		t.Fatalf("failed to seed service: %v", err)
	}
	router := setupRouter()

	k := keys.VirtualKey{ID: "expired", Scope: "read", Target: svc.ID, ExpiresAt: time.Now().Add(-time.Hour), RateLimit: 1}
	body, _ := json.Marshal(k)
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader(body))
	req.Header.Set("X-API-Key", u.APIKey)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if body := strings.TrimSpace(rr.Body.String()); body != "expires_at must be in the future" {
		t.Fatalf("unexpected body: %s", body)
	}
}
