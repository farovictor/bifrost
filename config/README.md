# Configuration Helpers

This package exposes helper functions that read configuration from
environment variables. Key variables include:

- `BIFROST_PORT` – HTTP port (default `3333`)
- `REDIS_ADDR` – Redis address (default `localhost:6379`)
- `REDIS_PASSWORD` – optional Redis password
- `REDIS_DB` – Redis DB index (default `0`)
- `REDIS_PROTOCOL` – Redis protocol version (default `3`)
- `POSTGRES_DSN` – Postgres connection string
- `BIFROST_ENABLE_METRICS` – enable Prometheus metrics when set
- `BIFROST_ADMIN_API_KEY` – API key for the admin, random when unset
- `BIFROST_ADMIN_NAME` – name for the admin user, defaults to `Admin`
- `BIFROST_ADMIN_EMAIL` – email for the admin user, defaults to `admin@example.com`

See the project `README.md` for more details and examples.
