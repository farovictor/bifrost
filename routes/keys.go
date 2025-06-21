package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/FokusInternal/bifrost/pkg/keys"
	"github.com/FokusInternal/bifrost/pkg/logging"
	"github.com/FokusInternal/bifrost/pkg/services"
)

// KeyStore holds the active VirtualKeys in memory.
var KeyStore = keys.NewStore()

// CreateKey handles POST /keys and stores a new VirtualKey.
func CreateKey(w http.ResponseWriter, r *http.Request) {
	var k keys.VirtualKey
	if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if !keys.ValidateScope(k.Scope) {
		http.Error(w, "invalid scope", http.StatusBadRequest)
		return
	}
	if k.RateLimit <= 0 {
		http.Error(w, "invalid rate_limit", http.StatusBadRequest)
		return
	}
	if !k.ExpiresAt.After(time.Now()) {
		http.Error(w, "expires_at must be in the future", http.StatusBadRequest)
		return
	}
	if _, err := ServiceStore.Get(k.Target); err != nil {
		if err == services.ErrServiceNotFound {
			http.Error(w, "service not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := KeyStore.Create(k); err != nil {
		switch err {
		case keys.ErrKeyExists:
			http.Error(w, "key already exists", http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("key_id", k.ID).Msg("created key")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(k)
}

// DeleteKey handles DELETE /keys/{id} and removes a VirtualKey.
func DeleteKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := KeyStore.Delete(id); err != nil {
		switch err {
		case keys.ErrKeyNotFound:
			http.Error(w, "not found", http.StatusNotFound)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("key_id", id).Msg("deleted key")
	w.WriteHeader(http.StatusNoContent)
}
