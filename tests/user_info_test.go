package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/users"
	routes "github.com/farovictor/bifrost/routes"
)

func TestGetUserInfo(t *testing.T) {
	routes.UserStore = users.NewMemoryStore()
	routes.OrgStore = orgs.NewMemoryStore()
	routes.MembershipStore = orgs.NewMembershipStore()

	u := users.User{ID: "u1", Name: "User", Email: "u@example.com", APIKey: "key"}
	routes.UserStore.Create(u)

	o := orgs.Organization{ID: "o1", Name: "Org", Domain: "example.com", Email: "org@example.com"}
	routes.OrgStore.Create(o)
	routes.MembershipStore.Create(orgs.Membership{UserID: u.ID, OrgID: o.ID, Role: orgs.RoleOwner})

	router := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/v1/user", nil)
	req.Header.Set("Authorization", "Bearer "+makeToken(u.ID))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp struct {
		ID    string
		Name  string
		Email string
		Orgs  []struct {
			OrgID string `json:"org_id"`
			Name  string `json:"name"`
			Role  string `json:"role"`
		}
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != u.ID || resp.Name != u.Name || resp.Email != u.Email {
		t.Fatalf("unexpected user info: %#v", resp)
	}
	if len(resp.Orgs) != 1 || resp.Orgs[0].OrgID != o.ID ||
		resp.Orgs[0].Name != o.Name || resp.Orgs[0].Role != orgs.RoleOwner {
		t.Fatalf("unexpected orgs: %#v", resp.Orgs)
	}
}
