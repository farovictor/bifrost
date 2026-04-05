package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	routes "github.com/farovictor/bifrost/routes"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/usage"
)

func setupBudgetEnv(t *testing.T, budget int) (s *routes.Server, router http.Handler, keyID string, backend *httptest.Server) {
	t.Helper()
	s = newTestServer(t)

	backend = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"resp"}`))
	}))
	t.Cleanup(backend.Close)

	rk := rootkeys.RootKey{ID: "rk-budget", APIKey: "real"}
	s.RootKeyStore.Create(rk)
	svc := services.Service{ID: "svc-budget", Endpoint: backend.URL, RootKeyID: rk.ID}
	s.ServiceStore.Create(svc)

	keyID = "vk-budget"
	k := keys.VirtualKey{
		ID:          keyID,
		Target:      svc.ID,
		Scope:       keys.ScopeWrite,
		RateLimit:   100,
		ExpiresAt:   time.Now().Add(time.Hour),
		TokenBudget: budget,
	}
	s.KeyStore.Create(k)

	router = setupRouter(s)
	return
}

func TestTokenBudget_AllowedWhenUnderBudget(t *testing.T) {
	s, router, keyID, _ := setupBudgetEnv(t, 1000)

	// Pre-seed some usage below the budget
	s.UsageStore.Record(usage.Event{
		KeyID:       keyID,
		Timestamp:   time.Now(),
		StatusCode:  200,
		Service:     "svc-budget",
		LatencyMS:   10,
		TotalTokens: 500,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/proxy/chat", nil)
	req.Header.Set("X-Virtual-Key", keyID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 under budget, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestTokenBudget_BlockedWhenExceeded(t *testing.T) {
	s, router, keyID, _ := setupBudgetEnv(t, 100)

	// Pre-seed usage at or above the budget
	s.UsageStore.Record(usage.Event{
		KeyID:       keyID,
		Timestamp:   time.Now(),
		StatusCode:  200,
		Service:     "svc-budget",
		LatencyMS:   10,
		TotalTokens: 100,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/proxy/chat", nil)
	req.Header.Set("X-Virtual-Key", keyID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 when budget exceeded, got %d: %s", rr.Code, rr.Body.String())
	}
	if msg := errorBody(t, rr); msg != "token budget exceeded" {
		t.Errorf("unexpected error message: %s", msg)
	}
}

func TestTokenBudget_ZeroMeansUnlimited(t *testing.T) {
	s, router, keyID, _ := setupBudgetEnv(t, 0)

	// Seed a huge amount — should still be allowed (budget=0 → unlimited)
	s.UsageStore.Record(usage.Event{
		KeyID:       keyID,
		Timestamp:   time.Now(),
		StatusCode:  200,
		Service:     "svc-budget",
		LatencyMS:   10,
		TotalTokens: 999_999_999,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/proxy/chat", nil)
	req.Header.Set("X-Virtual-Key", keyID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 when budget=0 (unlimited), got %d: %s", rr.Code, rr.Body.String())
	}
}
