package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
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
			s := newTestServer(t)

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
			if err := s.RootKeyStore.Create(rk); err != nil {
				t.Fatalf("seed rootkey: %v", err)
			}
			svc := services.Service{ID: "svc", Endpoint: backend.URL, RootKeyID: rk.ID}
			if err := s.ServiceStore.Create(svc); err != nil {
				t.Fatalf("seed service: %v", err)
			}
			k := keys.VirtualKey{ID: "vkey-" + tc.name, Target: svc.ID, Scope: keys.ScopeRead, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 100}
			if err := s.KeyStore.Create(k); err != nil {
				t.Fatalf("seed key: %v", err)
			}

			router := setupRouter(s)
			req := httptest.NewRequest(http.MethodGet, "/v1/proxy/backend?foo=bar", nil)
			if tc.useQuery {
				req = httptest.NewRequest(http.MethodGet, "/v1/proxy/backend?key="+k.ID+"&foo=bar", nil)
			} else {
				req.Header.Set("X-Virtual-Key", k.ID)
			}
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

func TestProxyMultiSegmentPath(t *testing.T) {
	s := newTestServer(t)

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok")
	}))
	defer backend.Close()

	rk := rootkeys.RootKey{ID: "rk-ms", APIKey: "real"}
	s.RootKeyStore.Create(rk)
	svc := services.Service{ID: "svc-ms", Endpoint: backend.URL, RootKeyID: rk.ID}
	s.ServiceStore.Create(svc)
	k := keys.VirtualKey{ID: "vk-ms", Target: svc.ID, Scope: keys.ScopeWrite, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 100}
	s.KeyStore.Create(k)

	router := setupRouter(s)
	req := httptest.NewRequest(http.MethodPost, "/v1/proxy/v1/chat/completions", nil)
	req.Header.Set("X-Virtual-Key", k.ID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — multi-segment proxy path not routed correctly", rr.Code)
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
			s := newTestServer(t)

			called := false
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				io.WriteString(w, "ok")
			}))
			defer backend.Close()

			rk := rootkeys.RootKey{ID: "rk-" + tc.name, APIKey: "real"}
			if err := s.RootKeyStore.Create(rk); err != nil {
				t.Fatalf("seed rootkey: %v", err)
			}
			svc := services.Service{ID: "svc-" + tc.name, Endpoint: backend.URL, RootKeyID: rk.ID}
			if err := s.ServiceStore.Create(svc); err != nil {
				t.Fatalf("seed service: %v", err)
			}
			k := keys.VirtualKey{ID: "vk-" + tc.name, Target: svc.ID, Scope: tc.scope, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 100}
			if err := s.KeyStore.Create(k); err != nil {
				t.Fatalf("seed key: %v", err)
			}

			router := setupRouter(s)
			req := httptest.NewRequest(tc.method, "/v1/proxy/backend", nil)
			req.Header.Set("X-Virtual-Key", k.ID)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tc.wantCode {
				t.Fatalf("expected %d, got %d", tc.wantCode, rr.Code)
			}
			if tc.wantCode == http.StatusForbidden {
				if msg := errorBody(t, rr); msg != "insufficient scope" {
					t.Fatalf("unexpected error: %s", msg)
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

func seedProxyBackend(t *testing.T, s *routes.Server) (svcID string, backend *httptest.Server) {
	t.Helper()
	backend = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok")
	}))
	t.Cleanup(backend.Close)
	rk := rootkeys.RootKey{ID: "rk-os-" + svcID, APIKey: "real"}
	s.RootKeyStore.Create(rk)
	svc := services.Service{ID: "svc-os", Endpoint: backend.URL, RootKeyID: rk.ID}
	s.ServiceStore.Create(svc)
	return "svc-os", backend
}

func TestOneShotKeyUsedOnce(t *testing.T) {
	s := newTestServer(t)
	svcID, _ := seedProxyBackend(t, s)

	k := keys.VirtualKey{ID: "vk-oneshot", Target: svcID, Scope: keys.ScopeRead, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 100, OneShot: true}
	s.KeyStore.Create(k)

	router := setupRouter(s)

	// First request — should succeed.
	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/path", nil)
	req.Header.Set("X-Virtual-Key", k.ID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", rr.Code)
	}

	// Second request — key is now used, must be rejected.
	req2 := httptest.NewRequest(http.MethodGet, "/v1/proxy/path", nil)
	req2.Header.Set("X-Virtual-Key", k.ID)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusUnauthorized {
		t.Fatalf("second request: expected 401, got %d", rr2.Code)
	}
	if msg := errorBody(t, rr2); msg != "key already used" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestOneShotKeyAppearsInList(t *testing.T) {
	env := newTestEnv(t)
	svcID, _ := seedProxyBackend(t, env.Server)

	k := keys.VirtualKey{ID: "vk-os-list", Target: svcID, Scope: keys.ScopeRead, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 10, OneShot: true}
	env.Server.KeyStore.Create(k)

	req := httptest.NewRequest(http.MethodGet, "/v1/keys", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	var listed []keys.VirtualKey
	if err := json.Unmarshal(rr.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, lk := range listed {
		if lk.ID == k.ID {
			if !lk.OneShot {
				t.Error("expected one_shot=true in list response")
			}
			return
		}
	}
	t.Fatal("one-shot key not found in list")
}

func TestRegularKeyNotAffectedByOneShotLogic(t *testing.T) {
	s := newTestServer(t)
	svcID, _ := seedProxyBackend(t, s)

	k := keys.VirtualKey{ID: "vk-regular-os", Target: svcID, Scope: keys.ScopeRead, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 100}
	s.KeyStore.Create(k)

	router := setupRouter(s)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/proxy/path", nil)
		req.Header.Set("X-Virtual-Key", k.ID)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}
}
