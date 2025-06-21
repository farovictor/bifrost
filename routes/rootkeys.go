package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FokusInternal/bifrost/pkg/rootkeys"
)

// RootKeyStore holds active root keys in memory.
var RootKeyStore = rootkeys.NewStore()

// CreateRootKey handles POST /rootkeys to store a new root key.
func CreateRootKey(w http.ResponseWriter, r *http.Request) {
	var k rootkeys.RootKey
	if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := RootKeyStore.Create(k); err != nil {
		switch err {
		case rootkeys.ErrKeyExists:
			http.Error(w, "root key already exists", http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(k)
}

// DeleteRootKey handles DELETE /rootkeys/{id} to remove a root key.
func DeleteRootKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := RootKeyStore.Delete(id); err != nil {
		switch err {
		case rootkeys.ErrKeyNotFound:
			http.Error(w, "not found", http.StatusNotFound)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
