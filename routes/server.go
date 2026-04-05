package routes

import (
	"encoding/json"
	"net/http"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/usage"
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
	UsageStore      usage.Store
}

// ErrorResponse is the standard error body returned by all endpoints.
type ErrorResponse struct {
	Error string `json:"error" example:"not found"`
}

// writeError writes a JSON {"error":"..."} response with the given status code.
func writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: message})
}
