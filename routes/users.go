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

// UserStore provides access to persisted users.
var UserStore users.Store

// CreateUser handles POST /users and generates an API key.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		OrgID   string `json:"org_id"`
		OrgName string `json:"org_name"`
		Role    string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Email == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	role := req.Role
	if role == "" {
		role = orgs.RoleMember
	}
	if !orgs.ValidateRole(role) {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	u := users.User{ID: utils.GenerateID(), Name: req.Name, Email: req.Email, APIKey: users.GenerateAPIKey()}
	if err := UserStore.Create(u); err != nil {
		switch err {
		case users.ErrUserExists:
			http.Error(w, "user already exists", http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	var orgID string
	if req.OrgName != "" && req.OrgID == "" {
		o := orgs.Organization{ID: utils.GenerateID(), Name: req.OrgName}
		if err := OrgStore.Create(o); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		orgID = o.ID
	} else if req.OrgID != "" {
		if _, err := OrgStore.Get(req.OrgID); err != nil {
			if err == orgs.ErrOrgNotFound {
				http.Error(w, "organization not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		orgID = req.OrgID
	}

	if orgID != "" {
		m := orgs.Membership{UserID: u.ID, OrgID: orgID, Role: role}
		if err := MembershipStore.Create(m); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	token, err := buildAuthToken(u.ID, orgID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
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

func buildAuthToken(userID, orgID string) (string, error) {
	t := auth.AuthToken{
		UserID:    userID,
		OrgID:     orgID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	return auth.Sign(t)
}

// GetUserInfo handles GET /user and returns details about the authenticated user.
func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	logging.Logger.Info().Msg("key used")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	raw := strings.TrimPrefix(authHeader, "Bearer ")
	tok, err := auth.Verify(raw)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("verify token")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	u, err := UserStore.Get(tok.UserID)
	if err != nil {
		if err == users.ErrUserNotFound {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	type orgInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	var orgsInfo []orgInfo
	for _, m := range MembershipStore.List() {
		if m.UserID != u.ID {
			continue
		}
		if o, err := OrgStore.Get(m.OrgID); err == nil {
			orgsInfo = append(orgsInfo, orgInfo{ID: o.ID, Name: o.Name})
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
