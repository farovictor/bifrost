package tests

import (
	"net/http"
	"testing"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/users"
	routes "github.com/farovictor/bifrost/routes"
)

// newTestServer returns a Server wired with empty in-memory stores.
func newTestServer(t *testing.T) *routes.Server {
	t.Helper()
	return &routes.Server{
		UserStore:       users.NewMemoryStore(),
		KeyStore:        keys.NewMemoryStore(),
		RootKeyStore:    rootkeys.NewMemoryStore(),
		ServiceStore:    services.NewMemoryStore(),
		OrgStore:        orgs.NewMemoryStore(),
		MembershipStore: orgs.NewMemoryMembershipStore(),
	}
}

// TestEnv bundles everything a management-endpoint test needs: a server with
// in-memory stores, a pre-built router, a seeded user, and a signed auth token.
type TestEnv struct {
	Server *routes.Server
	Router http.Handler
	User   users.User
	Token  string
}

// Authorize sets the API key and bearer token headers on req.
func (e *TestEnv) Authorize(req *http.Request) {
	req.Header.Set("X-API-Key", e.User.APIKey)
	req.Header.Set("Authorization", "Bearer "+e.Token)
}

// newTestEnv creates a TestEnv with a default user already seeded.
// Use this for tests that call management endpoints requiring auth.
func newTestEnv(t *testing.T) *TestEnv {
	t.Helper()
	s := newTestServer(t)
	u := users.User{ID: "u", Name: "U", Email: "u@example.com", APIKey: "secret"}
	if err := s.UserStore.Create(u); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return &TestEnv{
		Server: s,
		Router: setupRouter(s),
		User:   u,
		Token:  makeToken(u.ID),
	}
}
