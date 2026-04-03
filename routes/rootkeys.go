package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/farovictor/bifrost/pkg/rootkeys"
)

// CreateRootKey handles POST /rootkeys to store a new root key.
func (s *Server) CreateRootKey(w http.ResponseWriter, r *http.Request) {
	var k rootkeys.RootKey
	if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := s.RootKeyStore.Create(k); err != nil {
		switch err {
		case rootkeys.ErrKeyExists:
			http.Error(w, "root key already exists", http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("root_key_id", k.ID).Msg("created root key")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(k)
}

// DeleteRootKey handles DELETE /rootkeys/{id} to remove a root key.
func (s *Server) DeleteRootKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.RootKeyStore.Delete(id); err != nil {
		switch err {
		case rootkeys.ErrKeyNotFound:
			http.Error(w, "not found", http.StatusNotFound)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("root_key_id", id).Msg("deleted root key")
	w.WriteHeader(http.StatusNoContent)
}

// UpdateRootKey handles PUT /rootkeys/{id} to replace a stored root key.
func (s *Server) UpdateRootKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var k rootkeys.RootKey
	if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if k.ID == "" {
		k.ID = id
	} else if k.ID != id {
		http.Error(w, "id mismatch", http.StatusBadRequest)
		return
	}
	if err := s.RootKeyStore.Update(k); err != nil {
		switch err {
		case rootkeys.ErrKeyNotFound:
			http.Error(w, "not found", http.StatusNotFound)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("root_key_id", k.ID).Msg("updated root key")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(k)
}
