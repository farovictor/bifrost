package middlewares

import (
	"net/http"
	"strings"

	routes "github.com/farovictor/bifrost/routes"
)

// AuthMiddleware validates the API key provided by the client.
func AuthMiddleware() func(http.Handler) http.Handler {
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
			if key == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if _, err := routes.UserStore.GetByAPIKey(key); err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
