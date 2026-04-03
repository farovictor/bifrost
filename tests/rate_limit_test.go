package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	rl "github.com/farovictor/bifrost/middlewares"
	"github.com/go-chi/chi/v5"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/users"
	routes "github.com/farovictor/bifrost/routes"
	v1 "github.com/farovictor/bifrost/routes/v1"
)

func setupRouterRL(s *routes.Server) http.Handler {
	v1h := &v1.Handler{
		KeyStore:     s.KeyStore,
		ServiceStore: s.ServiceStore,
		RootKeyStore: s.RootKeyStore,
	}
	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(rl.AuthMiddleware(s.UserStore))
		r.Use(rl.OrgCtxMiddleware(s.MembershipStore))
		r.With(rl.RateLimitMiddleware(s.KeyStore)).Handle("/proxy/{rest:.*}", http.HandlerFunc(v1h.Proxy))
	})
	return r
}

func TestRateLimitExceeded(t *testing.T) {
	s := newTestServer()
	u := users.User{ID: "u", Name: "U", Email: "u@example.com", APIKey: "secret"}
	s.UserStore.Create(u)

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
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
	k := keys.VirtualKey{ID: "lim", Target: svc.ID, Scope: keys.ScopeRead, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
	if err := s.KeyStore.Create(k); err != nil {
		t.Fatalf("seed key: %v", err)
	}

	router := setupRouterRL(s)

	req1 := httptest.NewRequest(http.MethodGet, "/v1/proxy/test", nil)
	req1.Header.Set("X-Virtual-Key", k.ID)
	req1.Header.Set("X-API-Key", u.APIKey)
	req1.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr1 := httptest.NewRecorder()
	router.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/v1/proxy/test", nil)
	req2.Header.Set("X-Virtual-Key", k.ID)
	req2.Header.Set("X-API-Key", u.APIKey)
	req2.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr2.Code)
	}
}
