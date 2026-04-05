package main

import (
	"context"
	"flag"
	"net/http"
	"os"

	_ "github.com/farovictor/bifrost/docs/swagger"
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
	"github.com/go-chi/cors"
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
	migrateOnly := flag.Bool(
		"migrate-only",
		false,
		"run database migrations and exit",
	)
	flag.Parse()
	os.Setenv("BIFROST_DB", *dbFlag)
	os.Setenv("BIFROST_MODE", *modeFlag)
	if config.Mode() == "test" {
		os.Setenv("BIFROST_DB", "sqlite")
	}

	if *migrateOnly {
		dsn := config.PostgresDSN()
		db, err := database.Connect(config.DBType(), dsn)
		if err != nil {
			logging.Logger.Fatal().Err(err).Msg("connect for migration")
		}
		if err := db.AutoMigrate(
			&users.User{},
			&keys.VirtualKey{},
			&rootkeys.RootKey{},
			&services.Service{},
			&orgs.Organization{},
			&orgs.Membership{},
		); err != nil {
			logging.Logger.Fatal().Err(err).Msg("auto migrate")
		}
		logging.Logger.Info().Msg("migrations complete")
		return
	}

	dsn := config.PostgresDSN()
	if config.Mode() == "test" {
		dsn = "" // test mode always uses in-memory stores
	}
	dbType := config.DBType()

	encKey := config.EncryptionKey()
	if encKey != nil {
		logging.Logger.Info().Msg("Root key encryption enabled")
	} else {
		logging.Logger.Warn().Msg("BIFROST_ENCRYPTION_KEY not set — root keys stored without encryption")
	}

	var srv *routes.Server

	switch dbType {
	case "sqlite", "postgres":
		if dsn == "" {
			srv = &routes.Server{
				UserStore:       users.NewMemoryStore(),
				KeyStore:        keys.NewMemoryStore(),
				RootKeyStore:    rootkeys.NewMemoryStoreWithKey(encKey),
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
				RootKeyStore:    rootkeys.NewSQLStoreWithKey(db, encKey),
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
			RootKeyStore:    rootkeys.NewMemoryStoreWithKey(encKey),
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
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   config.CORSAllowedOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(rl.LoggingMiddleware())
	r.Use(rl.MetricsMiddleware())
	r.Use(middleware.Recoverer)

	r.Get("/healthz", routes.Healthz)
	r.Get("/version", routes.Version)
	r.Get("/docs/openapi.json", routes.OpenAPISpec)
	r.Get("/docs/openapi.yaml", routes.OpenAPISpecYAML)
	r.With(rl.AuthMiddleware(srv.UserStore)).Post("/mcp", srv.MCP)

	r.Route("/v1", func(r chi.Router) {
		r.Use(apiVersionCtx("v1"))

		// One-shot bootstrap — no auth, only works when no users exist
		r.Post("/setup", srv.Setup)

		// Endpoints that only require the auth token
		r.With(rl.OrgCtxMiddleware(srv.MembershipStore)).Post("/users", srv.CreateUser)
		r.With(rl.OrgCtxMiddleware(srv.MembershipStore)).Get("/user", srv.GetUserInfo)
		r.With(rl.OrgCtxMiddleware(srv.MembershipStore)).Post("/user/rootkeys", srv.CreateRootKey)
		r.Post("/token/refresh", srv.RefreshToken)

		// Proxy - authenticated by the virtual key; no API key or token required
		r.With(rl.RateLimitMiddleware(srv.KeyStore)).Handle("/proxy/*", http.HandlerFunc(v1h.Proxy))

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
			r.Put("/services/{id}", srv.UpdateService)
			r.Delete("/services/{id}", srv.DeleteService)

			r.Get("/orgs", srv.ListOrgs)
			r.Post("/orgs", srv.CreateOrg)
			r.Get("/orgs/{id}", srv.GetOrg)
			r.Delete("/orgs/{id}", srv.DeleteOrg)
			r.Get("/orgs/{id}/members", srv.ListOrgMembers)
			r.Post("/orgs/{id}/members", srv.AddOrgMember)
			r.Delete("/orgs/{id}/members/{userID}", srv.RemoveOrgMember)

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
