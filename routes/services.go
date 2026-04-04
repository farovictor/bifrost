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
//
// @Summary      Create service
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        body  body      services.Service  true  "Service to create"
// @Success      201   {object}  services.Service
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse  "root key not found"
// @Failure      409   {object}  ErrorResponse  "service already exists"
// @Failure      500   {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/services [post]
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
//
// @Summary      List services
// @Tags         services
// @Produce      json
// @Success      200  {array}   services.Service
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/services [get]
func (s *Server) ListServices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.ServiceStore.List())
}

// UpdateService handles PUT /services/{id} to replace a service.
//
// @Summary      Update service
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        id    path      string            true  "Service ID"
// @Param        body  body      services.Service  true  "Updated service"
// @Success      200   {object}  services.Service
// @Failure      400   {object}  ErrorResponse  "invalid request or id mismatch"
// @Failure      404   {object}  ErrorResponse  "service or root key not found"
// @Failure      500   {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/services/{id} [put]
func (s *Server) UpdateService(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var svc services.Service
	if err := json.NewDecoder(r.Body).Decode(&svc); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if svc.ID != id {
		writeError(w, "id mismatch", http.StatusBadRequest)
		return
	}
	if svc.RootKeyID != "" {
		if _, err := s.RootKeyStore.Get(svc.RootKeyID); err != nil {
			if err == rootkeys.ErrKeyNotFound {
				writeError(w, "root key not found", http.StatusNotFound)
				return
			}
			writeError(w, "internal error", http.StatusInternalServerError)
			return
		}
	}
	if err := s.ServiceStore.Update(svc); err != nil {
		switch err {
		case services.ErrServiceNotFound:
			writeError(w, "not found", http.StatusNotFound)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("service_id", id).Msg("updated service")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(svc)
}

// DeleteService handles DELETE /services/{id} to remove a service.
//
// @Summary      Delete service
// @Tags         services
// @Param        id   path      string  true  "Service ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/services/{id} [delete]
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
