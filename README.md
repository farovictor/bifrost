# Bifrost – Secure, Delegated API Access for Cloud-Native Environments

Bifrost is a lightweight, extensible API proxy written in Go that enables secure delegation of API access through virtual keys. Instead of exposing long-lived secrets or API tokens to clients, Bifrost maps short-lived, scoped virtual keys to real credentials stored in root keys — and transparently proxies the request to the target upstream service.

## Requirements

Go 1.23.8. To install the toolchain and dependencies:

```bash
make setup
```

## Running locally

```bash
make run          # starts with in-memory stores, console log format
```

With a PostgreSQL backend:

```bash
export POSTGRES_DSN="postgres://bifrost:bifrost@localhost:5432/bifrost?sslmode=disable"
make run
```

With SQLite:

```bash
export BIFROST_DB=sqlite
export DATABASE_DSN=./bifrost.db
make run
```

## Running with Docker Compose

```bash
docker-compose up -d
```

The stack starts Bifrost, Redis, and PostgreSQL. A one-off `setup-job` seeds an `admin` user in `demo-org` and prints the auth token on exit. To retrieve it:

```bash
docker compose logs setup-job | tail -1
```

Tear down:

```bash
docker-compose down
```

## Running tests

```bash
go test ./...
```

Tests use in-memory stores and require no external services. Generate a coverage report:

```bash
make test   # runs tests and writes coverage.out
go tool cover -html=coverage.out
```

## Configuration

| Variable | Description | Default |
|---|---|---|
| `BIFROST_PORT` | HTTP port | `3333` |
| `BIFROST_DB` | Database backend (`sqlite` or `postgres`) | `postgres` |
| `POSTGRES_DSN` | PostgreSQL connection string | *(empty — uses in-memory)* |
| `DATABASE_DSN` | SQLite file path (when `BIFROST_DB=sqlite`) | *(empty)* |
| `REDIS_ADDR` | Redis address for rate limiting | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password | *(empty)* |
| `REDIS_DB` | Redis DB index | `0` |
| `REDIS_PROTOCOL` | Redis protocol version | `3` |
| `BIFROST_MODE` | Set to `test` to accept the static API key | *(empty)* |
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

### Authentication

Most `/v1` management endpoints require **both**:
- `X-API-Key: <api_key>` — the user's API key
- `Authorization: Bearer <token>` — the signed auth token returned by `user-add` or `POST /v1/users`

The token encodes `user_id`, `org_id`, and expiry (24 h). It is verified with HMAC-SHA256 using `BIFROST_SIGNING_KEY`.

Exceptions:
- `POST /v1/users`, `GET /v1/user`, `POST /v1/user/rootkeys` — bearer token only (no API key required)
- `GET /v1/proxy/*` — virtual key only (`X-Virtual-Key` header or `key` query param)
- `GET /healthz`, `GET /version` — no auth

In `test` mode or SQLite mode, any bearer token is accepted and `BIFROST_STATIC_API_KEY` is used instead of a user lookup.

---

## HTTP API

All error responses are JSON: `{"error": "message"}`.

### Health / Version

```
GET /healthz          → 200 "ok"
GET /version          → 200 {"version":"..."}
```

### Users

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/v1/users` | Bearer token | Create user, optionally with org |
| `GET` | `/v1/user` | Bearer token | Get authenticated user info + orgs |

**POST /v1/users** body:
```json
{
  "name": "Alice",
  "email": "alice@example.com",
  "org_id": "existing-org-id",
  "org_name": "New Org Name",
  "role": "member"
}
```
- Supply `org_id` to join an existing org, or `org_name` to create a new one.
- `role`: `owner`, `admin`, or `member` (default: `member`).
- Returns `201` with the user object and a signed `token` field.

### Organizations

All org endpoints require API key + bearer token.

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/orgs` | Create organization |
| `GET` | `/v1/orgs` | List all organizations |
| `GET` | `/v1/orgs/{id}` | Get organization by ID |
| `DELETE` | `/v1/orgs/{id}` | Delete organization |
| `POST` | `/v1/orgs/{id}/members` | Add a user to the org |
| `GET` | `/v1/orgs/{id}/members` | List members of the org |
| `DELETE` | `/v1/orgs/{id}/members/{userID}` | Remove a member |

**POST /v1/orgs** body:
```json
{"id": "my-org", "name": "My Org", "domain": "myorg.com", "email": "admin@myorg.com"}
```

**POST /v1/orgs/{id}/members** body:
```json
{"user_id": "usr_abc", "role": "member"}
```

### Root Keys

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/v1/user/rootkeys` | Bearer token | Create root key (token only) |
| `GET` | `/v1/rootkeys` | API key + token | List root keys |
| `POST` | `/v1/rootkeys` | API key + token | Create root key |
| `PUT` | `/v1/rootkeys/{id}` | API key + token | Update root key |
| `DELETE` | `/v1/rootkeys/{id}` | API key + token | Delete root key |

**Body**:
```json
{"id": "my-root", "api_key": "sk-..."}
```

### Services

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/v1/services` | API key + token | List services |
| `POST` | `/v1/services` | API key + token | Create service |
| `DELETE` | `/v1/services/{id}` | API key + token | Delete service |

