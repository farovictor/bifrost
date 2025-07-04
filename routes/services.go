package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
)

// ServiceStore provides access to stored services.
var ServiceStore services.Store

// CreateService handles POST /services to store a new Service.
func CreateService(w http.ResponseWriter, r *http.Request) {
	var s services.Service
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if _, err := RootKeyStore.Get(s.RootKeyID); err != nil {
		if err == rootkeys.ErrKeyNotFound {
			http.Error(w, "root key not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := ServiceStore.Create(s); err != nil {
		switch err {
		case services.ErrServiceExists:
			http.Error(w, "service already exists", http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("service_id", s.ID).Msg("created service")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

// DeleteService handles DELETE /services/{id} to remove a service.
func DeleteService(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := ServiceStore.Delete(id); err != nil {
		switch err {
		case services.ErrServiceNotFound:
			http.Error(w, "not found", http.StatusNotFound)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("service_id", id).Msg("deleted service")
	w.WriteHeader(http.StatusNoContent)
}
