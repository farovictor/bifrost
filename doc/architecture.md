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

## 9. Security

### 9.1 Threat Model Summary

| Threat | Mitigation |
|--------|-----------|
| Credential leakage via API response | Root keys never returned in plaintext after creation; encrypted at rest (Story 1.3) |
| Token forgery | HMAC-SHA256 signing with `BIFROST_SIGNING_KEY`; verify on every request |
| Virtual key abuse | Scope enforcement, expiry, rate limiting, budget cap |
| Unauthorized management access | Dual-credential requirement: `X-API-Key` + Bearer token |
| Bootstrap takeover | `POST /v1/setup` blocked after first user exists (`Count() > 0`) |
| CORS exploitation | Explicit allowlist via `BIFROST_CORS_ORIGINS`; defaults to `*` in dev, must be locked in prod |
| Replay attacks | Token carries `ExpiresAt`; short-lived (configurable TTL, default 24h) |
| Enumeration | All stores return opaque UUIDs; no sequential IDs |
| Secret backend compromise | Vault with lease rotation (Epic 4); local backend uses AES-256-GCM |
| Excessive proxy usage | Per-key rate limit + budget cap; async tracker flags overruns |

### 9.2 Authentication Layers

```
Layer 1 — Identity (who are you?)
  X-API-Key: apk-...
  → AuthMiddleware validates against UserStore

Layer 2 — Context (which org, which role?)
  Authorization: Bearer <hmac-token>
  → OrgCtxMiddleware decodes {userID, orgID, expiresAt}
  → looks up Membership{role} in MembershipStore

Layer 3 — Proxy (is this virtual key valid?)
  ?key=vk-... or Authorization: Bearer vk-...
  → RateLimitMiddleware checks scope, expiry, budget, rate
```

Management endpoints require **Layer 1 + Layer 2**.
Proxy endpoints require **Layer 3 only**.
`/v1/setup`, `/healthz`, `/version`, `/docs/*` require **no auth**.

### 9.3 Token Security

```go
type AuthToken struct {
    UserID    string
    OrgID     string
    ExpiresAt time.Time
}
```

- Signed with HMAC-SHA256 using `BIFROST_SIGNING_KEY`
- Verified on every authenticated request — no session store, no DB lookup
- TTL configurable via `BIFROST_TOKEN_TTL` (NFR13); default 24h
- `POST /v1/token/refresh` issues a fresh token without re-authenticating credentials
- **Emergency revocation:** rotating `BIFROST_SIGNING_KEY` immediately invalidates all existing tokens

### 9.4 Encryption at Rest (Story 1.3)

Root keys are encrypted before storage:

```
plaintext credential
    │
    ▼  AES-256-GCM
    │  key: BIFROST_ENCRYPTION_KEY (32 bytes)
    │  nonce: 12 random bytes prepended to ciphertext
    ▼
EncryptedValue []byte  →  stored in DB
```

Decryption happens only inside the proxy handler, in memory, never logged. HTTP responses for `RootKey` never include `EncryptedValue` — only `ID`, `Name`, `ServiceID`, and `CredentialHeader` are returned.

### 9.5 Secret Backend (Epic 4)

```
┌─────────────────────────────────────────────┐
│  SecretBackend interface                    │
│  GetSecret(id) / PutSecret(id, val)         │
│  DeleteSecret(id)                           │
└────────────────┬────────────────────────────┘
                 │
     ┌───────────┴───────────┐
     ▼                       ▼
LocalBackend            VaultBackend
(AES-256-GCM            (HashiCorp Vault
 via RootKeyStore)       KV v2, leases)
```

Vault advantages: dynamic secrets with automatic rotation, fine-grained policies per path, full audit log, lease-based access.

### 9.6 Virtual Key Scoping

Every virtual key carries a `Scope` string validated against the target service name at proxy time. A key scoped to `"openai"` cannot be used against a `"stripe"` endpoint — returns 403 on mismatch. Wildcard `"*"` allows any service.

### 9.7 Rate Limiting & Budget Enforcement

```
Per-request (RateLimitMiddleware):
  1. Increment Redis counter  key: "rl:{keyID}:{window}"
  2. If counter > VirtualKey.RateLimit → 429

Per-request (Proxy handler):
  3. If SpentUSD >= BudgetUSD → 402 Payment Required
  4. After response: update SpentUSD async via UsageTracker (Epic 3)
```

