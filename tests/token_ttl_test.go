package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/farovictor/bifrost/pkg/auth"
)

// createUserWithTTL posts to /v1/users with an optional ttl field and returns
// the decoded token string.
func createUserWithTTL(t *testing.T, env *TestEnv, name, email, ttl string) string {
	t.Helper()
	payload := map[string]string{"name": name, "email": email}
	if ttl != "" {
		payload["ttl"] = ttl
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// CreateUser requires a bearer token (OrgCtxMiddleware)
	req.Header.Set("Authorization", "Bearer "+env.Token)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create user: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
	}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	return resp.Token
}

func TestCreateUser_DefaultTTL(t *testing.T) {
	env := newTestEnv(t)
	token := createUserWithTTL(t, env, "Bob", "bob@example.com", "")

	tok, err := auth.Verify(token)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	ttl := time.Until(tok.ExpiresAt)
	// Default is 24h — allow ±5 minutes for test execution time.
	if ttl < 23*time.Hour+55*time.Minute || ttl > 24*time.Hour+5*time.Minute {
		t.Errorf("expected ~24h TTL, got %v", ttl)
	}
}

func TestCreateUser_CustomTTL(t *testing.T) {
	env := newTestEnv(t)
	token := createUserWithTTL(t, env, "Carol", "carol@example.com", "30m")

	tok, err := auth.Verify(token)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	ttl := time.Until(tok.ExpiresAt)
	if ttl < 29*time.Minute || ttl > 31*time.Minute {
		t.Errorf("expected ~30m TTL, got %v", ttl)
	}
}

func TestCreateUser_InvalidTTL(t *testing.T) {
	env := newTestEnv(t)
	body := `{"name":"Dave","email":"dave@example.com","ttl":"notaduration"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+env.Token)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid ttl, got %d", rr.Code)
	}
}

func TestRefreshToken_DefaultTTL(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/token/refresh", nil)
	req.Header.Set("Authorization", "Bearer "+env.Token)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct{ Token string `json:"token"` }
	json.Unmarshal(rr.Body.Bytes(), &resp)
	tok, _ := auth.Verify(resp.Token)
	ttl := time.Until(tok.ExpiresAt)
	if ttl < 23*time.Hour+55*time.Minute || ttl > 24*time.Hour+5*time.Minute {
		t.Errorf("expected ~24h TTL on refresh, got %v", ttl)
	}
}

func TestRefreshToken_CustomTTL(t *testing.T) {
	env := newTestEnv(t)
	body := `{"ttl":"15m"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/token/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+env.Token)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct{ Token string `json:"token"` }
	json.Unmarshal(rr.Body.Bytes(), &resp)
	tok, _ := auth.Verify(resp.Token)
	ttl := time.Until(tok.ExpiresAt)
	if ttl < 14*time.Minute || ttl > 16*time.Minute {
		t.Errorf("expected ~15m TTL on refresh, got %v", ttl)
	}
}

func TestRefreshToken_InvalidTTL(t *testing.T) {
	env := newTestEnv(t)
	body := `{"ttl":"bad"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/token/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+env.Token)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid ttl on refresh, got %d", rr.Code)
	}
}
