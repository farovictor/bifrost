package routes

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/farovictor/bifrost/pkg/auth"
	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/users"
	"github.com/farovictor/bifrost/pkg/utils"
)

// CreateUserRequest is the payload for POST /v1/users.
type CreateUserRequest struct {
	Name    string `json:"name"     example:"Alice"`
	Email   string `json:"email"    example:"alice@example.com"`
	OrgID   string `json:"org_id"   example:""`
	OrgName string `json:"org_name" example:"Acme"`
	Role    string `json:"role"     example:"member"`
}

// CreateUserResponse is returned on successful user creation.
type CreateUserResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	APIKey string `json:"api_key"`
	Token  string `json:"token"`
}

// CreateUser handles POST /users and generates an API key.
//
// @Summary      Create user
// @Description  Creates a user and optionally joins or creates an organization. Returns the user and a signed bearer token.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      CreateUserRequest  true  "User creation request"
// @Success      201   {object}  CreateUserResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse  "organization not found"
// @Failure      409   {object}  ErrorResponse  "user already exists"
// @Failure      500   {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /v1/users [post]
func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Email == "" {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	role := req.Role
	if role == "" {
		role = orgs.RoleMember
	}
	if !orgs.ValidateRole(role) {
		writeError(w, "invalid role", http.StatusBadRequest)
		return
	}

	existing, err := s.UserStore.GetByEmail(req.Email)
	var u users.User
	if err == nil {
		u = existing
	} else if err == users.ErrUserNotFound {
		u = users.User{ID: utils.GenerateID(), Name: req.Name, Email: req.Email, APIKey: users.GenerateAPIKey()}
		if err := s.UserStore.Create(u); err != nil {
			switch err {
			case users.ErrUserExists:
				writeError(w, "user already exists", http.StatusConflict)
			default:
				writeError(w, "internal error", http.StatusInternalServerError)
			}
			return
		}
	} else {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	var orgID string
	if req.OrgName != "" && req.OrgID == "" {
		o := orgs.Organization{ID: utils.GenerateID(), Name: req.OrgName}
		if err := s.OrgStore.Create(o); err != nil {
			writeError(w, "internal error", http.StatusInternalServerError)
			return
		}
		orgID = o.ID
	} else if req.OrgID != "" {
		if _, err := s.OrgStore.Get(req.OrgID); err != nil {
			if err == orgs.ErrOrgNotFound {
				writeError(w, "organization not found", http.StatusNotFound)
				return
			}
			writeError(w, "internal error", http.StatusInternalServerError)
			return
		}
		orgID = req.OrgID
	}

	if orgID != "" {
		existingMems := s.MembershipStore.ListByUser(u.ID)
		if len(existingMems) > 0 {
			for _, mem := range existingMems {
				if mem.OrgID == orgID {
					writeError(w, "user already exists", http.StatusConflict)
					return
				}
			}
			writeError(w, "user already exists", http.StatusConflict)
			return
		}

		m := orgs.Membership{UserID: u.ID, OrgID: orgID, Role: role}
		if err := s.MembershipStore.Create(m); err != nil {
			writeError(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	token, err := buildAuthToken(u.ID, orgID)
	if err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := struct {
		users.User
		Token string `json:"token"`
	}{User: u, Token: token}

	logging.Logger.Info().Str("user_id", u.ID).Msg("created user")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// RefreshToken handles POST /token/refresh and issues a fresh 24h token.
//
// @Summary      Refresh bearer token
// @Description  Accepts a valid bearer token and returns a new one with a fresh 24h expiry.
// @Tags         users
// @Produce      json
// @Success      200  {object}  object  "New token: {\"token\":\"...\"}"
// @Failure      401  {object}  ErrorResponse  "invalid or expired token"
// @Failure      500  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /v1/token/refresh [post]
func (s *Server) RefreshToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	raw := strings.TrimPrefix(authHeader, "Bearer ")
	tok, err := auth.Verify(raw)
	if err != nil {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	token, err := buildAuthToken(tok.UserID, tok.OrgID)
	if err != nil {
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Token string `json:"token"`
	}{Token: token})
}

func buildAuthToken(userID, orgID string) (string, error) {
	t := auth.AuthToken{
		UserID:    userID,
		OrgID:     orgID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	return auth.Sign(t)
}

// GetUserInfo handles GET /user and returns details about the authenticated user.
//
// @Summary      Get authenticated user
// @Tags         users
// @Produce      json
// @Success      200  {object}  object  "User with orgs array"
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /v1/user [get]
func (s *Server) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	raw := strings.TrimPrefix(authHeader, "Bearer ")
	tok, err := auth.Verify(raw)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("token verification failed")
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	u, err := s.UserStore.Get(tok.UserID)
	if err != nil {
		if err == users.ErrUserNotFound {
			logging.Logger.Warn().Str("user_id", tok.UserID).Msg("user not found")
			writeError(w, "not found", http.StatusNotFound)
		} else {
			logging.Logger.Error().Err(err).Str("user_id", tok.UserID).Msg("get user")
			writeError(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	type orgInfo struct {
		OrgID string `json:"org_id"`
		Name  string `json:"name"`
		Role  string `json:"role"`
	}

	var orgsInfo []orgInfo
	for _, m := range s.MembershipStore.ListByUser(u.ID) {
		if o, err := s.OrgStore.Get(m.OrgID); err == nil {
			orgsInfo = append(orgsInfo, orgInfo{OrgID: o.ID, Name: o.Name, Role: m.Role})
		}
	}

	resp := struct {
		ID    string    `json:"id"`
		Name  string    `json:"name"`
		Email string    `json:"email"`
		Orgs  []orgInfo `json:"orgs,omitempty"`
	}{ID: u.ID, Name: u.Name, Email: u.Email, Orgs: orgsInfo}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
