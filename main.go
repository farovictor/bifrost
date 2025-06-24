package main

import (
	"context"
	"net/http"

	"github.com/farovictor/bifrost/config"
	rl "github.com/farovictor/bifrost/middlewares"
	"github.com/farovictor/bifrost/pkg/database"
	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/logging"
	"github.com/farovictor/bifrost/pkg/metrics"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/users"
	routes "github.com/farovictor/bifrost/routes"
	v1 "github.com/farovictor/bifrost/routes/v1"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	logging.Setup()
}

func main() {

	dsn := config.PostgresDSN()
	if dsn == "" {
		routes.UserStore = users.NewMemoryStore()
		routes.KeyStore = keys.NewMemoryStore()
		routes.RootKeyStore = rootkeys.NewMemoryStore()
		routes.ServiceStore = services.NewMemoryStore()
		routes.OrgStore = orgs.NewMemoryStore()
		routes.MembershipStore = orgs.NewMemoryMembershipStore()
		logging.Logger.Info().Msg("In-Memory Store set")
	} else {
		db, err := database.Connect(dsn)
		if err != nil {
			logging.Logger.Fatal().Err(err).Msg("connect postgres")
		}

		routes.UserStore = users.NewPostgresStore(db)
		routes.KeyStore = keys.NewPostgresStore(db)
		routes.RootKeyStore = rootkeys.NewPostgresStore(db)
		routes.ServiceStore = services.NewPostgresStore(db)
		routes.OrgStore = orgs.NewPostgresStore(db)
		routes.MembershipStore = orgs.NewPostgresMembershipStore(db)
		logging.Logger.Info().Msg("Postgres Store set")
	}

	if config.MetricsEnabled() {
		metrics.Register(prometheus.DefaultRegisterer)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(rl.LoggingMiddleware())
	r.Use(rl.MetricsMiddleware())
	r.Use(middleware.Recoverer)

	r.Get("/healthz", routes.Healthz)
	r.Get("/version", routes.Version)

	r.Route("/v1", func(r chi.Router) {
		r.Use(apiVersionCtx("v1"))

		// Endpoints that only require the auth token
		r.With(rl.OrgCtxMiddleware()).Post("/users", routes.CreateUser)
		r.With(rl.OrgCtxMiddleware()).Get("/user", routes.GetUserInfo)

		// Endpoints requiring API key and auth token
		r.Group(func(r chi.Router) {
			r.Use(rl.AuthMiddleware())
			r.Use(rl.OrgCtxMiddleware())

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
	})

	if config.MetricsEnabled() {
		r.Handle("/metrics", promhttp.Handler())
	}

	logging.Logger.Info().Msg("Initializing Server ...")

	if err := http.ListenAndServe(config.ServerPort(), r); err != nil {
		logging.Logger.Fatal().Err(err).Msg("listen and serve")
	}
}

func apiVersionCtx(version string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), "api.version", version))
			next.ServeHTTP(w, r)
		})
	}
}
