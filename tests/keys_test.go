package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FokusInternal/bifrost/pkg/keys"
	routes "github.com/FokusInternal/bifrost/routes"
)

func TestCreateKey(t *testing.T) {
	routes.KeyStore = keys.NewStore()
	router := setupRouter()

	k := keys.VirtualKey{ID: "abc", Scope: "test", Target: "svc", ExpiresAt: time.Now()}
	body, _ := json.Marshal(k)
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader(body))
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
	routes.KeyStore = keys.NewStore()
	k := keys.VirtualKey{ID: "dead", Scope: "x", Target: "svc", ExpiresAt: time.Now()}
	if err := routes.KeyStore.Create(k); err != nil {
		t.Fatalf("failed to seed store: %v", err)
	}

	router := setupRouter()
	req := httptest.NewRequest(http.MethodDelete, "/v1/keys/"+k.ID, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}

	if _, err := routes.KeyStore.Get(k.ID); err != keys.ErrKeyNotFound {
		t.Fatalf("key was not deleted")
	}
}
