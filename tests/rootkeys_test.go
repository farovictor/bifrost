package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/farovictor/bifrost/pkg/rootkeys"
)

func TestCreateRootKey(t *testing.T) {
	env := newTestEnv(t)
	rk := rootkeys.RootKey{ID: "rk", APIKey: "secret"}
	body, _ := json.Marshal(rk)
	req := httptest.NewRequest(http.MethodPost, "/v1/rootkeys", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	var resp rootkeys.RootKey
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != rk.ID {
		t.Fatalf("expected ID %s, got %s", rk.ID, resp.ID)
	}
}

func TestDeleteRootKey(t *testing.T) {
	env := newTestEnv(t)
	rk := rootkeys.RootKey{ID: "dead", APIKey: "k"}
	if err := env.Server.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("failed to seed store: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/rootkeys/"+rk.ID, nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if _, err := env.Server.RootKeyStore.Get(rk.ID); err != rootkeys.ErrKeyNotFound {
		t.Fatalf("root key was not deleted")
	}
}

func TestUpdateRootKey(t *testing.T) {
	env := newTestEnv(t)
	orig := rootkeys.RootKey{ID: "rk", APIKey: "old"}
	if err := env.Server.RootKeyStore.Create(orig); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}

	updated := rootkeys.RootKey{ID: "rk", APIKey: "new"}
	body, _ := json.Marshal(updated)
	req := httptest.NewRequest(http.MethodPut, "/v1/rootkeys/"+orig.ID, bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	rk, err := env.Server.RootKeyStore.Get(orig.ID)
	if err != nil {
		t.Fatalf("get rootkey: %v", err)
	}
	if rk.APIKey != updated.APIKey {
		t.Fatalf("update did not persist")
	}
}

func TestCreateRootKeyWithBearer(t *testing.T) {
	env := newTestEnv(t)
	rk := rootkeys.RootKey{ID: "rk2", APIKey: "secret"}
	body, _ := json.Marshal(rk)
	req := httptest.NewRequest(http.MethodPost, "/v1/user/rootkeys", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+env.Token)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	if _, err := env.Server.RootKeyStore.Get(rk.ID); err != nil {
		t.Fatalf("root key not stored: %v", err)
	}
}

func TestCreateRootKeyWithBearerUnauthorized(t *testing.T) {
	router := setupRouter(newTestServer(t))
	rk := rootkeys.RootKey{ID: "rk3", APIKey: "secret"}
	body, _ := json.Marshal(rk)
	req := httptest.NewRequest(http.MethodPost, "/v1/user/rootkeys", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}
