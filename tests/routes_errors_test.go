package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FokusInternal/bifrost/pkg/keys"
	routes "github.com/FokusInternal/bifrost/routes"
)

func TestCreateKeyInvalidJSON(t *testing.T) {
	routes.KeyStore = keys.NewStore()
	router := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/v1/keys", strings.NewReader("{bad"))
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
	routes.KeyStore = keys.NewStore()
	k := keys.VirtualKey{ID: "dup", Scope: "test", Target: "svc", ExpiresAt: time.Now()}
	if err := routes.KeyStore.Create(k); err != nil {
		t.Fatalf("failed to seed store: %v", err)
	}

	body, _ := json.Marshal(k)
	router := setupRouter()
	req := httptest.NewRequest(http.MethodPost, "/v1/keys", bytes.NewReader(body))
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
	routes.KeyStore = keys.NewStore()
	router := setupRouter()

	req := httptest.NewRequest(http.MethodDelete, "/v1/keys/unknown", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}

	if body := strings.TrimSpace(rr.Body.String()); body != "not found" {
		t.Fatalf("unexpected body: %s", body)
	}
}
