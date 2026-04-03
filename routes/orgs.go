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
func (s *Server) CreateOrg(w http.ResponseWriter, r *http.Request) {
	var o orgs.Organization
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if o.Name == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if o.ID == "" {
		o.ID = utils.GenerateID()
	}
	if err := s.OrgStore.Create(o); err != nil {
		switch err {
		case orgs.ErrOrgExists:
			http.Error(w, "organization already exists", http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("org_id", o.ID).Msg("created org")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(o)
}

// ListOrgs handles GET /orgs.
func (s *Server) ListOrgs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.OrgStore.List())
}

// GetOrg handles GET /orgs/{id}.
func (s *Server) GetOrg(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	o, err := s.OrgStore.Get(id)
	if err != nil {
		switch err {
		case orgs.ErrOrgNotFound:
			http.Error(w, "not found", http.StatusNotFound)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}

// DeleteOrg handles DELETE /orgs/{id}.
func (s *Server) DeleteOrg(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.OrgStore.Delete(id); err != nil {
		switch err {
		case orgs.ErrOrgNotFound:
			http.Error(w, "not found", http.StatusNotFound)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("org_id", id).Msg("deleted org")
	w.WriteHeader(http.StatusNoContent)
}

// ListOrgMembers handles GET /orgs/{id}/members.
func (s *Server) ListOrgMembers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := s.OrgStore.Get(id); err != nil {
		if err == orgs.ErrOrgNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.MembershipStore.ListByOrg(id))
}

// AddOrgMember handles POST /orgs/{id}/members.
func (s *Server) AddOrgMember(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "id")
	if _, err := s.OrgStore.Get(orgID); err != nil {
		if err == orgs.ErrOrgNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var m orgs.Membership
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if m.UserID == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	m.OrgID = orgID
	if m.Role == "" {
		m.Role = orgs.RoleMember
	}
	if !orgs.ValidateRole(m.Role) {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	if err := s.MembershipStore.Create(m); err != nil {
		switch err {
		case orgs.ErrMembershipExists:
			http.Error(w, "membership already exists", http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("org_id", orgID).Str("user_id", m.UserID).Msg("added member")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

// RemoveOrgMember handles DELETE /orgs/{id}/members/{userID}.
func (s *Server) RemoveOrgMember(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userID")
	if err := s.MembershipStore.Delete(userID, orgID); err != nil {
		switch err {
		case orgs.ErrMembershipNotFound:
			http.Error(w, "not found", http.StatusNotFound)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	logging.Logger.Info().Str("org_id", orgID).Str("user_id", userID).Msg("removed member")
	w.WriteHeader(http.StatusNoContent)
}
