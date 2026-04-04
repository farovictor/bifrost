# Bifrost — Backend Architecture

## 1. Introduction

Bifrost is a brownfield Go HTTP proxy that maps short-lived virtual keys to real API credentials. It acts as a credential management gateway between consumers (internal services, AI agents, CI pipelines) and upstream AI/API providers.

**Core value proposition:** Consumers never see real credentials. They receive scoped, rate-limited, budget-capped virtual keys. Real credentials are stored server-side and injected at request time.

**Current state:** Production-ready single-node deployment with in-memory or PostgreSQL backing, HMAC-signed auth tokens, CORS support, and an OpenAPI spec. This document covers both the existing system and planned extensions across Epic 1–5.

---

## 2. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Bifrost                                 │
│                                                                 │
│  ┌──────────────┐    ┌─────────────────┐    ┌───────────────┐  │
│  │  Management  │    │   Proxy Plane   │    │  MCP Server   │  │
│  │    Plane     │    │                 │    │  (Epic 2)     │  │
│  │  /v1/keys    │    │ /v1/proxy/{rest}│    │  /mcp         │  │
│  │  /v1/rootkeys│    │                 │    │  /mcp/sse     │  │
│  │  /v1/services│    │  RateLimit MW   │    └───────────────┘  │
│  │  /v1/orgs    │    │  KeyStore       │                        │
│  │  /v1/users   │    │  SecretBackend  │    ┌───────────────┐  │
│  │  /v1/setup   │    │  UsageTracker   │    │ Usage Tracker │  │
│  └──────┬───────┘    └────────┬────────┘    │  (Epic 3)     │  │
│         │                     │             └───────┬───────┘  │
│  ┌──────▼─────────────────────▼─────────────────────▼───────┐  │
│  │                      Store Layer                          │  │
│  │   UserStore  KeyStore  RootKeyStore  ServiceStore         │  │
│  │   OrgStore   MembershipStore  UsageStore (Epic 3)         │  │
│  └──────────────────────┬────────────────────────────────────┘  │
└─────────────────────────│───────────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │ In-Memory│    │PostgreSQL│    │  Vault   │
    │  (dev)   │    │  (prod)  │    │ (Epic 4) │
    └──────────┘    └──────────┘    └──────────┘
```

### Architectural patterns

- **Dependency injection** — stores and config injected into `routes.Server` and `v1.Handler` at startup; no globals
- **Middleware chain** — auth, org context, rate limiting composed via chi middleware
- **Dual-store** — every domain has `MemoryStore` (dev/test) and `SQLStore` (prod) behind a common interface
- **Stateless instances** — all state in PostgreSQL + Redis; any instance handles any request
- **Async side effects** — usage tracking via buffered channel; webhooks via dispatcher goroutine

---

## 3. Tech Stack

### Existing

| Layer | Technology | Notes |
|-------|-----------|-------|
| Language | Go 1.23 | `CGO_ENABLED=0`, static binary |
| HTTP router | `go-chi/chi v5` | Composable middleware |
| CORS | `go-chi/cors` | Configurable via `BIFROST_CORS_ORIGINS` |
| ORM | GORM | PostgreSQL + SQLite drivers |
| Auth | HMAC-SHA256 (`pkg/auth`) | Sign/verify bearer tokens |
| Logging | `rs/zerolog` | Structured JSON; console mode for dev |
| Metrics | `prometheus/client_golang` | Optional; `/metrics` endpoint |
| OpenAPI | `swaggo/swag` | Generated from handler annotations |
| CLI | `spf13/cobra` | `cmd/bifrost/` wraps HTTP API |

### Planned additions

| Epic | Technology | Purpose |
|------|-----------|---------|
| Epic 2 | JSON-RPC 2.0 + SSE | MCP server transport |
| Epic 3 | PostgreSQL `usage_events` table | Usage storage + retention |
| Epic 3 | `net/http/webhook` dispatcher | Webhook delivery |
| Epic 4 | HashiCorp Vault SDK | Pluggable secret backend |
| Epic 1 | AES-256-GCM (`crypto/aes`) | Encryption at rest for root keys |

---

## 4. Data Models

### Existing models

```go
// User — human or service identity
type User struct {
    ID     string
    Name   string
    Email  string
    APIKey string  // apk-... prefix
}