Redis is used when `REDIS_ADDR` is set; falls back to a local in-process counter (not suitable for multi-instance deployments).

### 9.8 MCP Security (Epic 2)

The `/mcp` endpoint authenticates via `X-API-Key` only. MCP-issued virtual keys are tagged `IssuedBy: "mcp"` and optionally carry `MCPClientID` for per-agent attribution. All org boundaries enforced identically to the REST API.

### 9.9 Production Security Checklist

- [ ] `BIFROST_SIGNING_KEY` — minimum 32 random bytes, stored in Secrets Manager / Vault
- [ ] `BIFROST_ENCRYPTION_KEY` — 32 random bytes, separate from signing key
- [ ] `BIFROST_CORS_ORIGINS` — set to Heimdall's exact origin, not `*`
- [ ] `BIFROST_TOKEN_TTL` — consider reducing to `1h` for high-security environments
- [ ] TLS terminated at load balancer / ingress — Bifrost runs plain HTTP internally
- [ ] Redis protected with `requirepass` or VPC-only access
- [ ] PostgreSQL — dedicated `bifrost` role with minimal privileges; SSL enabled
- [ ] Verify `POST /v1/setup` returns 409 after bootstrap
- [ ] Confirm logs do not capture `Authorization` or `X-API-Key` header values
- [ ] Secrets rotation plan documented — key rotation invalidates all tokens

---

## 10. Decisions & Trade-offs (ADRs)

### ADR-001: Dual-credential authentication (X-API-Key + Bearer token)

**Status:** Accepted

**Context:** Management endpoints need to identify both the user (for audit) and the org context (for resource scoping).

**Decision:** Require two credentials — `X-API-Key` for user identity and a short-lived HMAC Bearer token for org context carrying `{userID, orgID, expiresAt}`.

**Consequences:**
- ✅ Org context derived from token without a DB lookup on every request
- ✅ Revoking an API key immediately blocks access even with a valid token
- ✅ Tokens can be short-lived without frequent re-authentication (`/token/refresh`)
- ❌ Clients must manage two credentials
- ❌ No per-request org switching without a new token

---

### ADR-002: HMAC-SHA256 tokens over JWT

**Status:** Accepted

**Decision:** Keep custom HMAC tokens. Do not adopt JWT.

**Rationale:** No external dependencies; token format is internal; changing signing key immediately revokes all tokens. JWT adds complexity (header, claims, library surface) without benefit at this scale.

**Consequences:**
- ✅ Zero external library risk; minimal attack surface
- ❌ Cannot be verified by third-party services without a custom plugin
- ❌ No standard claims — migration required if federation is ever needed

---

### ADR-003: In-memory store as first-class runtime mode

**Status:** Accepted

**Decision:** Maintain `MemoryStore` permanently (not just for tests). Server starts in-memory when no `POSTGRES_DSN` is set.

**Consequences:**
- ✅ Zero-config local development; fast hermetic tests
- ❌ State lost on restart
- ❌ Dual implementation maintenance cost on every new store method

---

### ADR-004: Async usage tracking via buffered channel

**Status:** Accepted (Epic 3)

**Decision:** Proxy emits `UsageEvent` to a buffered channel (`cap=1000`). Background worker batch-inserts into PostgreSQL.

**Rationale:** Best-effort durability is acceptable for usage metrics. Buffered channel is idiomatic Go, zero dependencies, zero proxy latency impact.

**Consequences:**
- ✅ Zero proxy latency impact; no new infrastructure
- ❌ Events dropped on crash or buffer overflow (Prometheus counter tracks drops)
- ❌ Not suitable if usage data drives hard billing cutoffs

---

### ADR-005: Pluggable secret backend interface (Epic 4)

**Status:** Proposed

**Decision:** Introduce `SecretBackend` interface with `LocalBackend` (default) and `VaultBackend` implementations.

**Consequences:**
- ✅ No breaking change; existing deployments use `LocalBackend` by default
- ✅ Vault unlocks dynamic secrets, lease rotation, and audit logs
- ❌ Adds abstraction over what was previously a direct DB call

---

### ADR-006: JSON-RPC 2.0 for MCP server (Epic 2)

**Status:** Proposed

**Decision:** Implement `/mcp` as a standard JSON-RPC 2.0 endpoint per MCP specification. Expose `/mcp/sse` for streaming.

