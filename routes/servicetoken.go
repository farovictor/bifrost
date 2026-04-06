package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/serviceaccounts"
	"github.com/farovictor/bifrost/pkg/services"
)

// servicetokenRequest is the body for POST /v1/service-token.
type servicetokenRequest struct {
	Service     string `json:"service"`
	TTLSeconds  int    `json:"ttl_seconds"`
	RateLimit   int    `json:"rate_limit"`
}

// servicetokenResponse is returned on a successful token request.
type servicetokenResponse struct {
	Key       string    `json:"key"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ServiceToken handles POST /v1/service-token.
// The caller authenticates via the X-Service-Key header using a service account API key.
// It returns a short-lived virtual key scoped to the requested service.
//
// @Summary      Request a virtual key as a service account
// @Tags         service-accounts
// @Accept       json
// @Produce      json
// @Param        X-Service-Key  header    string                 true  "Service account API key"
// @Param        body           body      servicetokenRequest    true  "Token request"
// @Success      200            {object}  servicetokenResponse
// @Failure      400            {object}  ErrorResponse
// @Failure      401            {object}  ErrorResponse  "missing or invalid X-Service-Key"
// @Failure      403            {object}  ErrorResponse  "service not in allowed list"
// @Failure      404            {object}  ErrorResponse  "service not found"
// @Failure      500            {object}  ErrorResponse
// @Router       /v1/service-token [post]
func (s *Server) ServiceToken(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-Service-Key")
	if apiKey == "" {
		writeError(w, "missing X-Service-Key", http.StatusUnauthorized)
		return
	}

	sa, err := s.ServiceAccountStore.GetByAPIKey(apiKey)
	if err != nil {
		if err == serviceaccounts.ErrServiceAccountNotFound {
			writeError(w, "invalid service key", http.StatusUnauthorized)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	var req servicetokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Service == "" {
		writeError(w, "service is required", http.StatusBadRequest)
		return
	}

	// Enforce allowed_services when the list is non-empty.
	if len(sa.AllowedServices) > 0 {
		allowed := false
		for _, svc := range sa.AllowedServices {
			if svc == req.Service {
				allowed = true
				break
			}
		}
		if !allowed {
			writeError(w, "service not allowed for this service account", http.StatusForbidden)
			return
		}
	}

	// Verify the service exists.
	if _, err := s.ServiceStore.Get(req.Service); err != nil {
		if err == services.ErrServiceNotFound {
			writeError(w, "service not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	ttl := req.TTLSeconds
	if ttl <= 0 {
		ttl = 3600 // default 1 hour
	}
	rateLimit := req.RateLimit
	if rateLimit <= 0 {
		rateLimit = 60 // default 60 rpm
	}

	expiresAt := time.Now().Add(time.Duration(ttl) * time.Second)
	k := keys.VirtualKey{
		ID:        fmt.Sprintf("vk-sa-%d", time.Now().UnixNano()),
		Target:    req.Service,
		Scope:     keys.ScopeWrite,
		RateLimit: rateLimit,
		ExpiresAt: expiresAt,
		Source:    keys.SourceServiceAccount,
	}

	if err := s.KeyStore.Create(k); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servicetokenResponse{Key: k.ID, ExpiresAt: expiresAt})
}
