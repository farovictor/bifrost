package main

import (
	"context"
	"net/http"

	"github.com/farovictor/bifrost/config"
	rl "github.com/farovictor/bifrost/middlewares"
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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	logging.Setup()

	dsn := config.PostgresDSN()
	if dsn == "" {
		routes.UserStore = users.NewMemoryStore()
		routes.KeyStore = keys.NewMemoryStore()
		routes.RootKeyStore = rootkeys.NewMemoryStore()
		routes.ServiceStore = services.NewMemoryStore()
		routes.OrgStore = orgs.NewMemoryStore()
		routes.MembershipStore = orgs.NewMembershipStore()
	} else {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			logging.Logger.Fatal().Err(err).Msg("connect postgres")
		}
		sqlDB, err := db.DB()
		if err != nil {
			logging.Logger.Fatal().Err(err).Msg("db access")
		}
		if err := sqlDB.Ping(); err != nil {
			logging.Logger.Fatal().Err(err).Msg("ping postgres")
		}

		routes.UserStore = users.NewPostgresStore(db)
		routes.KeyStore = keys.NewPostgresStore(db)
		routes.RootKeyStore = rootkeys.NewPostgresStore(db)
		routes.ServiceStore = services.NewPostgresStore(db)
		routes.OrgStore = orgs.NewPostgresStore(db)
		routes.MembershipStore = orgs.NewMembershipStore()
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
		r.Use(rl.AuthMiddleware())
		r.Use(rl.OrgCtxMiddleware())
		r.Get("/hello", v1.SayHello)

		r.Post("/users", routes.CreateUser)

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

	if config.MetricsEnabled() {
		r.Handle("/metrics", promhttp.Handler())
	}

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
