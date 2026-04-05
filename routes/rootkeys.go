package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/farovictor/bifrost/pkg/rootkeys"
)

// CreateRootKey handles POST /rootkeys to store a new root key.
//
// @Summary      Create root key
// @Tags         root-keys
// @Accept       json
// @Produce      json
// @Param        body  body      rootkeys.RootKey  true  "Root key to create"
// @Success      201   {object}  rootkeys.RootKey
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse  "root key already exists"
// @Failure      500   {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/rootkeys [post]
func (s *Server) CreateRootKey(w http.ResponseWriter, r *http.Request) {
	var k rootkeys.RootKey
	if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if k.ID == "" || k.APIKey == "" {
		writeError(w, "id and api_key are required", http.StatusBadRequest)
		return
	}
	// Capture hint before the store encrypts and clears APIKey.
	plaintext := k.APIKey
	if err := s.RootKeyStore.Create(k); err != nil {
		switch err {
		case rootkeys.ErrKeyExists:
			writeError(w, "root key already exists", http.StatusConflict)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("root_key_id", k.ID).Msg("created root key")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Return api_key once so the caller can verify — it will not appear again.
	json.NewEncoder(w).Encode(struct {
		ID     string `json:"id"`
		APIKey string `json:"api_key"`
	}{ID: k.ID, APIKey: plaintext})
}

// ListRootKeys handles GET /rootkeys and returns all root keys.
//
// @Summary      List root keys
// @Tags         root-keys
// @Produce      json
// @Success      200  {array}   rootkeys.RootKey
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/rootkeys [get]
func (s *Server) ListRootKeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.RootKeyStore.List())
}

// DeleteRootKey handles DELETE /rootkeys/{id} to remove a root key.
//
// @Summary      Delete root key
// @Tags         root-keys
// @Param        id   path      string  true  "Root key ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/rootkeys/{id} [delete]
func (s *Server) DeleteRootKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.RootKeyStore.Delete(id); err != nil {
		switch err {
		case rootkeys.ErrKeyNotFound:
			writeError(w, "not found", http.StatusNotFound)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("root_key_id", id).Msg("deleted root key")
	w.WriteHeader(http.StatusNoContent)
}

// UpdateRootKey handles PUT /rootkeys/{id} to replace a stored root key.
//
// @Summary      Update root key
// @Tags         root-keys
// @Accept       json
// @Produce      json
// @Param        id    path      string            true  "Root key ID"
// @Param        body  body      rootkeys.RootKey  true  "Updated root key"
// @Success      200   {object}  rootkeys.RootKey
// @Failure      400   {object}  ErrorResponse  "invalid request or id mismatch"
// @Failure      404   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/rootkeys/{id} [put]
func (s *Server) UpdateRootKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var k rootkeys.RootKey
	if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if k.ID == "" {
		k.ID = id
	} else if k.ID != id {
		writeError(w, "id mismatch", http.StatusBadRequest)
		return
	}
	if err := s.RootKeyStore.Update(k); err != nil {
		switch err {
		case rootkeys.ErrKeyNotFound:
			writeError(w, "not found", http.StatusNotFound)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("root_key_id", k.ID).Msg("updated root key")
	w.WriteHeader(http.StatusNoContent)
}
