package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
)

func TestListKeys(t *testing.T) {
	env := newTestEnv(t)
	rk := rootkeys.RootKey{ID: "rk", APIKey: "k"}
	env.Server.RootKeyStore.Create(rk)
	svc := services.Service{ID: "svc", Endpoint: "http://example.com", RootKeyID: rk.ID}
	env.Server.ServiceStore.Create(svc)
	k1 := keys.VirtualKey{ID: "k1", Scope: keys.ScopeRead, Target: svc.ID, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 10}
	k2 := keys.VirtualKey{ID: "k2", Scope: keys.ScopeWrite, Target: svc.ID, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 5}
	env.Server.KeyStore.Create(k1)
	env.Server.KeyStore.Create(k2)

	req := httptest.NewRequest(http.MethodGet, "/v1/keys", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp []keys.VirtualKey
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(resp))
	}
}

func TestListRootKeys(t *testing.T) {
	env := newTestEnv(t)
	env.Server.RootKeyStore.Create(rootkeys.RootKey{ID: "rk1", APIKey: "a"})
	env.Server.RootKeyStore.Create(rootkeys.RootKey{ID: "rk2", APIKey: "b"})

	req := httptest.NewRequest(http.MethodGet, "/v1/rootkeys", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp []rootkeys.RootKey
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 root keys, got %d", len(resp))
	}
}

func TestListServices(t *testing.T) {
	env := newTestEnv(t)
	rk := rootkeys.RootKey{ID: "rk", APIKey: "k"}
	env.Server.RootKeyStore.Create(rk)
	env.Server.ServiceStore.Create(services.Service{ID: "s1", Endpoint: "http://a.com", RootKeyID: rk.ID})
	env.Server.ServiceStore.Create(services.Service{ID: "s2", Endpoint: "http://b.com", RootKeyID: rk.ID})

	req := httptest.NewRequest(http.MethodGet, "/v1/services", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp []services.Service
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 services, got %d", len(resp))
	}
}