**Consequences:**
- ✅ Compatible with any MCP-compliant AI agent
- ❌ JSON-RPC is verbose compared to REST for simple operations
- ❌ SSE requires careful connection lifecycle management

---

### ADR-007: Heimdall as a separate repository

**Status:** Accepted

**Decision:** Heimdall is a separate private Next.js repo consuming Bifrost's OpenAPI spec via orval codegen.

**Rationale:** Bifrost stays a pure Go project; clean API boundary; independent versioning and deployment.

**Consequences:**
- ✅ No Node.js build pipeline in this repo
- ❌ Two repos to keep in sync; spec drift caught by `swagger-check` CI job

---

### ADR-008: GORM over raw SQL

**Status:** Accepted

**Decision:** Use GORM with PostgreSQL and SQLite drivers.

**Rationale:** `AutoMigrate` at startup; single driver interface covers both dialects; `IsDuplicateError()` abstracts dialect-specific error codes.

**Consequences:**
- ✅ No migration files in early phases; SQLite support free
- ❌ `AutoMigrate` unsafe for destructive changes — needs `golang-migrate` before Phase 3
- ❌ Convention: keep queries explicit, avoid association preloading

---

### ADR-009: chi over gorilla/mux or stdlib ServeMux

**Status:** Accepted

**Decision:** Use `go-chi/chi v5`.

**Rationale:** `gorilla/mux` is archived. chi uses stdlib `context`, is composable via `r.With(...)`, and has no external dependencies. gin/echo add convenience wrappers that conflict with the existing `writeError` + `json.NewEncoder` pattern.

**Consequences:**
- ✅ Stdlib-compatible handlers; no framework lock-in
- ❌ No built-in binding/validation — deliberate; handled explicitly

---

### ADR-010: PostgreSQL + SQLite dual-database support

**Status:** Accepted

**Decision:** Support both via a single GORM `SQLStore`, selected by `BIFROST_DB`.

**Consequences:**
- ✅ New contributors run full stack with no Docker; fast hermetic integration tests
- ❌ SQLite not suitable for multi-instance deployments
- ❌ Schema differences can surface at edges — convention: avoid PostgreSQL-specific GORM features

---

### ADR-011: Cobra CLI wrapping the HTTP server

**Status:** Accepted

**Decision:** Wrap startup logic with `spf13/cobra` in `cmd/bifrost/`. Operational subcommands (`--migrate-only`, `--retention-job`) share the same binary.

**Consequences:**
- ✅ Single binary for all operational tasks; K8s Jobs invoke flags directly
- ❌ Cobra adds ~1MB to binary size — negligible

---

### ADR-012: Stateless tokens — no revocation store

**Status:** Accepted (with known trade-off)

**Decision:** No token blocklist. Emergency revocation via `BIFROST_SIGNING_KEY` rotation (invalidates all tokens instantly).

**Consequences:**
- ✅ No Redis required for auth; horizontal scaling with no shared auth state
- ❌ Compromised token valid until expiry — mitigated by short TTL

---

### ADR-013: zerolog over slog / zap

**Status:** Accepted

**Decision:** Use `rs/zerolog`.

**Rationale:** Zero allocations on the hot path; built-in `console` mode via env var; chainable API. `log/slog` considered but adds boilerplate to achieve the same zero-alloc guarantee.

**Consequences:**
- ✅ No allocations in proxy logging path; `console` mode needs no code change
- ❌ Not stdlib — one external logging dependency

---

### ADR-014: Prometheus over OpenTelemetry metrics

**Status:** Accepted

**Decision:** Use `prometheus/client_golang` directly. No OTel SDK for metrics.

**Rationale:** Prometheus is the de facto K8s standard; scrape-based pull model works out of the box with pod annotations; OTel adds significant setup complexity without portability benefit at current scale. Traces will adopt OTel in Phase 3 with a Prometheus bridge for existing metrics.

**Consequences:**
- ✅ Zero-config scraping in Kubernetes
- ❌ Vendor portability requires adding exporters later

---

### ADR-015: Opaque UUIDs over sequential IDs

**Status:** Accepted

**Decision:** All entity IDs generated via `utils.GenerateID()` (UUID v4).

**Rationale:** No enumeration risk; no coordination needed between `MemoryStore` and `SQLStore`; IDs can be generated before DB insertion.

