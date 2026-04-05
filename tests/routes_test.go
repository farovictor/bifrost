package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	rl "github.com/farovictor/bifrost/middlewares"
	"github.com/go-chi/chi/v5"

	routes "github.com/farovictor/bifrost/routes"
	v1 "github.com/farovictor/bifrost/routes/v1"
)

func setupRouter(s *routes.Server) http.Handler {
	v1h := &v1.Handler{
		KeyStore:     s.KeyStore,
		ServiceStore: s.ServiceStore,
		RootKeyStore: s.RootKeyStore,
		UsageStore:   s.UsageStore,
	}
	r := chi.NewRouter()
	r.Get("/healthz", routes.Healthz)
	r.Get("/version", routes.Version)
	r.With(rl.AuthMiddleware(s.UserStore)).Post("/mcp", s.MCP)
	r.Route("/v1", func(r chi.Router) {
		r.With(rl.OrgCtxMiddleware(s.MembershipStore)).Post("/users", s.CreateUser)
		r.With(rl.OrgCtxMiddleware(s.MembershipStore)).Get("/user", s.GetUserInfo)
		r.With(rl.OrgCtxMiddleware(s.MembershipStore)).Post("/user/rootkeys", s.CreateRootKey)

		r.With(rl.RateLimitMiddleware(s.KeyStore)).Handle("/proxy/*", http.HandlerFunc(v1h.Proxy))

		r.Group(func(r chi.Router) {
			r.Use(rl.AuthMiddleware(s.UserStore))
			r.Use(rl.OrgCtxMiddleware(s.MembershipStore))
			r.Get("/hello", v1.SayHello)
			r.Get("/keys", s.ListKeys)
			r.Post("/keys", s.CreateKey)
			r.Delete("/keys/{id}", s.DeleteKey)
			r.Get("/keys/{id}/usage", s.ListKeyUsage)
			r.Get("/rootkeys", s.ListRootKeys)
			r.Post("/rootkeys", s.CreateRootKey)
			r.Put("/rootkeys/{id}", s.UpdateRootKey)
			r.Delete("/rootkeys/{id}", s.DeleteRootKey)
			r.Get("/services", s.ListServices)
			r.Post("/services", s.CreateService)
			r.Delete("/services/{id}", s.DeleteService)

			r.Get("/orgs", s.ListOrgs)
			r.Post("/orgs", s.CreateOrg)
			r.Get("/orgs/{id}", s.GetOrg)
			r.Delete("/orgs/{id}", s.DeleteOrg)
			r.Get("/orgs/{id}/members", s.ListOrgMembers)
			r.Post("/orgs/{id}/members", s.AddOrgMember)
			r.Delete("/orgs/{id}/members/{userID}", s.RemoveOrgMember)
		})
	})
	return r
}

func TestHealthz(t *testing.T) {
	router := setupRouter(newTestServer(t))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); body != "ok" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestVersion(t *testing.T) {
	router := setupRouter(newTestServer(t))
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["version"] == "" {
		t.Fatalf("version field is empty")
	}
}

func TestV1Hello(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/hello", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); body != "hello world" {
		t.Fatalf("unexpected body: %s", body)
	}
}
