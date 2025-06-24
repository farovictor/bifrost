package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/farovictor/bifrost/pkg/auth"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/users"
	routes "github.com/farovictor/bifrost/routes"
)

func TestCreateUserReturnsToken(t *testing.T) {
	routes.UserStore = users.NewMemoryStore()
	routes.OrgStore = orgs.NewMemoryStore()
	routes.MembershipStore = orgs.NewMemoryMembershipStore()

	admin := users.User{ID: "admin", Name: "Admin", Email: "admin@example.com", APIKey: "secret"}
	routes.UserStore.Create(admin)

	router := setupRouter()

	payload := `{"name":"New User","email":"new@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/users", strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+makeToken(admin.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var resp struct {
		users.User
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID == "" {
		t.Fatalf("missing id")
	}
	if resp.Name != "New User" || resp.Email != "new@example.com" {
		t.Fatalf("unexpected user data: %#v", resp.User)
	}
	if resp.Token == "" {
		t.Fatalf("missing token")
	}

	tok, err := auth.Verify(resp.Token)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if tok.UserID != resp.ID || tok.OrgID != "" {
		t.Fatalf("unexpected token payload: %#v", tok)
	}
	if time.Until(tok.ExpiresAt) > 24*time.Hour || time.Until(tok.ExpiresAt) <= 0 {
		t.Fatalf("unexpected expiry")
	}
}
func TestCreateUserDuplicateSameOrg(t *testing.T) {
	routes.UserStore = users.NewMemoryStore()
	routes.OrgStore = orgs.NewMemoryStore()
	routes.MembershipStore = orgs.NewMemoryMembershipStore()

	admin := users.User{ID: "admin", Name: "Admin", Email: "admin@example.com", APIKey: "secret"}
	routes.UserStore.Create(admin)

	org := orgs.Organization{ID: "org1", Name: "Org", Domain: "example.com", Email: "org@example.com"}
	routes.OrgStore.Create(org)

	router := setupRouter()

	payload := `{"name":"User","email":"dup@example.com","org_id":"org1"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/users", strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+makeToken(admin.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("first create status %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/v1/users", strings.NewReader(payload))
	req2.Header.Set("Authorization", "Bearer "+makeToken(admin.ID))
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rr2.Code)
	}
	if body := strings.TrimSpace(rr2.Body.String()); body != "user already exists" {
		t.Fatalf("unexpected body: %s", body)
	}
}
