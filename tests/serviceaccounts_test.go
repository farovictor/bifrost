package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/serviceaccounts"
	"github.com/farovictor/bifrost/pkg/services"
)

// ---- management endpoint tests ----

func TestCreateServiceAccount_OK(t *testing.T) {
	env := newTestEnv(t)
	body := `{"name":"ci-pipeline"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/serviceaccounts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var sa serviceaccounts.ServiceAccount
	if err := json.Unmarshal(rr.Body.Bytes(), &sa); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if sa.ID == "" {
		t.Error("expected generated ID")
	}
	if sa.APIKey == "" {
		t.Error("expected generated api_key")
	}
	if sa.Name != "ci-pipeline" {
		t.Errorf("unexpected name: %s", sa.Name)
	}
}

func TestCreateServiceAccount_MissingName(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/serviceaccounts", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateServiceAccount_Duplicate(t *testing.T) {
	env := newTestEnv(t)
	env.Server.ServiceAccountStore.Create(serviceaccounts.ServiceAccount{
		ID: "sa-dup", Name: "dup", APIKey: "key-dup",
	})
	body := `{"id":"sa-dup","name":"dup2"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/serviceaccounts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestListServiceAccounts(t *testing.T) {
	env := newTestEnv(t)
	env.Server.ServiceAccountStore.Create(serviceaccounts.ServiceAccount{ID: "sa-1", Name: "one", APIKey: "k1"})
	env.Server.ServiceAccountStore.Create(serviceaccounts.ServiceAccount{ID: "sa-2", Name: "two", APIKey: "k2"})

	req := httptest.NewRequest(http.MethodGet, "/v1/serviceaccounts", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var list []serviceaccounts.ServiceAccount
	if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(list))
	}
}

func TestDeleteServiceAccount_OK(t *testing.T) {
	env := newTestEnv(t)
	env.Server.ServiceAccountStore.Create(serviceaccounts.ServiceAccount{ID: "sa-del", Name: "del", APIKey: "k-del"})

	req := httptest.NewRequest(http.MethodDelete, "/v1/serviceaccounts/sa-del", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestDeleteServiceAccount_NotFound(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodDelete, "/v1/serviceaccounts/no-such", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ---- POST /v1/service-token tests ----

func seedServiceForToken(t *testing.T, env *TestEnv) string {
	t.Helper()
	rk := rootkeys.RootKey{ID: "rk-st", APIKey: "real"}
	env.Server.RootKeyStore.Create(rk)
	svc := services.Service{ID: "svc-st", Endpoint: "http://localhost", RootKeyID: rk.ID}
	env.Server.ServiceStore.Create(svc)
	return svc.ID
}

func TestServiceToken_OK(t *testing.T) {
	env := newTestEnv(t)
	svcID := seedServiceForToken(t, env)
	env.Server.ServiceAccountStore.Create(serviceaccounts.ServiceAccount{
		ID:     "sa-tok",
		Name:   "ci",
		APIKey: "sa-key-1",
	})

	body, _ := json.Marshal(map[string]interface{}{"service": svcID, "ttl_seconds": 60})
	req := httptest.NewRequest(http.MethodPost, "/v1/service-token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Key", "sa-key-1")
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Key       string    `json:"key"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Key == "" {
		t.Error("expected a virtual key ID")
	}
	if resp.ExpiresAt.IsZero() {
		t.Error("expected expires_at")
	}

	// Verify the virtual key was actually stored with the right source.
	k, err := env.Server.KeyStore.Get(resp.Key)
	if err != nil {
		t.Fatalf("get key: %v", err)
	}
	if k.Source != keys.SourceServiceAccount {
		t.Errorf("expected source=%q, got %q", keys.SourceServiceAccount, k.Source)
	}
}

func TestServiceToken_MissingServiceKey(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/service-token", bytes.NewBufferString(`{"service":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestServiceToken_InvalidServiceKey(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/service-token", bytes.NewBufferString(`{"service":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Key", "bad-key")
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestServiceToken_ServiceNotAllowed(t *testing.T) {
	env := newTestEnv(t)
	seedServiceForToken(t, env)
	env.Server.ServiceAccountStore.Create(serviceaccounts.ServiceAccount{
		ID:              "sa-restricted",
		Name:            "restricted",
		APIKey:          "sa-key-r",
		AllowedServices: serviceaccounts.StringList{"other-svc"},
	})

	body, _ := json.Marshal(map[string]interface{}{"service": "svc-st"})
	req := httptest.NewRequest(http.MethodPost, "/v1/service-token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Key", "sa-key-r")
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestServiceToken_ServiceNotFound(t *testing.T) {
	env := newTestEnv(t)
	env.Server.ServiceAccountStore.Create(serviceaccounts.ServiceAccount{
		ID: "sa-any", Name: "any", APIKey: "sa-key-any",
	})

	body, _ := json.Marshal(map[string]interface{}{"service": "no-such-svc"})
	req := httptest.NewRequest(http.MethodPost, "/v1/service-token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Key", "sa-key-any")
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestServiceToken_EmptyAllowedServicesAllowsAny(t *testing.T) {
	env := newTestEnv(t)
	svcID := seedServiceForToken(t, env)
	env.Server.ServiceAccountStore.Create(serviceaccounts.ServiceAccount{
		ID:              "sa-open",
		Name:            "open",
		APIKey:          "sa-key-open",
		AllowedServices: serviceaccounts.StringList{}, // empty = allow all
	})

	body, _ := json.Marshal(map[string]interface{}{"service": svcID})
	req := httptest.NewRequest(http.MethodPost, "/v1/service-token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Key", "sa-key-open")
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 (empty allowed_services = allow all), got %d: %s", rr.Code, rr.Body.String())
	}
}
