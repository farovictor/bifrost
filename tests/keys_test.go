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

func TestCreateKey(t *testing.T) {
	routes.KeyStore = keys.NewMemoryStore()
	routes.ServiceStore = services.NewMemoryStore()
	routes.RootKeyStore = rootkeys.NewMemoryStore()
	routes.UserStore = users.NewMemoryStore()
	u := users.User{ID: "u", Name: "U", Email: "u@example.com", APIKey: "secret"}
	routes.UserStore.Create(u)
	rk := rootkeys.RootKey{ID: "rk", APIKey: "k"}
	if err := routes.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}
	svc := services.Service{ID: "svc", Endpoint: "http://example.com", RootKeyID: rk.ID}
	if err := routes.ServiceStore.Create(svc); err != nil {
		t.Fatalf("failed to seed service: %v", err)
	}
	router := setupRouter()

	k := keys.VirtualKey{ID: "abc", Scope: "read", Target: svc.ID, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
	body, _ := json.Marshal(k)
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader(body))
	req.Header.Set("X-API-Key", u.APIKey)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var resp keys.VirtualKey
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != k.ID {
		t.Fatalf("expected ID %s, got %s", k.ID, resp.ID)
	}
}

func TestDeleteKey(t *testing.T) {
	routes.KeyStore = keys.NewMemoryStore()
	routes.UserStore = users.NewMemoryStore()
	u := users.User{ID: "u", Name: "U", Email: "u@example.com", APIKey: "secret"}
	routes.UserStore.Create(u)
	k := keys.VirtualKey{ID: "dead", Scope: "x", Target: "svc", ExpiresAt: time.Now(), RateLimit: 1}
	if err := routes.KeyStore.Create(k); err != nil {
		t.Fatalf("failed to seed store: %v", err)
	}

	router := setupRouter()
	req := httptest.NewRequest(http.MethodDelete, "/v1/keys/"+k.ID, nil)
	req.Header.Set("X-API-Key", u.APIKey)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}

	if _, err := routes.KeyStore.Get(k.ID); err != keys.ErrKeyNotFound {
		t.Fatalf("key was not deleted")
	}
}

func TestCreateKeyExampleJSON(t *testing.T) {
	routes.KeyStore = keys.NewMemoryStore()
	routes.ServiceStore = services.NewMemoryStore()
	routes.RootKeyStore = rootkeys.NewMemoryStore()
	routes.UserStore = users.NewMemoryStore()
	u := users.User{ID: "u", Name: "U", Email: "u@example.com", APIKey: "secret"}
	routes.UserStore.Create(u)
	rk := rootkeys.RootKey{ID: "rk2", APIKey: "k"}
	if err := routes.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}
	svc := services.Service{ID: "svc", Endpoint: "http://example.com", RootKeyID: rk.ID}
	if err := routes.ServiceStore.Create(svc); err != nil {
		t.Fatalf("failed to seed service: %v", err)
	}
	router := setupRouter()

	payload := `{"id":"jsonex","scope":"read","target":"svc","expires_at":"2050-01-02T15:04:05Z","rate_limit":1}`
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", strings.NewReader(payload))
	req.Header.Set("X-API-Key", u.APIKey)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var resp keys.VirtualKey
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expTime, _ := time.Parse(time.RFC3339, "2050-01-02T15:04:05Z")
	if resp.ID != "jsonex" || !resp.ExpiresAt.Equal(expTime) || resp.Scope != "read" || resp.Target != "svc" || resp.RateLimit != 1 {
		t.Fatalf("unexpected response: %#v", resp)
	}
}
