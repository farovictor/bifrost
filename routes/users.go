package routes

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/farovictor/bifrost/pkg/users"
)

// UserStore provides access to user storage.
var UserStore users.Store = users.NewStore()

// CreateUser handles POST /users and generates an API key.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.ID == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
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
	logging.Logger.Info().Str("user_id", u.ID).Msg("created user")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func generateKey() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
