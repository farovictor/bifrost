package routes

import (
	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/users"
)

// Server holds all store dependencies for HTTP handlers.
type Server struct {
	UserStore       users.Store
	KeyStore        keys.Store
	RootKeyStore    rootkeys.Store
	ServiceStore    services.Store
	OrgStore        orgs.Store
	MembershipStore orgs.MembershipStore
}
