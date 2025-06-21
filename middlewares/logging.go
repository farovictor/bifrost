package middlewares

import (
	"net/http"
	"time"

	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/go-chi/chi/v5/middleware"
)

// LoggingMiddleware records method, path, status and duration of each request.
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			logging.Logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", ww.Status()).
				Dur("duration", time.Since(start)).
				Msg("request")
		})
	}
}
