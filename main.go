package main

import (
	"context"
	"flag"
	"net/http"
	"os"

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
	dbFlag := flag.String(
		"db",
		config.DBType(),
		"database backend to use (sqlite or postgres). Flag takes precedence over BIFROST_DB",
	)
	modeFlag := flag.String(
		"mode",
		config.Mode(),
		"application mode. Flag takes precedence over BIFROST_MODE",
	)
	flag.Parse()
	os.Setenv("BIFROST_DB", *dbFlag)
	os.Setenv("BIFROST_MODE", *modeFlag)
	if config.Mode() == "test" {
		os.Setenv("BIFROST_DB", "sqlite")
	}

	dsn := config.PostgresDSN()
	dbType := config.DBType()

	var srv *routes.Server

	switch dbType {
	case "sqlite", "postgres":
		if dsn == "" {
			srv = &routes.Server{
				UserStore:       users.NewMemoryStore(),
				KeyStore:        keys.NewMemoryStore(),
				RootKeyStore:    rootkeys.NewMemoryStore(),
				ServiceStore:    services.NewMemoryStore(),
				OrgStore:        orgs.NewMemoryStore(),
				MembershipStore: orgs.NewMemoryMembershipStore(),
			}
			logging.Logger.Info().Msg("In-Memory Store set")
		} else {
			db, err := database.Connect(dbType, dsn)
			if err != nil {
				logging.Logger.Fatal().Err(err).Msg("connect " + dbType)
			}
			srv = &routes.Server{
				UserStore:       users.NewSQLStore(db),
				KeyStore:        keys.NewSQLStore(db),
				RootKeyStore:    rootkeys.NewSQLStore(db),
				ServiceStore:    services.NewSQLStore(db),
				OrgStore:        orgs.NewSQLStore(db),
				MembershipStore: orgs.NewSQLMembershipStore(db),
			}
			logging.Logger.Info().Str("db", dbType).Msg("Store set")
		}
	default:
		srv = &routes.Server{
			UserStore:       users.NewMemoryStore(),
			KeyStore:        keys.NewMemoryStore(),
			RootKeyStore:    rootkeys.NewMemoryStore(),
			ServiceStore:    services.NewMemoryStore(),
			OrgStore:        orgs.NewMemoryStore(),
			MembershipStore: orgs.NewMemoryMembershipStore(),
		}
		logging.Logger.Info().Msg("In-Memory Store set")
	}

	v1h := &v1.Handler{
		KeyStore:     srv.KeyStore,
		ServiceStore: srv.ServiceStore,
		RootKeyStore: srv.RootKeyStore,
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
		r.With(rl.OrgCtxMiddleware(srv.MembershipStore)).Post("/users", srv.CreateUser)
		r.With(rl.OrgCtxMiddleware(srv.MembershipStore)).Get("/user", srv.GetUserInfo)
		r.With(rl.OrgCtxMiddleware(srv.MembershipStore)).Post("/user/rootkeys", srv.CreateRootKey)

		// Proxy - authenticated by the virtual key; no API key or token required
		r.With(rl.RateLimitMiddleware(srv.KeyStore)).Handle("/proxy/{rest:.*}", http.HandlerFunc(v1h.Proxy))

		// Endpoints requiring API key and auth token
		r.Group(func(r chi.Router) {
			r.Use(rl.AuthMiddleware(srv.UserStore))
			r.Use(rl.OrgCtxMiddleware(srv.MembershipStore))

			r.Get("/hello", v1.SayHello)

			r.Get("/keys", srv.ListKeys)
			r.Post("/keys", srv.CreateKey)
			r.Delete("/keys/{id}", srv.DeleteKey)

			r.Get("/rootkeys", srv.ListRootKeys)
			r.Post("/rootkeys", srv.CreateRootKey)
			r.Put("/rootkeys/{id}", srv.UpdateRootKey)
			r.Delete("/rootkeys/{id}", srv.DeleteRootKey)

			r.Get("/services", srv.ListServices)
			r.Post("/services", srv.CreateService)
			r.Delete("/services/{id}", srv.DeleteService)

			r.With(rl.RateLimitMiddleware(srv.KeyStore)).Post("/rate", v1.SayHello)
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
