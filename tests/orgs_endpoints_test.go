package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/farovictor/bifrost/pkg/orgs"
)

func TestCreateOrg(t *testing.T) {
	env := newTestEnv(t)
	o := orgs.Organization{Name: "Acme", Domain: "acme.com", Email: "admin@acme.com"}
	body, _ := json.Marshal(o)
	req := httptest.NewRequest(http.MethodPost, "/v1/orgs", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp orgs.Organization
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID == "" || resp.Name != o.Name {
		t.Fatalf("unexpected org: %#v", resp)
	}
}

func TestCreateOrgDuplicate(t *testing.T) {
	env := newTestEnv(t)
	o := orgs.Organization{ID: "dup", Name: "Dup", Domain: "d.com", Email: "d@d.com"}
	env.Server.OrgStore.Create(o)

	body, _ := json.Marshal(o)
	req := httptest.NewRequest(http.MethodPost, "/v1/orgs", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestListOrgs(t *testing.T) {
	env := newTestEnv(t)
	env.Server.OrgStore.Create(orgs.Organization{ID: "o1", Name: "One", Domain: "a.com", Email: "a@a.com"})
	env.Server.OrgStore.Create(orgs.Organization{ID: "o2", Name: "Two", Domain: "b.com", Email: "b@b.com"})

	req := httptest.NewRequest(http.MethodGet, "/v1/orgs", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp []orgs.Organization
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 orgs, got %d", len(resp))
	}
}

func TestGetOrg(t *testing.T) {
	env := newTestEnv(t)
	o := orgs.Organization{ID: "o1", Name: "One", Domain: "a.com", Email: "a@a.com"}
	env.Server.OrgStore.Create(o)

	req := httptest.NewRequest(http.MethodGet, "/v1/orgs/o1", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp orgs.Organization
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != o.ID || resp.Name != o.Name {
		t.Fatalf("unexpected org: %#v", resp)
	}
}

func TestGetOrgNotFound(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/orgs/missing", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDeleteOrgEndpoint(t *testing.T) {
	env := newTestEnv(t)
	o := orgs.Organization{ID: "todel", Name: "Del", Domain: "d.com", Email: "d@d.com"}
	env.Server.OrgStore.Create(o)

	req := httptest.NewRequest(http.MethodDelete, "/v1/orgs/todel", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if _, err := env.Server.OrgStore.Get(o.ID); err != orgs.ErrOrgNotFound {
		t.Fatalf("org was not deleted")
	}
}

func TestAddAndListOrgMembers(t *testing.T) {
	env := newTestEnv(t)
	o := orgs.Organization{ID: "org1", Name: "Org1", Domain: "x.com", Email: "x@x.com"}
	env.Server.OrgStore.Create(o)

	m := orgs.Membership{UserID: env.User.ID, Role: orgs.RoleMember}
	body, _ := json.Marshal(m)
	req := httptest.NewRequest(http.MethodPost, "/v1/orgs/org1/members", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("add member: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/v1/orgs/org1/members", nil)
	env.Authorize(req2)
	rr2 := httptest.NewRecorder()
	env.Router.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("list members: expected 200, got %d", rr2.Code)
	}
	var members []orgs.Membership
	if err := json.Unmarshal(rr2.Body.Bytes(), &members); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(members) != 1 || members[0].UserID != env.User.ID {
		t.Fatalf("unexpected members: %#v", members)
	}
}

func TestRemoveOrgMember(t *testing.T) {
	env := newTestEnv(t)
	o := orgs.Organization{ID: "org2", Name: "Org2", Domain: "y.com", Email: "y@y.com"}
	env.Server.OrgStore.Create(o)
	env.Server.MembershipStore.Create(orgs.Membership{UserID: env.User.ID, OrgID: o.ID, Role: orgs.RoleMember})

	req := httptest.NewRequest(http.MethodDelete, "/v1/orgs/org2/members/"+env.User.ID, nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if _, err := env.Server.MembershipStore.Get(env.User.ID, o.ID); err != orgs.ErrMembershipNotFound {
		t.Fatalf("membership was not removed")
	}
}
