package main

import (
	"context"
	"net/http"

	"github.com/FokusInternal/bifrost/config"
	rl "github.com/FokusInternal/bifrost/middlewares"
	"github.com/FokusInternal/bifrost/pkg/logging"
	routes "github.com/FokusInternal/bifrost/routes"
	v1 "github.com/FokusInternal/bifrost/routes/v1"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	logging.Setup()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(rl.LoggingMiddleware())
	r.Use(middleware.Recoverer)

	r.Get("/healthz", routes.Healthz)
	r.Get("/version", routes.Version)

	r.Route("/v1", func(r chi.Router) {
		r.Use(apiVersionCtx("v1"))
		r.Get("/hello", v1.SayHello)

		r.Post("/keys", routes.CreateKey)
		r.Delete("/keys/{id}", routes.DeleteKey)

		r.Post("/rootkeys", routes.CreateRootKey)
		r.Put("/rootkeys/{id}", routes.UpdateRootKey)
		r.Delete("/rootkeys/{id}", routes.DeleteRootKey)

		r.Post("/services", routes.CreateService)
		r.Delete("/services/{id}", routes.DeleteService)

		r.With(rl.RateLimitMiddleware()).Post("/rate", v1.SayHello)

		r.With(rl.RateLimitMiddleware()).Handle("/proxy/{rest:.*}", http.HandlerFunc(v1.Proxy))
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