**Consequences:**
- ✅ Safe to expose in URLs and logs; generation is store-independent
- ❌ 36-char strings vs. 4-byte int — negligible index size impact at current scale

---

### ADR-016: Defer OPA integration to Phase 3

**Status:** Deferred

**Decision:** Authorization remains hardcoded role checks. No OPA until there are >5 distinct authorization rules or a customer requirement for externally auditable policies.

**Consequences:**
- ✅ No new infrastructure; auth logic readable inline in handlers
- ❌ New role or cross-resource rule requires code change and redeploy

---

### ADR-017: No JWT pass-through verification

**Status:** Deferred

**Decision:** Bifrost uses its own HMAC tokens exclusively. No customer JWT verification against a JWKS endpoint.

**Rationale:** JWT verification (JWKS fetch, key rotation, claim validation) is significant complexity; consumers that need IdP integration can place an API gateway in front of Bifrost.

**Consequences:**
- ✅ No JWKS fetching, no token cache, no clock-skew handling
- ❌ Cannot natively enforce IdP-gated proxy access

---

### ADR-018: Scope as a plain string over bitmask or enum

**Status:** Accepted

**Decision:** `VirtualKey.Scope` is a plain string matching `Service.Name`; wildcard `"*"` allows any service.

**Rationale:** Services are user-defined — a bitmask or enum cannot cover arbitrary names without recompilation. Single string comparison at proxy time; self-documenting in logs.

**Consequences:**
- ✅ New services scopeable immediately; no code change required
- ❌ No compile-time validation — invalid scopes return 403 at proxy time

---

### ADR-019: Single binary, no microservices split

**Status:** Accepted

**Decision:** All planes (management API, proxy, MCP, usage tracker, retention) run in one process, selected by startup flags.

**Rationale:** Current traffic does not justify service-mesh overhead. Horizontal pod autoscaling handles mixed workloads adequately until load is measured.

**Trigger for splitting:** Proxy throughput requires dedicated pods; compliance requires isolating secret backend access.

**Consequences:**
- ✅ One image, one deployment, one log stream; no inter-service network calls
- ❌ Memory issue in one component (e.g., webhook dispatcher) affects the entire process

---

## Appendix: Proxy Credential Injection

`Service.CredentialHeader` controls how the root key is forwarded upstream:

| Value | Injected header |
|-------|----------------|
| `""` or `"X-API-Key"` | `X-API-Key: <key>` |
| `"Authorization"` | `Authorization: Bearer <key>` |
| any other value | `<value>: <key>` |

## 11. Future Roadmap

### 11.1 Epic Overview

| Epic | Theme | Priority | Status |
|------|-------|----------|--------|
| Epic 1 | Production Hardening | High | In progress |
| Epic 2 | AI Agent Native / MCP Server | High ⭐ | Planned |
| Epic 3 | Spend & Usage Visibility | Medium | Planned |
| Epic 4 | Vault Backend | Medium | Planned |
| Epic 5 | Heimdall Integration | Medium | Planned |
| Phase 3 | Platform Expansion | Low | Spike only |

---

### 11.2 Epic 1 — Production Hardening

**Goal:** Make the current system production-safe before adding features.

| Story | Description |
|-------|-------------|
| 1.1 | Docker Compose with setup-job |
| 1.2 | GitHub Actions CI (test + build + swagger-check) |
| 1.3 | AES-256-GCM encryption at rest for root keys |
| 1.4 | `--migrate-only` flag + K8s Job manifest |
| 1.5 | Deployment docs (Docker Compose + K8s + Helm) |

**Key decisions pending:** encryption key rotation strategy; `golang-migrate` adoption vs. GORM `AutoMigrate`.

---

### 11.3 Epic 2 — AI Agent Native / MCP Server ⭐

**Goal:** Expose Bifrost as an MCP-compatible tool server so AI agents can request and use virtual keys autonomously.

```
AI Agent (Claude, GPT, etc.)
    │  JSON-RPC 2.0
    ▼
POST /mcp
    ├── tools/list      → available tools
    ├── tools/call      → get_key, create_key, list_services,
    │                      revoke_key, list_rootkeys
    └── resources/read  → service definitions (future)
```

