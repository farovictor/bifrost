package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/farovictor/bifrost/pkg/services"
)

// CreateKey handles POST /keys and stores a new VirtualKey.
//
// @Summary      Create virtual key
// @Tags         virtual-keys
// @Accept       json
// @Produce      json
// @Param        body  body      keys.VirtualKey  true  "Virtual key to create"
// @Success      201   {object}  keys.VirtualKey
// @Failure      400   {object}  ErrorResponse  "invalid scope, rate_limit, or expires_at"
// @Failure      404   {object}  ErrorResponse  "service not found"
// @Failure      409   {object}  ErrorResponse  "key already exists"
// @Failure      500   {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/keys [post]
func (s *Server) CreateKey(w http.ResponseWriter, r *http.Request) {
	var k keys.VirtualKey
	if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if !keys.ValidateScope(k.Scope) {
		writeError(w, "invalid scope", http.StatusBadRequest)
		return
	}
	if k.RateLimit <= 0 {
		writeError(w, "invalid rate_limit", http.StatusBadRequest)
		return
	}
	if !k.ExpiresAt.After(time.Now()) {
		writeError(w, "expires_at must be in the future", http.StatusBadRequest)
		return
	}
	if _, err := s.ServiceStore.Get(k.Target); err != nil {
		if err == services.ErrServiceNotFound {
			writeError(w, "service not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := s.KeyStore.Create(k); err != nil {
		switch err {
		case keys.ErrKeyExists:
			writeError(w, "key already exists", http.StatusConflict)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("key_id", k.ID).Msg("created key")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(k)
}

// ListKeys handles GET /keys and returns all VirtualKeys.
//
// @Summary      List virtual keys
// @Tags         virtual-keys
// @Produce      json
// @Success      200  {array}   keys.VirtualKey
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/keys [get]
func (s *Server) ListKeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.KeyStore.List())
}

// DeleteKey handles DELETE /keys/{id} and removes a VirtualKey.
//
// @Summary      Revoke virtual key
// @Tags         virtual-keys
// @Param        id   path      string  true  "Virtual key ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/keys/{id} [delete]
func (s *Server) DeleteKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.KeyStore.Delete(id); err != nil {
		switch err {
		case keys.ErrKeyNotFound:
			writeError(w, "not found", http.StatusNotFound)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("key_id", id).Msg("deleted key")
	w.WriteHeader(http.StatusNoContent)
}
