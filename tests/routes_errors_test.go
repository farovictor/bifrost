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
)

func TestCreateKeyInvalidJSON(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", strings.NewReader("{bad"))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "invalid request" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestCreateKeyDuplicate(t *testing.T) {
	env := newTestEnv(t)
	svc := services.Service{ID: "svc", Endpoint: "http://example.com", RootKeyID: "rk"}
	if err := env.Server.ServiceStore.Create(svc); err != nil {
		t.Fatalf("failed to seed service: %v", err)
	}
	k := keys.VirtualKey{ID: "dup", Scope: "read", Target: "svc", ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
	if err := env.Server.KeyStore.Create(k); err != nil {
		t.Fatalf("failed to seed store: %v", err)
	}

	body, _ := json.Marshal(k)
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "key already exists" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestDeleteKeyNotFound(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodDelete, "/v1/keys/unknown", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "not found" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestCreateKeyMissingService(t *testing.T) {
	env := newTestEnv(t)
	k := keys.VirtualKey{ID: "nosvc", Scope: "read", Target: "missing", ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
	body, _ := json.Marshal(k)
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "service not found" {
		t.Fatalf("unexpected error: %s", msg)
	}
	if _, err := env.Server.KeyStore.Get(k.ID); err != keys.ErrKeyNotFound {
		t.Fatalf("key should not have been created")
	}
}

func TestCreateKeyInvalidScope(t *testing.T) {
	env := newTestEnv(t)
	rk := rootkeys.RootKey{ID: "rk-scope", APIKey: "k"}
	if err := env.Server.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}
	svc := services.Service{ID: "svc-scope", Endpoint: "http://example.com", RootKeyID: rk.ID}
	if err := env.Server.ServiceStore.Create(svc); err != nil {
		t.Fatalf("failed to seed service: %v", err)
	}

	k := keys.VirtualKey{ID: "badscope", Scope: "unknown", Target: svc.ID, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
	body, _ := json.Marshal(k)
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "invalid scope" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestCreateKeyEmptyScope(t *testing.T) {
	env := newTestEnv(t)
	rk := rootkeys.RootKey{ID: "rk-empty", APIKey: "k"}
	if err := env.Server.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}
	svc := services.Service{ID: "svc-empty", Endpoint: "http://example.com", RootKeyID: rk.ID}
	if err := env.Server.ServiceStore.Create(svc); err != nil {
		t.Fatalf("failed to seed service: %v", err)
	}

	k := keys.VirtualKey{ID: "emptyscope", Scope: "", Target: svc.ID, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
	body, _ := json.Marshal(k)
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "invalid scope" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestCreateKeyPastExpiration(t *testing.T) {
	env := newTestEnv(t)
	rk := rootkeys.RootKey{ID: "rk-exp", APIKey: "k"}
	if err := env.Server.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}
	svc := services.Service{ID: "svc-exp", Endpoint: "http://example.com", RootKeyID: rk.ID}
	if err := env.Server.ServiceStore.Create(svc); err != nil {
		t.Fatalf("failed to seed service: %v", err)
	}

	k := keys.VirtualKey{ID: "expired", Scope: "read", Target: svc.ID, ExpiresAt: time.Now().Add(-time.Hour), RateLimit: 1}
	body, _ := json.Marshal(k)
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "expires_at must be in the future" {
		t.Fatalf("unexpected error: %s", msg)
	}
}
