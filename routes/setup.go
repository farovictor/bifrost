package routes

import (
	"encoding/json"
	"net/http"

	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/users"
	"github.com/farovictor/bifrost/pkg/utils"
)

// SetupRequest is the payload for the one-shot bootstrap endpoint.
type SetupRequest struct {
	Name    string `json:"name"     example:"Admin"`
	Email   string `json:"email"    example:"admin@example.com"`
	OrgName string `json:"org_name" example:"My Org"`
}

// SetupResponse is returned after successful bootstrap.
type SetupResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	APIKey string `json:"api_key"`
	Token  string `json:"token"`
}

// Setup handles POST /v1/setup — one-shot bootstrap that creates the first admin
// user and org. Returns 409 if any user already exists.
//
// @Summary      Bootstrap Bifrost
// @Description  Creates the first admin user and organization. Returns 409 if the system is already initialized. No authentication required.
// @Tags         setup
// @Accept       json
// @Produce      json
// @Param        body  body      SetupRequest   true  "Bootstrap request"
// @Success      201   {object}  SetupResponse
// @Failure      400   {object}  ErrorResponse  "missing required fields"
// @Failure      409   {object}  ErrorResponse  "already initialized"
// @Failure      500   {object}  ErrorResponse
// @Router       /v1/setup [post]
func (s *Server) Setup(w http.ResponseWriter, r *http.Request) {
	count, err := s.UserStore.Count()
	if err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		writeError(w, "already initialized", http.StatusConflict)
		return
	}

	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Email == "" {
		writeError(w, "name and email are required", http.StatusBadRequest)
		return
	}
	if req.OrgName == "" {
		req.OrgName = "Default"
	}

	u := users.User{
		ID:     utils.GenerateID(),
		Name:   req.Name,
		Email:  req.Email,
		APIKey: users.GenerateAPIKey(),
	}
	if err := s.UserStore.Create(u); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	o := orgs.Organization{ID: utils.GenerateID(), Name: req.OrgName}
	if err := s.OrgStore.Create(o); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	m := orgs.Membership{UserID: u.ID, OrgID: o.ID, Role: orgs.RoleOwner}
	if err := s.MembershipStore.Create(m); err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	token, err := buildAuthToken(u.ID, o.ID)
	if err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SetupResponse{
		ID:     u.ID,
		Name:   u.Name,
		Email:  u.Email,
		APIKey: u.APIKey,
		Token:  token,
	})
}
