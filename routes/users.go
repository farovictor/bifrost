package routes

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/farovictor/bifrost/pkg/auth"
	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/users"
)

// UserStore holds registered users in memory.
var UserStore = users.NewStore()

// CreateUser handles POST /users and generates an API key.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID      string `json:"id"`
		OrgID   string `json:"org_id"`
		OrgName string `json:"org_name"`
		Role    string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.ID == "" {
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

	u := users.User{ID: req.ID, APIKey: generateKey()}
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
		o := orgs.Organization{ID: generateKey(), Name: req.OrgName}
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

func generateKey() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func buildAuthToken(userID, orgID string) (string, error) {
	t := auth.AuthToken{
		UserID:    userID,
		OrgID:     orgID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	return auth.Sign(t)
}
