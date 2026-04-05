package v1

import (
	"encoding/json"
	"net/http"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/usage"
)

// injectCredential sets the appropriate header on r based on credentialHeader.
// "" or "X-API-Key" → X-API-Key: <apiKey>
// "Authorization"   → Authorization: Bearer <apiKey>
// anything else     → <credentialHeader>: <apiKey>
func injectCredential(r *http.Request, credentialHeader, apiKey string) {
	switch credentialHeader {
	case "", services.CredentialHeaderXAPIKey:
		r.Header.Set(services.CredentialHeaderXAPIKey, apiKey)
	case services.CredentialHeaderBearer:
		r.Header.Set("Authorization", "Bearer "+apiKey)
	default:
		r.Header.Set(credentialHeader, apiKey)
	}
}

// Handler holds store dependencies for v1 route handlers.
type Handler struct {
	KeyStore     keys.Store
	ServiceStore services.Store
	RootKeyStore rootkeys.Store
	UsageStore   usage.Store
}

func writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: message})
}
