package v1

import (
	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
)

// Handler holds store dependencies for v1 route handlers.
type Handler struct {
	KeyStore     keys.Store
	ServiceStore services.Store
	RootKeyStore rootkeys.Store
}
