package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FokusInternal/bifrost/pkg/keys"
	"github.com/FokusInternal/bifrost/pkg/rootkeys"
	"github.com/FokusInternal/bifrost/pkg/services"
	routes "github.com/FokusInternal/bifrost/routes"
)

func TestProxy(t *testing.T) {
	routes.ServiceStore = services.NewStore()
	routes.KeyStore = keys.NewStore()
	routes.RootKeyStore = rootkeys.NewStore()

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "real" {
			t.Fatalf("missing injected api key")
		}
		if r.URL.Path != "/backend" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		io.WriteString(w, "proxied")
	}))
	defer backend.Close()

	rk := rootkeys.RootKey{ID: "rk", APIKey: "real"}
	if err := routes.RootKeyStore.Create(rk); err != nil {
		t.Fatalf("seed rootkey: %v", err)
	}
	svc := services.Service{ID: "svc", Endpoint: backend.URL, RootKeyID: rk.ID}
	if err := routes.ServiceStore.Create(svc); err != nil {
		t.Fatalf("seed service: %v", err)
	}
	k := keys.VirtualKey{ID: "vkey", Target: svc.ID, Scope: "test", ExpiresAt: time.Now().Add(time.Hour)}
	if err := routes.KeyStore.Create(k); err != nil {
		t.Fatalf("seed key: %v", err)
	}

	router := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/backend", nil)
	req.Header.Set("X-Virtual-Key", k.ID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); body != "proxied" {
		t.Fatalf("unexpected body: %s", body)
	}
}
