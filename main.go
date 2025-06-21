package main

import (
	"context"
	"net/http"

	"github.com/FokusInternal/bifrost/config"
	m "github.com/FokusInternal/bifrost/middlewares"
	routes "github.com/FokusInternal/bifrost/routes"
	v1 "github.com/FokusInternal/bifrost/routes/v1"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", routes.Healthz)
	r.Get("/version", routes.Version)

	r.Route("/v1", func(r chi.Router) {
		r.Use(apiVersionCtx("v1"))
		r.Get("/hello", v1.SayHello)

		r.Post("/keys", routes.CreateKey)
		r.Delete("/keys/{id}", routes.DeleteKey)

		r.With(m.RateLimitMiddleware()).Post("/rate", v1.SayHello)
	})

	http.ListenAndServe(config.ServerPort(), r)
}

func apiVersionCtx(version string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), "api.version", version))
			next.ServeHTTP(w, r)
		})
	}
}