// Organization — tenant boundary
type Organization struct {
    ID   string
    Name string
}

// Membership — user ↔ org relationship
type Membership struct {
    UserID string
    OrgID  string
    Role   string  // owner | admin | member
}

// VirtualKey — short-lived proxy credential
type VirtualKey struct {
    ID          string
    Name        string
    RootKeyID   string
    OrgID       string
    Scope       string
    ExpiresAt   time.Time
    RateLimit   int
    BudgetUSD   float64
    // Epic 3 additions:
    SpentUSD    float64
    RequestCount int64
    LastUsedAt  *time.Time
    // Epic 2 additions:
    IssuedBy    string   // "mcp" | "api"
    MCPClientID string
    TTL         *time.Duration
}

// RootKey — real upstream credential (encrypted at rest — Story 1.3)
type RootKey struct {
    ID               string
    Name             string
    OrgID            string
    EncryptedValue   []byte  // AES-256-GCM ciphertext
    CredentialHeader string  // injection header name
    ServiceID        string
}

// Service — upstream API definition
type Service struct {
    ID               string
    Name             string
    BaseURL          string
    CredentialHeader string
}
```

### New models (Epic 3)

```go
// UsageEvent — immutable proxy request record
type UsageEvent struct {
    ID           string
    VirtualKeyID string
    ServiceID    string
    OrgID        string
    Timestamp    time.Time
    LatencyMS    int64
    StatusCode   int
    InputTokens  int
    OutputTokens int
    CostUSD      float64
}

// WebhookConfig — org-level webhook subscription
type WebhookConfig struct {
    ID        string
    OrgID     string
    URL       string
    Events    []string  // ["usage.threshold", "key.expired"]
    Secret    string    // HMAC signing secret
    Active    bool
}

