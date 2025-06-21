package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/users"
	routes "github.com/farovictor/bifrost/routes"
)

func TestProxy(t *testing.T) {
	cases := []struct {
		name     string
		useQuery bool
	}{
		{name: "header", useQuery: false},
		{name: "query", useQuery: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			routes.ServiceStore = services.NewStore()
			routes.KeyStore = keys.NewStore()
			routes.RootKeyStore = rootkeys.NewStore()
			routes.UserStore = users.NewStore()
			u := users.User{ID: "u", APIKey: "secret"}
			routes.UserStore.Create(u)

			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-API-Key") != "real" {
					t.Fatalf("missing injected api key")
				}
				if r.Header.Get("X-Virtual-Key") != "" {
					t.Fatalf("virtual key header leaked")
				}
				if r.URL.Path != "/backend" {
					t.Fatalf("unexpected path %s", r.URL.Path)
				}
				if r.URL.Query().Get("key") != "" {
					t.Fatalf("virtual key leaked in query")
				}
				if r.URL.RawQuery != "foo=bar" {
					t.Fatalf("unexpected query %s", r.URL.RawQuery)
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
			k := keys.VirtualKey{ID: "vkey", Target: svc.ID, Scope: keys.ScopeRead, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
			if err := routes.KeyStore.Create(k); err != nil {
				t.Fatalf("seed key: %v", err)
			}

			router := setupRouter()
			url := "/v1/proxy/backend?foo=bar"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			if tc.useQuery {
				req = httptest.NewRequest(http.MethodGet, "/v1/proxy/backend?key="+k.ID+"&foo=bar", nil)
			} else {
				req.Header.Set("X-Virtual-Key", k.ID)
			}
			req.Header.Set("X-API-Key", u.APIKey)
			req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rr.Code)
			}
			if body := rr.Body.String(); body != "proxied" {
				t.Fatalf("unexpected body: %s", body)
			}
		})
	}
}

func TestProxyScopeEnforcement(t *testing.T) {
	cases := []struct {
		name     string
		scope    string
		method   string
		wantCode int
	}{
		{name: "read-get", scope: keys.ScopeRead, method: http.MethodGet, wantCode: http.StatusOK},
		{name: "read-post", scope: keys.ScopeRead, method: http.MethodPost, wantCode: http.StatusForbidden},
		{name: "write-post", scope: keys.ScopeWrite, method: http.MethodPost, wantCode: http.StatusOK},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			routes.ServiceStore = services.NewStore()
			routes.KeyStore = keys.NewStore()
			routes.RootKeyStore = rootkeys.NewStore()
			routes.UserStore = users.NewStore()
			u := users.User{ID: "u", APIKey: "secret"}
			routes.UserStore.Create(u)

			called := false
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				io.WriteString(w, "ok")
			}))
			defer backend.Close()

			rk := rootkeys.RootKey{ID: "rk-" + tc.name, APIKey: "real"}
			if err := routes.RootKeyStore.Create(rk); err != nil {
				t.Fatalf("seed rootkey: %v", err)
			}
			svc := services.Service{ID: "svc-" + tc.name, Endpoint: backend.URL, RootKeyID: rk.ID}
			if err := routes.ServiceStore.Create(svc); err != nil {
				t.Fatalf("seed service: %v", err)
			}
			k := keys.VirtualKey{ID: "vk-" + tc.name, Target: svc.ID, Scope: tc.scope, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
			if err := routes.KeyStore.Create(k); err != nil {
				t.Fatalf("seed key: %v", err)
			}

			router := setupRouter()
			req := httptest.NewRequest(tc.method, "/v1/proxy/backend", nil)
			req.Header.Set("X-Virtual-Key", k.ID)
			req.Header.Set("X-API-Key", u.APIKey)
			req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tc.wantCode {
				t.Fatalf("expected %d, got %d", tc.wantCode, rr.Code)
			}
			if tc.wantCode == http.StatusForbidden {
				if body := strings.TrimSpace(rr.Body.String()); body != "insufficient scope" {
					t.Fatalf("unexpected body: %s", body)
				}
				if called {
					t.Fatalf("backend should not be called")
				}
			} else {
				if !called {
					t.Fatalf("backend not called")
				}
			}
		})
	}
}
