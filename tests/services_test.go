package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FokusInternal/bifrost/pkg/services"
	routes "github.com/FokusInternal/bifrost/routes"
)

func TestCreateService(t *testing.T) {
	routes.ServiceStore = services.NewStore()
	router := setupRouter()

	svc := services.Service{ID: "svc", Endpoint: "http://example.com", APIKey: "k"}
	body, _ := json.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/v1/services", bytes.NewReader(body))
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
	routes.ServiceStore = services.NewStore()
	svc := services.Service{ID: "dead", Endpoint: "http://example.com", APIKey: "k"}
	if err := routes.ServiceStore.Create(svc); err != nil {
		t.Fatalf("failed to seed store: %v", err)
	}
	router := setupRouter()
	req := httptest.NewRequest(http.MethodDelete, "/v1/services/"+svc.ID, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if _, err := routes.ServiceStore.Get(svc.ID); err != services.ErrServiceNotFound {
		t.Fatalf("service was not deleted")
	}
}
