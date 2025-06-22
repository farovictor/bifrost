package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	rl "github.com/farovictor/bifrost/middlewares"
	"github.com/go-chi/chi/v5"

	"github.com/farovictor/bifrost/pkg/auth"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/users"
	routes "github.com/farovictor/bifrost/routes"
)

func setupOrgCtxRouter() http.Handler {
	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(rl.AuthMiddleware())
		r.Use(rl.OrgCtxMiddleware())
		r.Post("/users", routes.CreateUser)
		r.Get("/ctx", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(rl.OrgFromContext(r.Context()))
		})
	})
	return r
}

func setupCtxRouter() http.Handler {
	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(rl.AuthMiddleware())
		r.Use(rl.OrgCtxMiddleware())
		r.Get("/ctx", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(rl.OrgFromContext(r.Context()))
		})
	})
	return r
}

func TestUserCreationOrgContext(t *testing.T) {
	cases := []struct {
		name    string
		orgName string
		orgID   string
	}{
		{name: "org_name", orgName: "Acme"},
		{name: "org_id", orgID: "existing"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			routes.UserStore = users.NewMemoryStore()
			routes.OrgStore = orgs.NewMemoryStore()
			routes.MembershipStore = orgs.NewMembershipStore()

			admin := users.User{ID: "admin", APIKey: "admink"}
			routes.UserStore.Create(admin)

			if tc.orgID != "" {
				routes.OrgStore.Create(orgs.Organization{ID: tc.orgID, Name: "Existing Org"})
			}

			router := setupOrgCtxRouter()

			payload := map[string]string{"id": "new"}
			if tc.orgName != "" {
				payload["org_name"] = tc.orgName
			}
			if tc.orgID != "" {
				payload["org_id"] = tc.orgID
			}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
			req.Header.Set("X-API-Key", admin.APIKey)
			req.Header.Set("Authorization", "Bearer "+makeToken(admin.ID))
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusCreated {
				t.Fatalf("expected 201, got %d", rr.Code)
			}

			var resp struct {
				users.User
				Token string `json:"token"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp.Token == "" {
				t.Fatalf("missing token")
			}

			tok, err := auth.Verify(resp.Token)
			if err != nil {
				t.Fatalf("verify token: %v", err)
			}

			org, err := routes.OrgStore.Get(tok.OrgID)
			if err != nil {
				t.Fatalf("org not stored: %v", err)
			}
			if tc.orgName != "" && org.Name != tc.orgName {
				t.Fatalf("unexpected org name %s", org.Name)
			}
			if tc.orgID != "" && org.ID != tc.orgID {
				t.Fatalf("unexpected org id %s", org.ID)
			}

			mem, err := routes.MembershipStore.Get(resp.ID, tok.OrgID)
			if err != nil {
				t.Fatalf("membership: %v", err)
			}
			if mem.Role != orgs.RoleMember {
				t.Fatalf("unexpected role %s", mem.Role)
			}

			req2 := httptest.NewRequest(http.MethodGet, "/v1/ctx", nil)
			req2.Header.Set("X-API-Key", resp.APIKey)
			req2.Header.Set("Authorization", "Bearer "+resp.Token)
			rr2 := httptest.NewRecorder()
			router.ServeHTTP(rr2, req2)

			if rr2.Code != http.StatusOK {
				t.Fatalf("ctx status %d", rr2.Code)
			}
			var ctxResp rl.OrgContext
			if err := json.Unmarshal(rr2.Body.Bytes(), &ctxResp); err != nil {
				t.Fatalf("decode ctx: %v", err)
			}
			if ctxResp.UserID != resp.ID || ctxResp.OrgID != tok.OrgID || ctxResp.Role != mem.Role {
				t.Fatalf("unexpected ctx %#v", ctxResp)
			}
		})
	}
}

func TestOrgCtxMiddlewareFailures(t *testing.T) {
	u := users.User{ID: "u1", APIKey: "secret"}
	o := orgs.Organization{ID: "o1", Name: "Org"}

	cases := []struct {
		name  string
		token string
		want  int
		role  string
	}{
		{name: "invalid token", token: "bad", want: http.StatusUnauthorized},
		{
			name: "expired token",
			token: func() string {
				tkn, _ := auth.Sign(auth.AuthToken{UserID: u.ID, OrgID: o.ID, ExpiresAt: time.Now().Add(-time.Hour)})
				return tkn
			}(),
			want: http.StatusUnauthorized,
		},
		{
			name: "missing membership",
			token: func() string {
				tkn, _ := auth.Sign(auth.AuthToken{UserID: u.ID, OrgID: o.ID, ExpiresAt: time.Now().Add(time.Hour)})
				return tkn
			}(),
			want: http.StatusOK,
			role: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			routes.UserStore = users.NewMemoryStore()
			routes.OrgStore = orgs.NewMemoryStore()
			routes.MembershipStore = orgs.NewMembershipStore()
			routes.UserStore.Create(u)
			routes.OrgStore.Create(o)

			router := setupCtxRouter()

			req := httptest.NewRequest(http.MethodGet, "/v1/ctx", nil)
			req.Header.Set("X-API-Key", u.APIKey)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, rr.Code)
			}

			if tc.want == http.StatusOK {
				var ctxResp rl.OrgContext
				if err := json.Unmarshal(rr.Body.Bytes(), &ctxResp); err != nil {
					t.Fatalf("decode ctx: %v", err)
				}
				if ctxResp.UserID != u.ID || ctxResp.OrgID != o.ID || ctxResp.Role != tc.role {
					t.Fatalf("unexpected ctx %#v", ctxResp)
				}
			}
		})
	}
}
