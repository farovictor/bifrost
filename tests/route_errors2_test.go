package tests

// Additional error-path coverage for routes that are below 70%.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
)

// ── rootkeys ──────────────────────────────────────────────────────────────────

func TestCreateRootKeyInvalidJSON(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/rootkeys", strings.NewReader("{bad"))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "invalid request" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestCreateRootKeyDuplicate(t *testing.T) {
	env := newTestEnv(t)
	rk := rootkeys.RootKey{ID: "dup", APIKey: "k"}
	env.Server.RootKeyStore.Create(rk)
	body, _ := json.Marshal(rk)
	req := httptest.NewRequest(http.MethodPost, "/v1/rootkeys", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "root key already exists" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestDeleteRootKeyNotFound(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodDelete, "/v1/rootkeys/missing", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateRootKeyInvalidJSON(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodPut, "/v1/rootkeys/rk", strings.NewReader("{bad"))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUpdateRootKeyIDMismatch(t *testing.T) {
	env := newTestEnv(t)
	rk := rootkeys.RootKey{ID: "other", APIKey: "k"}
	body, _ := json.Marshal(rk)
	req := httptest.NewRequest(http.MethodPut, "/v1/rootkeys/rk", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "id mismatch" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestUpdateRootKeyNotFound(t *testing.T) {
	env := newTestEnv(t)
	rk := rootkeys.RootKey{ID: "missing", APIKey: "k"}
	body, _ := json.Marshal(rk)
	req := httptest.NewRequest(http.MethodPut, "/v1/rootkeys/missing", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── services ─────────────────────────────────────────────────────────────────

func TestCreateServiceInvalidJSON(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/services", strings.NewReader("{bad"))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateServiceRootKeyNotFound(t *testing.T) {
	env := newTestEnv(t)
	svc := services.Service{ID: "svc", Endpoint: "http://x.com", RootKeyID: "missing"}
	body, _ := json.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/v1/services", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "root key not found" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestCreateServiceDuplicate(t *testing.T) {
	env := newTestEnv(t)
	rk := rootkeys.RootKey{ID: "rk-svc-dup", APIKey: "k"}
	env.Server.RootKeyStore.Create(rk)
	svc := services.Service{ID: "svc-dup", Endpoint: "http://x.com", RootKeyID: rk.ID}
	env.Server.ServiceStore.Create(svc)
	body, _ := json.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/v1/services", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "service already exists" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestDeleteServiceNotFound(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodDelete, "/v1/services/missing", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── orgs ─────────────────────────────────────────────────────────────────────

func TestCreateOrgInvalidJSON(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/orgs", strings.NewReader("{bad"))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateOrgMissingName(t *testing.T) {
	env := newTestEnv(t)
	body, _ := json.Marshal(orgs.Organization{ID: "o1"})
	req := httptest.NewRequest(http.MethodPost, "/v1/orgs", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestDeleteOrgNotFound(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodDelete, "/v1/orgs/missing", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestListOrgMembersOrgNotFound(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/orgs/missing/members", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestAddOrgMemberOrgNotFound(t *testing.T) {
	env := newTestEnv(t)
	m := orgs.Membership{UserID: env.User.ID, Role: orgs.RoleMember}
	body, _ := json.Marshal(m)
	req := httptest.NewRequest(http.MethodPost, "/v1/orgs/missing/members", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestAddOrgMemberInvalidJSON(t *testing.T) {
	env := newTestEnv(t)
	o := orgs.Organization{ID: "o-json", Name: "OrgJ", Domain: "x.com", Email: "x@x.com"}
	env.Server.OrgStore.Create(o)
	req := httptest.NewRequest(http.MethodPost, "/v1/orgs/o-json/members", strings.NewReader("{bad"))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAddOrgMemberMissingUserID(t *testing.T) {
	env := newTestEnv(t)
	o := orgs.Organization{ID: "o-noid", Name: "OrgN", Domain: "x.com", Email: "x@x.com"}
	env.Server.OrgStore.Create(o)
	body, _ := json.Marshal(orgs.Membership{Role: orgs.RoleMember})
	req := httptest.NewRequest(http.MethodPost, "/v1/orgs/o-noid/members", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAddOrgMemberInvalidRole(t *testing.T) {
	env := newTestEnv(t)
	o := orgs.Organization{ID: "o-role", Name: "OrgR", Domain: "x.com", Email: "x@x.com"}
	env.Server.OrgStore.Create(o)
	body, _ := json.Marshal(orgs.Membership{UserID: env.User.ID, Role: "superadmin"})
	req := httptest.NewRequest(http.MethodPost, "/v1/orgs/o-role/members", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "invalid role" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestAddOrgMemberDuplicate(t *testing.T) {
	env := newTestEnv(t)
	o := orgs.Organization{ID: "o-dup", Name: "OrgD", Domain: "x.com", Email: "x@x.com"}
	env.Server.OrgStore.Create(o)
	env.Server.MembershipStore.Create(orgs.Membership{UserID: env.User.ID, OrgID: o.ID, Role: orgs.RoleMember})
	body, _ := json.Marshal(orgs.Membership{UserID: env.User.ID, Role: orgs.RoleMember})
	req := httptest.NewRequest(http.MethodPost, "/v1/orgs/o-dup/members", bytes.NewReader(body))
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestRemoveOrgMemberNotFound(t *testing.T) {
	env := newTestEnv(t)
	req := httptest.NewRequest(http.MethodDelete, "/v1/orgs/o5/members/nobody", nil)
	env.Authorize(req)
	rr := httptest.NewRecorder()
	env.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── proxy additional paths ────────────────────────────────────────────────────

func TestProxyServiceNotFound(t *testing.T) {
	s := newTestServer(t)
	// key points to a service that doesn't exist
	k := keys.VirtualKey{ID: "vk-nosvc", Target: "ghost-svc", Scope: keys.ScopeRead, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 100}
	s.KeyStore.Create(k)

	router := setupRouter(s)
	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/test", nil)
	req.Header.Set("X-Virtual-Key", k.ID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	if msg := errorBody(t, rr); msg != "service not found" {
		t.Fatalf("unexpected error: %s", msg)
	}
}

func TestProxyRootKeyNotFound(t *testing.T) {
	s := newTestServer(t)
	// service references a root key that doesn't exist
	svc := services.Service{ID: "svc-nork", Endpoint: "http://x.com", RootKeyID: "ghost-rk"}
	s.ServiceStore.Create(svc)
	k := keys.VirtualKey{ID: "vk-nork", Target: svc.ID, Scope: keys.ScopeRead, ExpiresAt: time.Now().Add(time.Hour), RateLimit: 100}
	s.KeyStore.Create(k)

	router := setupRouter(s)
	req := httptest.NewRequest(http.MethodGet, "/v1/proxy/test", nil)
	req.Header.Set("X-Virtual-Key", k.ID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
