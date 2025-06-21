package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FokusInternal/bifrost/pkg/rootkeys"
	routes "github.com/FokusInternal/bifrost/routes"
)

func TestCreateRootKey(t *testing.T) {
	routes.RootKeyStore = rootkeys.NewStore()
	router := setupRouter()

	rk := rootkeys.RootKey{ID: "rk", APIKey: "secret"}
	body, _ := json.Marshal(rk)
	req := httptest.NewRequest(http.MethodPost, "/v1/rootkeys", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

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
	routes.RootKeyStore = rootkeys.NewStore()
	rk := rootkeys.RootKey{ID: "dead", APIKey: "k"}
	if err := routes.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("failed to seed store: %v", err)
	}
	router := setupRouter()
	req := httptest.NewRequest(http.MethodDelete, "/v1/rootkeys/"+rk.ID, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if _, err := routes.RootKeyStore.Get(rk.ID); err != rootkeys.ErrKeyNotFound {
		t.Fatalf("root key was not deleted")
	}
}
