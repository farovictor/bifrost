package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
)

// CreateService handles POST /services to store a new Service.
func (s *Server) CreateService(w http.ResponseWriter, r *http.Request) {
	var svc services.Service
	if err := json.NewDecoder(r.Body).Decode(&svc); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if _, err := s.RootKeyStore.Get(svc.RootKeyID); err != nil {
		if err == rootkeys.ErrKeyNotFound {
			writeError(w, "root key not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := s.ServiceStore.Create(svc); err != nil {
		switch err {
		case services.ErrServiceExists:
			writeError(w, "service already exists", http.StatusConflict)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("service_id", svc.ID).Msg("created service")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(svc)
}

// ListServices handles GET /services and returns all services.
func (s *Server) ListServices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.ServiceStore.List())
}

// DeleteService handles DELETE /services/{id} to remove a service.
func (s *Server) DeleteService(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.ServiceStore.Delete(id); err != nil {
		switch err {
		case services.ErrServiceNotFound:
			writeError(w, "not found", http.StatusNotFound)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("service_id", id).Msg("deleted service")
	w.WriteHeader(http.StatusNoContent)
}
