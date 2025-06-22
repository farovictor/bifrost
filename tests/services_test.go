package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/users"
	routes "github.com/farovictor/bifrost/routes"
)

func TestCreateService(t *testing.T) {
	routes.ServiceStore = services.NewMemoryStore()
	routes.RootKeyStore = rootkeys.NewMemoryStore()
	routes.UserStore = users.NewMemoryStore()
	u := users.User{ID: "u", APIKey: "secret"}
	routes.UserStore.Create(u)
	rk := rootkeys.RootKey{ID: "rk", APIKey: "k"}
	if err := routes.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}
	router := setupRouter()

	svc := services.Service{ID: "svc", Endpoint: "http://example.com", RootKeyID: rk.ID}
	body, _ := json.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/v1/services", bytes.NewReader(body))
	req.Header.Set("X-API-Key", u.APIKey)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var resp services.Service
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != svc.ID {
		t.Fatalf("expected ID %s, got %s", svc.ID, resp.ID)
	}
}

func TestDeleteService(t *testing.T) {
	routes.ServiceStore = services.NewMemoryStore()
	routes.RootKeyStore = rootkeys.NewMemoryStore()
	routes.UserStore = users.NewMemoryStore()
	u := users.User{ID: "u", APIKey: "secret"}
	routes.UserStore.Create(u)
	rk := rootkeys.RootKey{ID: "rkdead", APIKey: "k"}
	if err := routes.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}
	svc := services.Service{ID: "dead", Endpoint: "http://example.com", RootKeyID: rk.ID}
	if err := routes.ServiceStore.Create(svc); err != nil {
		t.Fatalf("failed to seed store: %v", err)
	}
	router := setupRouter()
	req := httptest.NewRequest(http.MethodDelete, "/v1/services/"+svc.ID, nil)
	req.Header.Set("X-API-Key", u.APIKey)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if _, err := routes.ServiceStore.Get(svc.ID); err != services.ErrServiceNotFound {
		t.Fatalf("service was not deleted")
	}
}
