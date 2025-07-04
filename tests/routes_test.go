package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	rl "github.com/farovictor/bifrost/middlewares"
	"github.com/go-chi/chi/v5"

	"github.com/farovictor/bifrost/pkg/users"
	routes "github.com/farovictor/bifrost/routes"
	v1 "github.com/farovictor/bifrost/routes/v1"
)

func setupRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", routes.Healthz)
	r.Get("/version", routes.Version)
	r.Route("/v1", func(r chi.Router) {
		r.With(rl.OrgCtxMiddleware()).Post("/users", routes.CreateUser)
		r.With(rl.OrgCtxMiddleware()).Get("/user", routes.GetUserInfo)
		r.With(rl.OrgCtxMiddleware()).Post("/user/rootkeys", routes.CreateRootKey)

		r.Group(func(r chi.Router) {
			r.Use(rl.AuthMiddleware())
			r.Use(rl.OrgCtxMiddleware())
			r.Get("/hello", v1.SayHello)
			r.Post("/keys", routes.CreateKey)
			r.Delete("/keys/{id}", routes.DeleteKey)
			r.Post("/rootkeys", routes.CreateRootKey)
			r.Put("/rootkeys/{id}", routes.UpdateRootKey)
			r.Delete("/rootkeys/{id}", routes.DeleteRootKey)
			r.Post("/services", routes.CreateService)
			r.Delete("/services/{id}", routes.DeleteService)
			r.Handle("/proxy/{rest:.*}", http.HandlerFunc(v1.Proxy))
		})
	})
	return r
}

func TestHealthz(t *testing.T) {
	router := setupRouter()
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
	router := setupRouter()
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
	routes.UserStore = users.NewMemoryStore()
	u := users.User{ID: "u", Name: "U", Email: "u@example.com", APIKey: "secret"}
	routes.UserStore.Create(u)
	router := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/v1/hello", nil)
	req.Header.Set("X-API-Key", u.APIKey)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	if body := rr.Body.String(); body != "hello world" {
		t.Fatalf("unexpected body: %s", body)
	}
}
