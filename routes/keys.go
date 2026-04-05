package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/usage"
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
	if k.ID == "" || k.Target == "" {
		writeError(w, "id and target are required", http.StatusBadRequest)
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

// usageListResponse is the envelope returned by the usage list endpoint.
type usageListResponse struct {
	Events []usage.Event `json:"events"`
	Total  int64         `json:"total"`
}

// ListKeyUsage handles GET /keys/{id}/usage. Returns paginated usage events for
// the given virtual key, optionally filtered by ?from= and ?to= (RFC3339).
//
// @Summary      List usage events for a virtual key
// @Tags         virtual-keys
// @Produce      json
// @Param        id       path      string  true  "Virtual key ID"
// @Param        from     query     string  false "Start time (RFC3339)"
// @Param        to       query     string  false "End time (RFC3339)"
// @Param        page     query     int     false "Page number (default 1)"
// @Param        per_page query     int     false "Page size (default 20)"
// @Success      200  {object}  usageListResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/keys/{id}/usage [get]
func (s *Server) ListKeyUsage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := s.KeyStore.Get(id); err != nil {
		if err == keys.ErrKeyNotFound {
			writeError(w, "not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	var from, to time.Time
	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, "invalid from: use RFC3339", http.StatusBadRequest)
			return
		}
		from = t
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, "invalid to: use RFC3339", http.StatusBadRequest)
			return
		}
		to = t
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	events, total, err := s.UsageStore.List(id, from, to, page, perPage)
	if err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if events == nil {
		events = []usage.Event{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usageListResponse{Events: events, Total: total})
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
