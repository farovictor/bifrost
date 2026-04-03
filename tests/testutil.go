package tests

import (
	"time"

	"github.com/farovictor/bifrost/pkg/auth"
	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/users"
	routes "github.com/farovictor/bifrost/routes"
)

// newTestServer returns a Server wired with empty in-memory stores.
func newTestServer() *routes.Server {
	return &routes.Server{
		UserStore:       users.NewMemoryStore(),
		KeyStore:        keys.NewMemoryStore(),
		RootKeyStore:    rootkeys.NewMemoryStore(),
		ServiceStore:    services.NewMemoryStore(),
		OrgStore:        orgs.NewMemoryStore(),
		MembershipStore: orgs.NewMemoryMembershipStore(),
	}
}

// makeToken generates a short-lived auth token for a user.
func makeToken(userID string) string {
	t := auth.AuthToken{
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	tok, _ := auth.Sign(t)
	return tok
}