**POST /v1/services** body:
```json
{
  "id": "my-svc",
  "endpoint": "https://api.openai.com",
  "root_key_id": "my-root",
  "credential_header": "Authorization"
}
```

`credential_header` controls how the root key is injected into upstream requests:
- `""` or `"X-API-Key"` (default) → `X-API-Key: <key>`
- `"Authorization"` → `Authorization: Bearer <key>`
- Any other value → `<header>: <key>`

### Virtual Keys

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/v1/keys` | API key + token | List virtual keys |
| `POST` | `/v1/keys` | API key + token | Create virtual key |
| `DELETE` | `/v1/keys/{id}` | API key + token | Revoke virtual key |

**POST /v1/keys** body:
```json
{
  "id": "vk-alice",
  "scope": "read",
  "target": "my-svc",
  "expires_at": "2026-12-31T23:59:59Z",
  "rate_limit": 60
}
```

- `scope`: `read` (GET/HEAD only) or `write` (all methods)
- `rate_limit`: maximum requests per minute
- `expires_at`: must be in the future

### Proxy

```
GET|POST|... /v1/proxy/{path}
  X-Virtual-Key: <key>     (or ?key=<key> query param)
```

Bifrost validates the key, enforces scope and rate limit, strips the virtual key from the forwarded request, injects the root key credential (per `credential_header`), and proxies to the upstream service endpoint.

### Metrics

When `BIFROST_ENABLE_METRICS=true`:
```
GET /metrics     → Prometheus text format
```

Available metrics: `request_total`, `request_duration_seconds`, `key_usage_total`.

---

## CLI

```bash
# Seed an admin user + org (run once against a fresh database)
go run ./cmd/bifrost init-admin

# Create a user (returns api_key and token)
go run ./cmd/bifrost user-add --name Alice --email alice@acme.com --org-name acme --role owner

# Register a root key
go run ./cmd/bifrost rootkey-add --id my-root --apikey sk-secret

# Update a root key
go run ./cmd/bifrost rootkey-update --id my-root --apikey sk-new-secret

# Register a service
go run ./cmd/bifrost service-add --id my-svc --endpoint https://api.example.com --rootkey my-root

# Issue a virtual key
go run ./cmd/bifrost issue --id vk-1 --target my-svc --scope read --ttl 10m --rate-limit 60

# Revoke a virtual key
go run ./cmd/bifrost revoke vk-1

# Delete a service
go run ./cmd/bifrost service-delete my-svc

# Delete a root key
go run ./cmd/bifrost rootkey-delete my-root

# Check server health
go run ./cmd/bifrost check
```

Use `--addr` to point to a non-default server address (default: `http://localhost:3333`).

---

## End-to-End Example

```bash
# 1. Start the server
make run

# 2. Create a root key
go run ./cmd/bifrost rootkey-add --id demo-root --apikey my-secret-key

# 3. Register an upstream service (default X-API-Key injection)
go run ./cmd/bifrost service-add --id demo-svc --endpoint http://localhost:8081 --rootkey demo-root

# 4. Create a user and org (save the printed token)
go run ./cmd/bifrost user-add --name Admin --email admin@example.com --org-name demo --role owner
# → prints: api_key=<key>  token=<token>

# 5. Issue a virtual key (requires API key + token)
curl -X POST http://localhost:3333/v1/keys \
  -H "X-API-Key: <api_key>" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"id":"demo-key","target":"demo-svc","scope":"read","expires_at":"2027-01-01T00:00:00Z","rate_limit":60}'

# 6. Make a proxied request using the virtual key
curl -H "X-Virtual-Key: demo-key" http://localhost:3333/v1/proxy/hello
```

---

## Architecture

```
Client
  │
  ▼
chi Router
  ├─ GET  /healthz, /version          (no auth)
  ├─ POST /v1/users                   (OrgCtxMiddleware)
  ├─ GET  /v1/user                    (OrgCtxMiddleware)
  ├─ POST /v1/user/rootkeys           (OrgCtxMiddleware)
  ├─ */v1/proxy/{rest}                (RateLimitMiddleware → Redis)
  │    └─ Proxy handler: validates virtual key, injects root key, forwards
  └─ /v1/* management group           (AuthMiddleware → OrgCtxMiddleware)
       ├─ /keys, /rootkeys, /services
       └─ /orgs, /orgs/{id}/members
```

Stores (`Server` struct):
- `UserStore`, `KeyStore`, `RootKeyStore`, `ServiceStore`, `OrgStore`, `MembershipStore`
- Each has an in-memory implementation (tests / no-DSN mode) and a SQL implementation (GORM).

---

## Roles

- `owner` — full control over the organization and its members
- `admin` — manage resources, cannot demote owners
- `member` — limited to explicitly granted resources

---

## Planned Extensions

- Open Policy Agent (OPA) integration for dynamic authorization
- Multiple target backend presets (OpenAI, Stripe, etc.)
- Web-based management dashboard
- JWT pass-through with verification hooks
- Vault secret backend
- Kubernetes operator with CRD support