// WebhookDelivery — delivery attempt log
type WebhookDelivery struct {
    ID         string
    WebhookID  string
    EventType  string
    Payload    []byte
    StatusCode int
    Attempts   int
    DeliveredAt *time.Time
    Error      string
}
```

---

## 5. Components

### Management Plane (`routes/`)

HTTP handlers as methods on `*routes.Server`. Each resource has its own file (`keys.go`, `rootkeys.go`, `services.go`, `users.go`, `orgs.go`, `setup.go`). `server.go` holds the `Server` struct and the shared `writeError()` helper.

Authentication: `X-API-Key` header validated by `AuthMiddleware`; org context extracted from bearer token by `OrgCtxMiddleware`.

### Proxy Plane (`routes/v1/proxy.go`)

- Reads virtual key from `Authorization: Bearer vk-...` or `?key=vk-...`
- Validates scope, expiry, budget
- Calls `SecretBackend.GetSecret(rootKeyID)` to retrieve plaintext credential
- Injects credential via `injectCredential()` (header name from `Service.CredentialHeader`)
- Reverse-proxies request to upstream
- Emits `UsageEvent` to `UsageTracker` (non-blocking)

### MCP Server (Epic 2 — `routes/mcp/`)

JSON-RPC 2.0 endpoint implementing the Model Context Protocol. Exposes tools: `get_key`, `list_services`, `create_key`, `revoke_key`, `list_rootkeys`. SSE transport at `/mcp/sse` for streaming responses.

### Middleware Stack (`middlewares/`)

| Middleware | Constructor | Purpose |
|-----------|------------|---------|
| `AuthMiddleware` | `AuthMiddleware(UserStore)` | Validates `X-API-Key` |
| `OrgCtxMiddleware` | `OrgCtxMiddleware(MembershipStore)` | Extracts org from bearer token |
| `RateLimitMiddleware` | `RateLimitMiddleware(KeyStore)` | Redis or local counter per key |
| `LoggingMiddleware` | — | Structured request/response logging |
| `MetricsMiddleware` | — | Prometheus counter + histogram |

### Store Layer (`pkg/*/`)

Every domain package exposes a `Store` interface following the pattern:

```
Create, Get, Delete, [Update], List, [ListBy*]
```

Each has a `MemoryStore` (dev/test) and `SQLStore` (GORM, PostgreSQL/SQLite). New methods must be implemented on both.

### Secret Backend (`pkg/secrets/` — Epic 4)

```go
type SecretBackend interface {
    GetSecret(rootKeyID string) (string, error)
    PutSecret(rootKeyID, value string) error
    DeleteSecret(rootKeyID string) error
}
```

Implementations: `LocalBackend` (default — AES-256-GCM via RootKeyStore), `VaultBackend` (Epic 4).

### Usage Tracker (`pkg/usage/` — Epic 3)

Buffered channel (`chan UsageEvent`, capacity 1000) drained by a background worker that batch-inserts into PostgreSQL. Non-blocking emit on the hot proxy path. On buffer full: drops oldest event and increments a Prometheus counter.

### Webhook Dispatcher (`pkg/webhooks/` — Epic 3)

Subscribes to usage events, evaluates trigger conditions (budget threshold, key expiry), delivers HTTP POST with HMAC-signed payload. Retries up to 3 times with exponential backoff. Logs delivery attempts to `WebhookDelivery`.

### Retention Job (Epic 3)

Runs as a Kubernetes `CronJob` (daily at 02:00 UTC) or `--retention-job` CLI flag. Deletes `UsageEvent` rows older than `BIFROST_RETENTION_DAYS` (default 90).

---

## 6. Core Workflows

### 6.1 Proxy Request Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant RL as RateLimitMiddleware
    participant P as Proxy Handler
    participant KS as KeyStore
    participant SB as SecretBackend
    participant UT as UsageTracker
    participant U as Upstream API

    C->>RL: GET /v1/proxy/path?key=vk-xxx
    RL->>KS: Get(vk-xxx)
    KS-->>RL: VirtualKey{scope, expiry, rootKeyID, budget}
    RL->>RL: check rate limit (Redis or local)
    RL-->>C: 429 Too Many Requests (if exceeded)
    RL->>P: pass (key in context)
    P->>P: validate expiry, scope match
    P->>SB: GetSecret(rootKeyID)
    SB-->>P: plaintext credential
    P->>P: injectCredential(req, credential)
    P->>U: proxied request
    U-->>P: response
    P->>UT: emit UsageEvent (async, buffered)
    P-->>C: upstream response
```

### 6.2 Bootstrap Flow

```mermaid
sequenceDiagram
    participant C as Client (Heimdall)
    participant S as Setup Handler
    participant US as UserStore
    participant OS as OrgStore
    participant MS as MembershipStore

    C->>S: POST /v1/setup {name, email, org_name}
    S->>US: Count()
    US-->>S: 0
    S->>US: Create(User)
    S->>OS: Create(Org)
    S->>MS: Create(Membership{role: owner})
    S->>S: buildAuthToken(userID, orgID)
    S-->>C: 201 {id, name, email, api_key, token}

    Note over S: Returns 409 if Count() > 0
```

### 6.3 Virtual Key Creation Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant AM as AuthMiddleware
    participant OM as OrgCtxMiddleware
    participant H as Keys Handler
    participant KS as KeyStore
    participant RKS as RootKeyStore

    C->>AM: POST /v1/keys {X-API-Key, Authorization}
    AM->>AM: validate X-API-Key via UserStore
    AM->>AM: verify Bearer token (HMAC)
    AM->>OM: pass (userID in context)
    OM->>OM: extract orgID from token
    OM->>H: pass (orgID in context)
    H->>RKS: Get(rootKeyID) — validate rootKey exists
    H->>KS: Create(VirtualKey{scope, expiry, budget, orgID})
    KS-->>H: saved
    H-->>C: 201 {id, key_id, ...}
```

### 6.4 MCP Tool Call Flow (Epic 2)

```mermaid
sequenceDiagram
    participant A as AI Agent
    participant MCP as MCP Server
    participant KS as KeyStore

    A->>MCP: POST /mcp {jsonrpc, method: "tools/call", params: {name: "get_key"}}
    MCP->>MCP: validate JSON-RPC 2.0 envelope
    MCP->>KS: Create(VirtualKey{scope, ttl, rootKeyID})
    KS-->>MCP: VirtualKey
    MCP-->>A: {result: {key: "vk-xxx", expires_at: "..."}}

    Note over A,MCP: Agent uses vk-xxx directly against /v1/proxy/*
```

### 6.5 Async Usage Tracking Flow (Epic 3)

```mermaid
sequenceDiagram
    participant P as Proxy Handler
    participant UT as UsageTracker
    participant W as Worker Goroutine
    participant DB as PostgreSQL

    P->>UT: emit(UsageEvent) — non-blocking send to buffered channel
    Note over P: returns immediately, no latency impact

    loop batch flush (every N events or T interval)
        W->>W: drain channel into batch[]
        W->>DB: INSERT INTO usage_events (batch)
    end
```

### 6.6 Secret Backend Retrieval Flow (Epic 4)

```mermaid
sequenceDiagram
    participant P as Proxy Handler
    participant SB as SecretBackend interface
    participant LS as LocalStore (default)
    participant V as Vault (Epic 4)

    P->>SB: GetSecret(rootKeyID)
    alt LocalStore backend
        SB->>LS: RootKeyStore.Get(rootKeyID)
        LS->>LS: AES-256-GCM decrypt (Story 1.3)
        LS-->>SB: plaintext credential
    else Vault backend
        SB->>V: GET /v1/secret/data/{rootKeyID}
        V-->>SB: plaintext credential
    end
    SB-->>P: plaintext credential
```

---

## 7. API Design

### Authentication

| Header | Value | Purpose |
|--------|-------|---------|
| `X-API-Key` | `apk-...` | Identifies the user |
| `Authorization` | `Bearer <token>` | HMAC-signed token carrying `user_id` + `org_id` |

Proxy requests authenticate via virtual key only (`?key=vk-...` or `Authorization: Bearer vk-...`).

### Error Response Shape

```json
{ "error": "human-readable message" }
```

`Content-Type: application/json` is always set, even on errors.

### Endpoints

#### Setup (no auth)

| Method | Path | Success | Errors |
|--------|------|---------|--------|
| `POST` | `/v1/setup` | 201 `SetupResponse` | 400, 409, 500 |

#### Users

| Method | Path | Auth | Success | Errors |
|--------|------|------|---------|--------|
| `POST` | `/v1/users` | Token | 201 `CreateUserResponse` | 400, 404, 409, 500 |
| `GET` | `/v1/user` | Token | 200 user + orgs | 401, 404, 500 |
| `POST` | `/v1/token/refresh` | Bearer | 200 `{token}` | 401, 500 |

#### Virtual Keys

| Method | Path | Auth | Success | Errors |
|--------|------|------|---------|--------|
| `GET` | `/v1/keys` | API Key + Token | 200 `[]VirtualKey` | 401, 500 |
| `POST` | `/v1/keys` | API Key + Token | 201 `VirtualKey` | 400, 401, 500 |
| `DELETE` | `/v1/keys/{id}` | API Key + Token | 204 | 401, 404, 500 |

#### Root Keys

| Method | Path | Auth | Success | Errors |
|--------|------|------|---------|--------|
| `GET` | `/v1/rootkeys` | API Key + Token | 200 `[]RootKey` | 401, 500 |
| `POST` | `/v1/rootkeys` | API Key + Token | 201 `RootKey` | 400, 401, 500 |
| `PUT` | `/v1/rootkeys/{id}` | API Key + Token | 200 `RootKey` | 400, 401, 404, 500 |
| `DELETE` | `/v1/rootkeys/{id}` | API Key + Token | 204 | 401, 404, 500 |
| `POST` | `/v1/user/rootkeys` | Token | 201 `RootKey` | 400, 401, 500 |

#### Services

| Method | Path | Auth | Success | Errors |
|--------|------|------|---------|--------|
| `GET` | `/v1/services` | API Key + Token | 200 `[]Service` | 401, 500 |
| `POST` | `/v1/services` | API Key + Token | 201 `Service` | 400, 401, 500 |
| `PUT` | `/v1/services/{id}` | API Key + Token | 200 `Service` | 400, 401, 404, 500 |
| `DELETE` | `/v1/services/{id}` | API Key + Token | 204 | 401, 404, 500 |

#### Organizations

| Method | Path | Auth | Success | Errors |
|--------|------|------|---------|--------|
| `GET` | `/v1/orgs` | API Key + Token | 200 `[]Org` | 401, 500 |
| `POST` | `/v1/orgs` | API Key + Token | 201 `Org` | 400, 401, 500 |
| `GET` | `/v1/orgs/{id}` | API Key + Token | 200 `Org` | 401, 404, 500 |
| `DELETE` | `/v1/orgs/{id}` | API Key + Token | 204 | 401, 404, 500 |
| `GET` | `/v1/orgs/{id}/members` | API Key + Token | 200 `[]Membership` | 401, 404, 500 |
| `POST` | `/v1/orgs/{id}/members` | API Key + Token | 201 | 400, 401, 404, 500 |
| `DELETE` | `/v1/orgs/{id}/members/{userID}` | API Key + Token | 204 | 401, 404, 500 |

#### Proxy

| Method | Path | Auth |
|--------|------|------|
| `*` | `/v1/proxy/{rest}` | Virtual key |

#### Utility

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/healthz` | `{"status":"ok"}` |
| `GET` | `/version` | `{"version":"..."}` |
| `GET` | `/docs/openapi.json` | OpenAPI spec (JSON) |
| `GET` | `/docs/openapi.yaml` | OpenAPI spec (YAML) |
| `GET` | `/metrics` | Prometheus metrics (if `BIFROST_METRICS=true`) |

#### MCP Server (Epic 2 — planned)

| Method | Path | Auth |
|--------|------|------|
| `POST` | `/mcp` | API Key |
| `GET` | `/mcp/sse` | API Key |

Tools: `get_key`, `list_services`, `create_key`, `revoke_key`, `list_rootkeys`

---

## 8. Infrastructure & Deployment

### 8.1 Deployment Topologies

#### Single-node (Docker Compose)

```
┌─────────────────────────────────┐
│  Docker Compose                 │
│                                 │
│  ┌───────────┐  ┌─────────────┐ │
│  │  bifrost  │  │  postgres   │ │
│  │  :8080    │──│  :5432      │ │
│  └───────────┘  └─────────────┘ │
│        │                        │
│  ┌───────────┐                  │
│  │   redis   │                  │
│  │  :6379    │                  │
│  └───────────┘                  │
└─────────────────────────────────┘
```

#### Active-passive HA (production)

```
              ┌──────────────┐
              │ Load Balancer │
              └──────┬───────┘
                     │
          ┌──────────┴──────────┐
          ▼                     ▼
   ┌─────────────┐       ┌─────────────┐
   │  bifrost-1  │       │  bifrost-2  │
   └──────┬──────┘       └──────┬──────┘
          │                     │
          └──────────┬──────────┘
                     │
        ┌────────────┼────────────┐
        ▼            ▼            ▼
  ┌──────────┐ ┌──────────┐ ┌──────────┐
  │ postgres │ │  redis   │ │  vault   │
  │ (primary)│ │ cluster  │ │ (Epic 4) │
  └──────────┘ └──────────┘ └──────────┘
```

Bifrost instances are fully stateless — all state lives in PostgreSQL and Redis.

### 8.2 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BIFROST_PORT` | `:8080` | Listen address |
| `BIFROST_MODE` | `production` | `production` \| `test` |
| `BIFROST_DB` | `sqlite` | `sqlite` \| `postgres` |
| `POSTGRES_DSN` | — | PostgreSQL connection string |
| `BIFROST_SIGNING_KEY` | _(required)_ | HMAC key for token signing |
| `BIFROST_CORS_ORIGINS` | `*` | Comma-separated allowed origins |
| `REDIS_ADDR` | — | Redis address for rate limiting |
| `BIFROST_LOG_FORMAT` | `json` | `json` \| `console` |
| `BIFROST_METRICS` | `false` | Enable Prometheus `/metrics` |
| `BIFROST_ENCRYPTION_KEY` | _(required in prod)_ | AES-256-GCM key for root key encryption (Story 1.3) |
| `VAULT_ADDR` | — | Vault server address (Epic 4) |
| `VAULT_TOKEN` | — | Vault auth token (Epic 4) |
| `BIFROST_RETENTION_DAYS` | `90` | Usage event retention in days (NFR12) |
| `BIFROST_TOKEN_TTL` | `24h` | Auth token lifetime (NFR13) |

### 8.3 Docker Compose

```yaml
services:
  bifrost:
    build: .
    ports: ["8080:8080"]
    environment:
      BIFROST_DB: postgres
      POSTGRES_DSN: postgres://bifrost:bifrost@postgres:5432/bifrost?sslmode=disable
      REDIS_ADDR: redis:6379
      BIFROST_SIGNING_KEY: ${BIFROST_SIGNING_KEY}
    depends_on: [postgres, redis]

  setup-job:
    image: curlimages/curl
    depends_on: [bifrost]
    command: >
      curl -sf -X POST http://bifrost:8080/v1/setup
      -H "Content-Type: application/json"
      -d '{"name":"Admin","email":"${ADMIN_EMAIL}","org_name":"Default"}'
    restart: "no"

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: bifrost
      POSTGRES_USER: bifrost
      POSTGRES_PASSWORD: bifrost
    volumes: [pgdata:/var/lib/postgresql/data]

  redis:
    image: redis:7-alpine
    volumes: [redisdata:/data]

volumes:
  pgdata:
  redisdata:
```

### 8.4 Container Image

```dockerfile
FROM golang:1.23-alpine AS builder
ARG VERSION=dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o bifrost-server main.go

FROM scratch
COPY --from=builder /app/bifrost-server /bifrost-server
EXPOSE 8080
ENTRYPOINT ["/bifrost-server"]
```

Target image size: < 20 MB (scratch base + statically linked binary).

### 8.5 Health & Observability

| Signal | Endpoint / Sink | Notes |
|--------|----------------|-------|
| Liveness | `GET /healthz` | Load balancer probe |
| Metrics | `GET /metrics` | Prometheus; opt-in via `BIFROST_METRICS=true` |
| Logs | stdout (JSON) | Forwarded by container runtime |
| Traces | — | Planned Phase 3 (OpenTelemetry) |

Prometheus metrics: `bifrost_requests_total`, `bifrost_request_duration_seconds`, `bifrost_proxy_requests_total`.

### 8.6 Kubernetes

Full manifests: `Deployment`, `Service`, `Ingress`, `HorizontalPodAutoscaler`, `Job/bifrost-migrate` (pre-upgrade hook), `CronJob/bifrost-retention` (Epic 3).

Key settings:
- `strategy.rollingUpdate.maxUnavailable: 0` — zero-downtime rollouts
- `minReplicas: 2` — always HA
- HPA target: 70% CPU, max 10 replicas
- Pre-upgrade migration job runs `--migrate-only` before new pods start

### 8.7 Helm Chart

Chart lives at `charts/bifrost/`. Key `values.yaml` toggles:

```yaml
externalSecrets:
  enabled: false          # flip to true in production
  secretStoreName: aws-secretsmanager

serviceAccount:
  annotations:
    eks.amazonaws.com/role-arn: ""   # IRSA
    # iam.gke.io/gcp-service-account: ""  # Workload Identity

migration:
  enabled: true

retention:
  enabled: false          # enable when Epic 3 ships
  schedule: "0 2 * * *"
```

### 8.8 CI/CD (GitHub Actions)

| Workflow | Trigger | Jobs |
|----------|---------|------|
| `ci.yml` | push / PR | `test`, `build`, `swagger-check`, `helm-lint` |
| `release.yml` | push to `main` or `v*` tag | `release` (build + push to `ghcr.io`) |

Image tags: `main`, `sha-<short>`, semver on tag push.

The `swagger-check` job regenerates the spec and fails if committed output differs — enforces `make swagger` discipline.

### 8.9 Secrets Management

| Environment | Method | Backend |
|-------------|--------|---------|
| Local / dev | env vars / `.env` | In-memory |
| Docker Compose | `.env` file (gitignored) | Postgres |
| K8s + AWS | IRSA → ESO → K8s Secret | Secrets Manager |
| K8s + GCP | Workload Identity → ESO → K8s Secret | Secret Manager |
| K8s + Vault | K8s auth → ESO → K8s Secret | Vault (Epic 4) |

No static cloud credentials in the cluster. Pods authenticate via projected OIDC tokens (IRSA / Workload Identity).

---

## Appendix: Proxy Credential Injection

`Service.CredentialHeader` controls how the root key is forwarded upstream:

| Value | Injected header |
|-------|----------------|
| `""` or `"X-API-Key"` | `X-API-Key: <key>` |
| `"Authorization"` | `Authorization: Bearer <key>` |
| any other value | `<value>: <key>` |

## Appendix: Roles

| Role | Capabilities |
|------|-------------|
| `owner` | Full control over org and members |
| `admin` | Manage resources; cannot demote owners |
| `member` | Read access to explicitly granted resources |