| Story | Description |
|-------|-------------|
| 2.1 | JSON-RPC 2.0 dispatcher + tool registry |
| 2.2 | `get_key` and `list_services` tools |
| 2.3 | `create_key` and `revoke_key` tools |
| 2.4 | SSE transport (`/mcp/sse`) |
| 2.5 | MCP auth (API key) + rate limiting |

**Spike needed:** MCP session lifecycle — stateless vs. stateful connection model for SSE transport.

---

### 11.4 Epic 3 — Spend & Usage Visibility

**Goal:** Give operators visibility into proxy traffic, cost, and budget consumption.

```
Proxy Handler
    │  UsageEvent (async)
    ▼
UsageTracker (buffered chan)
    │  batch insert
    ▼
usage_events table (PostgreSQL)
    │
    ├── GET /v1/usage          → aggregated by key/service/org
    ├── GET /v1/usage/events   → raw event stream (paginated)
    └── Webhook dispatcher     → threshold alerts
```

| Story | Description |
|-------|-------------|
| 3.1 | UsageEvent model + async tracker + batch insert |
| 3.2 | Usage query API (`/v1/usage`) |
| 3.3 | Webhook config + delivery + HMAC signing |
| 3.4 | Retention CronJob + `BIFROST_RETENTION_DAYS` |

**Open question:** Token/cost attribution — parse upstream response bodies (latency + coupling risk) vs. caller-reporting (unreliable).

---

### 11.5 Epic 4 — Vault Backend

**Goal:** Allow enterprises to store root keys in HashiCorp Vault instead of the local encrypted store.

```
BIFROST_SECRET_BACKEND=vault

Proxy Handler
    │  GetSecret(rootKeyID)
    ▼
VaultBackend
    │  GET /v1/secret/data/{rootKeyID}
    ▼
HashiCorp Vault KV v2
```

| Story | Description |
|-------|-------------|
| 4.1 | `SecretBackend` interface + `LocalBackend` refactor |
| 4.2 | `VaultBackend` implementation (KV v2) |
| 4.3 | Vault K8s auth method integration |
| 4.4 | Secret lease renewal + rotation hook |

**Spike needed:** Vault dynamic secrets for upstream APIs (requires upstream support — not yet available for OpenAI).

---

### 11.6 Epic 5 — Heimdall Integration

**Goal:** Keep Bifrost's OpenAPI contract stable and Heimdall's orval codegen always producing correct typed clients.

| Story | Description |
|-------|-------------|
| 5.1 | `swagger-check` CI job blocks PRs with spec drift |
| 5.2 | Heimdall updates to `POST /v1/setup` + typed user API |
| 5.3 | Contract test suite (Heimdall → Bifrost against spec) |
| 5.4 | Versioned spec endpoint (`/docs/openapi/v1.json`) |

---

### 11.7 Phase 3 — Platform Expansion (Spikes)

Directional only — each requires a dedicated spike before story creation.

#### Spike A: OpenTelemetry Traces

Adopt OTel SDK for distributed tracing across the proxy handler. Bridge existing Prometheus metrics via `otelprometheus` to avoid dashboard breakage.

#### Spike B: OPA Authorization

Replace hardcoded role checks with Rego policies evaluated by an embedded OPA engine. Trigger: >5 distinct authorization rules or per-org policy customization requirement.

#### Spike C: Kubernetes Operator + CRDs

Define `BifrostVirtualKey`, `BifrostRootKey`, `BifrostService` CRDs. A controller reconciles CRD spec against the Bifrost API for GitOps-style credential management.

#### Spike D: Multi-region Active-Active

Replace single PostgreSQL primary with CockroachDB or Citus. Redis cluster with cross-region replication for rate limit counters.

#### Spike E: Usage-based Billing Webhooks

Extend Epic 3 webhooks to emit structured events compatible with Stripe Billing metered usage API.

---

### 11.8 Dependency Graph

```
Epic 1 (hardening) ── prerequisite for all other epics

Epic 2 (MCP)    ─── independent
Epic 3 (usage)  ─── independent
Epic 4 (vault)  ─── independent

Epic 5 (Heimdall) ── depends on Epic 1 (stable spec)

Phase 3 Spike A (OTel)    ── after Epic 1
Phase 3 Spike B (OPA)     ── after Epic 2
Phase 3 Spike C (CRDs)    ── after Epic 2 + Epic 4
Phase 3 Spike D (HA)      ── after Epic 3
Phase 3 Spike E (billing) ── after Epic 3
```

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
