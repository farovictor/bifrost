package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/FokusInternal/bifrost/config"
	"github.com/FokusInternal/bifrost/pkg/metrics"
	"github.com/go-chi/chi/v5/middleware"
)

// MetricsMiddleware records request metrics if enabled.
func MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.MetricsEnabled() {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			metrics.RequestTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(ww.Status())).Inc()
			metrics.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
		})
	}
}
