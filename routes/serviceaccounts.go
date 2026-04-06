package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/farovictor/bifrost/pkg/serviceaccounts"
	"github.com/farovictor/bifrost/pkg/utils"
)

// CreateServiceAccount handles POST /v1/serviceaccounts.
//
// @Summary      Create service account
// @Tags         service-accounts
// @Accept       json
// @Produce      json
// @Param        body  body      serviceaccounts.ServiceAccount  true  "Service account to create"
// @Success      201   {object}  serviceaccounts.ServiceAccount
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse  "service account already exists"
// @Failure      500   {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/serviceaccounts [post]
func (s *Server) CreateServiceAccount(w http.ResponseWriter, r *http.Request) {
	var sa serviceaccounts.ServiceAccount
	if err := json.NewDecoder(r.Body).Decode(&sa); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if sa.Name == "" {
		writeError(w, "name is required", http.StatusBadRequest)
		return
	}
	if sa.ID == "" {
		sa.ID = utils.GenerateID()
	}
	if sa.APIKey == "" {
		sa.APIKey = "sa-" + utils.GenerateID()
	}
	if sa.AllowedServices == nil {
		sa.AllowedServices = serviceaccounts.StringList{}
	}

	if err := s.ServiceAccountStore.Create(sa); err != nil {
		switch err {
		case serviceaccounts.ErrServiceAccountExists:
			writeError(w, "service account already exists", http.StatusConflict)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sa)
}

// ListServiceAccounts handles GET /v1/serviceaccounts.
//
// @Summary      List service accounts
// @Tags         service-accounts
// @Produce      json
// @Success      200  {array}   serviceaccounts.ServiceAccount
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/serviceaccounts [get]
func (s *Server) ListServiceAccounts(w http.ResponseWriter, r *http.Request) {
	list := s.ServiceAccountStore.List()
	if list == nil {
		list = []serviceaccounts.ServiceAccount{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// DeleteServiceAccount handles DELETE /v1/serviceaccounts/{id}.
//
// @Summary      Delete service account
// @Tags         service-accounts
// @Param        id   path      string  true  "Service account ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/serviceaccounts/{id} [delete]
func (s *Server) DeleteServiceAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.ServiceAccountStore.Delete(id); err != nil {
		switch err {
		case serviceaccounts.ErrServiceAccountNotFound:
			writeError(w, "not found", http.StatusNotFound)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
