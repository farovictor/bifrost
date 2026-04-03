package middlewares

import (
	"net/http"
	"strings"

	"github.com/farovictor/bifrost/config"
	"github.com/farovictor/bifrost/pkg/users"
)

// AuthMiddleware validates the API key provided by the client.
//
// In test or sqlite modes, authentication is performed using the static API
// key from config.StaticAPIKey(), and user lookups are skipped.
func AuthMiddleware(store users.Store) func(http.Handler) http.Handler {
	bypass := config.Mode() == "test" || config.DBType() == "sqlite"
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				auth := r.Header.Get("Authorization")
				if strings.HasPrefix(auth, "Bearer ") {
					key = strings.TrimPrefix(auth, "Bearer ")
				} else {
					key = auth
				}
			}
			if bypass {
				staticKey := config.StaticAPIKey()
				if staticKey != "" && key != staticKey {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
				return
			}
			if key == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if _, err := store.GetByAPIKey(key); err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
