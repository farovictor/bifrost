package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/utils"
)

// CreateOrg handles POST /orgs.
//
// @Summary      Create organization
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        body  body      orgs.Organization  true  "Organization to create"
// @Success      201   {object}  orgs.Organization
// @Failure      400   {object}  ErrorResponse  "missing name"
// @Failure      409   {object}  ErrorResponse  "organization already exists"
// @Failure      500   {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/orgs [post]
func (s *Server) CreateOrg(w http.ResponseWriter, r *http.Request) {
	var o orgs.Organization
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if o.Name == "" {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if o.ID == "" {
		o.ID = utils.GenerateID()
	}
	if err := s.OrgStore.Create(o); err != nil {
		switch err {
		case orgs.ErrOrgExists:
			writeError(w, "organization already exists", http.StatusConflict)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("org_id", o.ID).Msg("created org")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(o)
}

// ListOrgs handles GET /orgs.
//
// @Summary      List organizations
// @Tags         organizations
// @Produce      json
// @Success      200  {array}   orgs.Organization
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/orgs [get]
func (s *Server) ListOrgs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.OrgStore.List())
}

// GetOrg handles GET /orgs/{id}.
//
// @Summary      Get organization
// @Tags         organizations
// @Produce      json
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  orgs.Organization
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/orgs/{id} [get]
func (s *Server) GetOrg(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	o, err := s.OrgStore.Get(id)
	if err != nil {
		switch err {
		case orgs.ErrOrgNotFound:
			writeError(w, "not found", http.StatusNotFound)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}

// DeleteOrg handles DELETE /orgs/{id}.
//
// @Summary      Delete organization
// @Tags         organizations
// @Param        id   path      string  true  "Organization ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/orgs/{id} [delete]
func (s *Server) DeleteOrg(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.OrgStore.Delete(id); err != nil {
		switch err {
		case orgs.ErrOrgNotFound:
			writeError(w, "not found", http.StatusNotFound)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("org_id", id).Msg("deleted org")
	w.WriteHeader(http.StatusNoContent)
}

// ListOrgMembers handles GET /orgs/{id}/members.
//
// @Summary      List organization members
// @Tags         organizations
// @Produce      json
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {array}   orgs.Membership
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/orgs/{id}/members [get]
func (s *Server) ListOrgMembers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := s.OrgStore.Get(id); err != nil {
		if err == orgs.ErrOrgNotFound {
			writeError(w, "not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.MembershipStore.ListByOrg(id))
}

// AddOrgMember handles POST /orgs/{id}/members.
//
// @Summary      Add member to organization
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        id    path      string          true  "Organization ID"
// @Param        body  body      orgs.Membership true  "Membership to create"
// @Success      201   {object}  orgs.Membership
// @Failure      400   {object}  ErrorResponse  "missing user_id or invalid role"
// @Failure      404   {object}  ErrorResponse  "organization not found"
// @Failure      409   {object}  ErrorResponse  "membership already exists"
// @Failure      500   {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/orgs/{id}/members [post]
func (s *Server) AddOrgMember(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "id")
	if _, err := s.OrgStore.Get(orgID); err != nil {
		if err == orgs.ErrOrgNotFound {
			writeError(w, "not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	var m orgs.Membership
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if m.UserID == "" {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	m.OrgID = orgID
	if m.Role == "" {
		m.Role = orgs.RoleMember
	}
	if !orgs.ValidateRole(m.Role) {
		writeError(w, "invalid role", http.StatusBadRequest)
		return
	}

	if err := s.MembershipStore.Create(m); err != nil {
		switch err {
		case orgs.ErrMembershipExists:
			writeError(w, "membership already exists", http.StatusConflict)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("org_id", orgID).Str("user_id", m.UserID).Msg("added member")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

// RemoveOrgMember handles DELETE /orgs/{id}/members/{userID}.
//
// @Summary      Remove member from organization
// @Tags         organizations
// @Param        id      path      string  true  "Organization ID"
// @Param        userID  path      string  true  "User ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Security     ApiKeyAuth
// @Security     BearerAuth
// @Router       /v1/orgs/{id}/members/{userID} [delete]
func (s *Server) RemoveOrgMember(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userID")
	if err := s.MembershipStore.Delete(userID, orgID); err != nil {
		switch err {
		case orgs.ErrMembershipNotFound:
			writeError(w, "not found", http.StatusNotFound)
		default:
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("org_id", orgID).Str("user_id", userID).Msg("removed member")
	w.WriteHeader(http.StatusNoContent)
}
