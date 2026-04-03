# Configuration

All configuration is via environment variables.

| Variable | Description | Default |
|---|---|---|
| `BIFROST_PORT` | HTTP port | `3333` |
| `BIFROST_DB` | Database backend (`sqlite` or `postgres`) | `postgres` |
| `DATABASE_DSN` | Database connection string (PostgreSQL DSN or SQLite file path) | *(empty — uses in-memory)* |
| `REDIS_ADDR` | Redis address for rate limiting | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password | *(empty)* |
| `REDIS_DB` | Redis DB index | `0` |
| `REDIS_PROTOCOL` | Redis protocol version | `3` |
| `BIFROST_MODE` | Set to `test` to accept any bearer token | *(empty)* |
| `BIFROST_STATIC_API_KEY` | API key accepted in test/sqlite mode | `secret` |
| `BIFROST_SIGNING_KEY` | Base64 HMAC key for auth tokens | random each start |
| `BIFROST_LOG_LEVEL` | `debug`, `info`, `warn`, `error` | `info` |
| `BIFROST_LOG_FORMAT` | `json` or `console` | `json` |
| `BIFROST_ENABLE_METRICS` | Expose Prometheus `/metrics` | `false` |
| `BIFROST_ADMIN_API_KEY` | Seeded admin API key | random |
| `BIFROST_ADMIN_NAME` | Seeded admin user name | `Admin` |
| `BIFROST_ADMIN_EMAIL` | Seeded admin email | `admin@example.com` |
| `BIFROST_ADMIN_ORG_NAME` | Seeded admin org name | `Admin` |
| `BIFROST_ADMIN_ORG_EMAIL` | Seeded admin org contact email | `admin@example.com` |
| `BIFROST_ADMIN_ORG_DOMAIN` | Seeded admin org domain | `example.com` |
| `BIFROST_ADMIN_ROLE` | Seeded admin role | `owner` |

## Running locally

```bash
make run          # starts with in-memory stores, console log format
```

With PostgreSQL:
```bash
export DATABASE_DSN="postgres://bifrost:bifrost@localhost:5432/bifrost?sslmode=disable"
make run
```

With SQLite:
```bash
export BIFROST_DB=sqlite
export DATABASE_DSN=./bifrost.db
make run
```

## Docker Compose

```bash
docker-compose up -d
```

The stack starts Bifrost, Redis, and PostgreSQL. A one-off `setup-job` seeds an `admin` user in `demo-org` and prints the auth token on exit:

```bash
docker compose logs setup-job | tail -1
```

Tear down:
```bash
docker-compose down
```
