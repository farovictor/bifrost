package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	rl "github.com/farovictor/bifrost/middlewares"
	"github.com/farovictor/bifrost/pkg/users"
)

// ── LoggingMiddleware ─────────────────────────────────────────────────────────

func TestLoggingMiddleware(t *testing.T) {
	called := false
	handler := rl.LoggingMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler was not called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

// ── AuthMiddleware ────────────────────────────────────────────────────────────

func TestAuthMiddlewareNoKey(t *testing.T) {
	store := users.NewMemoryStore()
	handler := rl.AuthMiddleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddlewareInvalidKey(t *testing.T) {
	store := users.NewMemoryStore()
	handler := rl.AuthMiddleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "bad-key")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddlewareValidKey(t *testing.T) {
	// In test mode AuthMiddleware validates against StaticAPIKey ("secret" by default).
	t.Setenv("BIFROST_STATIC_API_KEY", "secret")
	store := users.NewMemoryStore()
	u := users.User{ID: "u1", Name: "U", Email: "u@u.com", APIKey: "secret"}
	store.Create(u)

	called := false
	handler := rl.AuthMiddleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "secret")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler was not called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAuthMiddlewareKeyFromAuthHeader(t *testing.T) {
	// In test mode AuthMiddleware validates against StaticAPIKey ("secret" by default).
	t.Setenv("BIFROST_STATIC_API_KEY", "secret")
	store := users.NewMemoryStore()
	u := users.User{ID: "u1", Name: "U", Email: "u@u.com", APIKey: "secret"}
	store.Create(u)

	called := false
	handler := rl.AuthMiddleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	// key passed via Authorization: Bearer <key>
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler was not called")
	}
}
