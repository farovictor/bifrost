# Bifrost Product Requirements Document (PRD)

## Change Log

| Date | Version | Description | Author |
|---|---|---|---|
| 2026-04-04 | 1.0 | Initial PRD created from Project Brief | Morgan |

---

## 1. Goals and Background Context

### Goals
- Enable any team to issue, manage, and revoke virtual API keys without exposing real credentials to consumers
- Provide per-key usage visibility (request counts, rate limits, expiry) with zero changes to existing API call code
- Support multi-user and multi-org access control out of the box, free and self-hosted
- Deliver a < 10-minute setup experience from `git clone` to first proxied request
- Position Bifrost as the standard API key management layer for AI agent workloads

### Background Context

API keys are the primary credential for third-party services, yet most teams manage them unsafely — shared across employees, hardcoded in codebases, and rotated only after incidents. The root problem is that there is no lightweight, provider-agnostic layer between consumers and real credentials. When a key leaks or an employee leaves, rotating it breaks every service simultaneously.

Bifrost solves this by introducing a virtual key layer: consumers receive short-lived, scoped virtual keys while real credentials stay server-side. This pattern is well-established in enterprise API gateways (Kong, Apigee) but unavailable to small teams without significant infrastructure investment. Bifrost fills that gap — a single Go binary, optional Postgres and Redis, deployable in minutes.

---

## 2. Requirements

### Functional Requirements

- **FR1:** Users can register a real API key for an upstream service via the management API
- **FR2:** Administrators can create virtual keys scoped to a specific service, with configurable expiry and rate limits
- **FR3:** Virtual keys can be issued to individual users or service accounts within an organization
- **FR4:** Requests authenticated with a valid virtual key are proxied to the upstream service with the real credential injected
- **FR5:** Virtual keys can be revoked instantly without affecting other keys or rotating the real credential
- **FR6:** The system enforces rate limits per virtual key (requests per time window)
- **FR7:** Virtual keys respect expiry — expired keys are rejected with a clear error response
- **FR8:** All management endpoints require authentication via `X-API-Key` + `Authorization: Bearer <token>`
- **FR9:** Administrators can list, inspect, and delete keys, users, orgs, and services via REST API
- **FR10:** The proxy returns structured JSON errors (`{"error":"..."}`) for all failure cases
- **FR11:** Organizations can have multiple members with scoped access to key management
- **FR12:** The system exposes a `/metrics` endpoint with Prometheus-compatible usage data

### Non-Functional Requirements

- **NFR1:** Proxy overhead must add < 10ms latency per request under normal load
- **NFR2:** The service must run as a single binary with no required external dependencies for development (`make run` with in-memory stores)
- **NFR3:** PostgreSQL and SQLite must both be supported as production database backends
- **NFR4:** Redis is optional — rate limiting must fall back to in-memory when Redis is unavailable
- **NFR5:** All credentials stored server-side must never appear in proxy response bodies or logs
- **NFR6:** Test coverage must be maintained at ≥ 75% across the codebase
- **NFR7:** The management API must follow RESTful conventions with consistent JSON error responses
- **NFR8:** The service must support structured logging (zerolog) with configurable log levels
- **NFR9:** OpenAPI spec must be auto-generated and served at `/docs/openapi.json` and `/docs/openapi.yaml` for all management endpoints
- **NFR10:** System initialization must be gated behind a one-shot `POST /v1/setup` endpoint that returns 409 once any user exists. `POST /v1/users` must require authentication at all times
- **NFR11:** Bifrost supports active-passive HA for production deployments — multiple instances behind a load balancer sharing PostgreSQL and Redis, with no instance-local state in the request path. Single-instance deployment is fully supported and the default for local development
- **NFR12:** All stored data has configurable retention with the following defaults:
  - Virtual keys: deleted on revocation or expiry + 90 day audit trail
  - Usage events (request history): 30 days (`BIFROST_USAGE_RETENTION_DAYS`, default: 30)
  - Audit logs (key create/revoke/update): 365 days (`BIFROST_AUDIT_RETENTION_DAYS`, default: 365)
  - Webhook delivery events: 30 days

  Retention cleanup runs as a background job on startup and daily thereafter.

  When `BIFROST_LOG_SINK=bucket`, audit and usage logs are streamed to an external blob store (S3/GCS). In this mode, Bifrost does not manage log retention — lifecycle policies are configured directly on the bucket. DB-side retention still applies to operational data (keys, webhooks).
- **NFR13:** Bearer token expiry must be configurable via `BIFROST_TOKEN_TTL` (duration string, e.g. `24h`, `7d`). Default: `24h`. Applies to tokens issued by `/v1/setup`, `/v1/users`, and `/v1/token/refresh`. Global setting — no per-user or per-org overrides in MVP.

---

## 3. User Interface Design Goals

**Bifrost is API-only by design — there is no UI in this repo, now or in future phases.**

Heimdall is the officially supported management UI, distributed as a companion service in a separate repository (`/Workspace/Personal/heimdall`). Bifrost exposes `/docs/openapi.json`; Heimdall generates a fully typed HTTP client from it via orval.

**Bifrost's UI commitments:**
- OpenAPI spec at `/docs/openapi.json` — the contract Heimdall depends on; must remain stable and versioned
- Consistent JSON error responses (`{"error":"..."}`) across all endpoints
- Meaningful HTTP status codes — Heimdall's UX depends on correct status semantics
- `/metrics` Prometheus endpoint — Heimdall or any observability layer can consume it

**Out of scope for Bifrost (belongs in Heimdall's PRD):**
- Key management dashboard
- Usage graphs and org administration
- Login / authentication screens
- Any frontend build tooling or assets

---

## 4. Technical Assumptions

**Repository Structure:** Monorepo — single Go module (`github.com/farovictor/bifrost`)

**Service Architecture:** Monolith — single binary serving management API + proxy handler. No microservices planned for MVP.

**Testing Requirements:** Unit + integration. Current coverage 77.6%. Tests never touch a real database — `MemoryStore` for unit tests, SQL integration tests in `tests/`. Coverage must stay ≥ 75%.

**Additional Technical Assumptions:**
- Go 1.23.8, chi v5 router, GORM v1.30
- PostgreSQL (production) + SQLite (lightweight/dev) both supported
- Redis optional — rate limiting falls back to in-memory when unavailable
- Prometheus metrics at `/metrics` (opt-in via `BIFROST_ENABLE_METRICS`)
- OpenAPI spec auto-generated via `swaggo/swag`, served at `/docs/openapi.json` and `/docs/openapi.yaml`
- CORS configurable via `BIFROST_CORS_ORIGINS` (default: `*` for dev)
- Bifrost is headless — no frontend bundled; Heimdall is the companion UI (separate repo)
- `POST /v1/setup` is the one-shot bootstrap endpoint; all other management endpoints require auth
- HMAC-SHA256 token signing via `pkg/auth/`

---

## 5. Epic List

- **Epic 1 — Production Hardening:** Establish a secure, production-ready deployment baseline. Covers bootstrap flow (`/v1/setup`), CORS configuration, credential encryption at rest, deployment docs, and Docker Compose packaging.

- **Epic 2 — AI Agent Native (MCP Server Mode):** Expose Bifrost as a Model Context Protocol server so AI agents can request and use virtual keys on-demand. Covers MCP server implementation, agent identity header injection, one-shot ephemeral keys, and token budget enforcement.

- **Epic 3 — Spend & Usage Visibility:** Give operators full visibility into API consumption per key. Covers token-count tracking for LLM APIs, per-key cost attribution, usage history endpoints, and webhook alerts.

- **Epic 4 — Vault Backend:** Integrate enterprise-grade credential storage. Covers HashiCorp Vault, AWS Secrets Manager, and GCP Secret Manager as pluggable backends.

- **Epic 5 — Heimdall Integration & Distribution:** Officially support Heimdall as the companion management UI. Covers OpenAPI contract stability, Docker Compose co-deployment, typed client codegen pipeline, and end-to-end setup flow validation.

---

## 6. Epic Details

### Epic 1 — Production Hardening

*Goal: Establish a secure, production-ready deployment baseline so Bifrost can be safely exposed beyond localhost. Delivers a hardened instance with encrypted credentials, containerized packaging, and verified setup flow.*

---

**Story 1.1 — Bootstrap Endpoint** ✅ *Shipped*
As a first-time operator, I want a one-shot setup endpoint so that I can initialize Bifrost without pre-existing credentials.

*Acceptance Criteria:*
1. `POST /v1/setup` returns 201 with `api_key` and `token` when no users exist
2. `POST /v1/setup` returns 409 `{"error":"already initialized"}` on all subsequent calls
3. `POST /v1/users` requires Bearer token at all times — no unauthenticated access
4. OpenAPI spec documents `/v1/setup` with typed request/response schemas

---

**Story 1.2 — CORS Configuration** ✅ *Shipped*
As an operator, I want configurable CORS so that Heimdall (or any web client) can make cross-origin requests to Bifrost.

*Acceptance Criteria:*
1. `OPTIONS` preflight requests return correct `Access-Control-Allow-*` headers
2. Default allows all origins (`*`) for development
3. `BIFROST_CORS_ORIGINS` env var restricts origins in production (comma-separated)
4. Allowed headers include `X-API-Key`, `Authorization`, `Content-Type`

---

**Story 1.3 — API Key Encryption at Rest**
As a security-conscious operator, I want real API keys encrypted in the database so that a DB breach does not expose upstream credentials.

*Acceptance Criteria:*
1. Root keys (real upstream credentials) are encrypted before writing to DB using AES-256-GCM
2. Decryption happens in-process at proxy time — plaintext never persisted
3. Encryption key sourced from `BIFROST_ENCRYPTION_KEY` env var (32-byte hex)
4. Bifrost refuses to start if `BIFROST_ENCRYPTION_KEY` is unset in production mode
5. Existing unencrypted records are handled gracefully with a migration path documented

---

**Story 1.4 — Docker Compose Packaging**
As an operator, I want an official Docker Compose file so that I can run Bifrost with Postgres and Redis in one command.

*Acceptance Criteria:*
1. `docker-compose.yml` at repo root starts Bifrost + Postgres + Redis with `docker compose up`
2. Bifrost container builds from a `Dockerfile` in the repo root using a multi-stage Go build
3. All required env vars have documented defaults in `.env.example`
4. `BIFROST_CORS_ORIGINS`, `BIFROST_ENCRYPTION_KEY`, `DATABASE_DSN` are included in `.env.example`
5. Health check passes (`GET /healthz` returns 200) before Bifrost accepts traffic

---

**Story 1.5 — Deployment Hardening Docs**
As an operator, I want a deployment guide so that I know how to securely expose Bifrost in production.

*Acceptance Criteria:*
1. `doc/deployment.md` covers: env vars reference, Docker Compose setup, reverse proxy config (nginx/caddy example), firewall recommendations
2. Documents which endpoints must not be publicly exposed (`/v1/setup`, `/metrics`, `/docs`)
3. Includes a security checklist: encryption key set, CORS locked down, management API behind VPN/firewall
4. Getting-started tutorial updated to reference `/v1/setup` instead of manual curl

---

### Epic 2 — AI Agent Native (MCP Server Mode) ⭐

*Goal: Expose Bifrost as a Model Context Protocol (MCP) server so AI agents can discover available services, request virtual keys on-demand, and make governed API calls.*

---

**Story 2.1 — MCP Server Endpoint**
As an AI agent framework operator, I want Bifrost to expose an MCP server endpoint so that agents can connect to it as a tool provider.

*Acceptance Criteria:*
1. `POST /mcp` handles JSON-RPC 2.0 MCP requests (initialize, tools/list, tools/call)
2. MCP endpoint is documented in OpenAPI spec and `doc/deployment.md`
3. MCP server identifies itself with name `bifrost` and version matching the binary version
4. `tools/list` returns the available Bifrost tools with descriptions and input schemas
5. MCP endpoint requires `X-API-Key` authentication — unauthenticated requests return 401

---

**Story 2.2 — MCP Tool: Request Virtual Key**
As an AI agent, I want to request a virtual key for a named service so that I can make authenticated API calls without knowing the real credential.

*Acceptance Criteria:*
1. MCP tool `request_key` accepts `service_name` (required), `ttl_seconds` (optional, default 3600), `rate_limit` (optional)
2. Returns `virtual_key` and `expires_at` on success
3. Returns structured error if service does not exist or rate limit is invalid
4. Issued key is scoped to the calling user's org — agents cannot request keys across orgs
5. Key appears in `GET /v1/keys` list with `source: mcp` label

---

**Story 2.3 — MCP Tool: List Available Services**
As an AI agent, I want to discover which upstream services are available so that I can request the appropriate key.

*Acceptance Criteria:*
1. MCP tool `list_services` returns all services the calling user's org has access to
2. Response includes `name`, `description`, and `base_url` for each service
3. Real upstream credentials are never included in the response
4. Empty list returned (not error) when no services are configured

---

**Story 2.4 — Agent Identity Header Injection**
As an operator, I want Bifrost to inject agent identity metadata into proxied requests so that upstream services can attribute calls to specific agents.

*Acceptance Criteria:*
1. Proxied requests include `X-Bifrost-Agent-ID` header when key was issued via MCP
2. `X-Bifrost-Key-ID` is injected on all proxied requests (MCP and non-MCP)
3. Headers are injected server-side — consumers cannot spoof them
4. Headers are documented in `doc/api.md`

---

**Story 2.5 — Ephemeral One-Shot Keys**
As a CI/CD pipeline operator, I want keys that expire after a single use so that leaked pipeline credentials cannot be replayed.

*Acceptance Criteria:*
1. `POST /v1/keys` and `request_key` MCP tool accept `one_shot: true` parameter
2. One-shot keys are invalidated immediately after the first successful proxied request
3. A second request with a one-shot key returns 401 `{"error":"key already used"}`
4. One-shot keys appear in `GET /v1/keys` with `one_shot: true` and `used: true/false` fields
5. Unused one-shot keys expire normally at their `expires_at` time

---

### Epic 3 — Spend & Usage Visibility

*Goal: Give operators full visibility into API consumption per virtual key. Delivers per-key request tracking, token-count attribution for LLM APIs, cost estimation, and alerting.*

---

**Story 3.1 — Per-Key Request History**
As an operator, I want to see request history per virtual key so that I can audit who called what and when.

*Acceptance Criteria:*
1. `GET /v1/keys/{id}/usage` returns a paginated list of request events (timestamp, status code, upstream service, latency)
2. Events are stored per proxied request — no sampling
3. History is queryable by date range via `?from=` and `?to=` query params
4. Response includes total request count for the queried period
5. Usage data retention governed by `BIFROST_USAGE_RETENTION_DAYS` (default: 30); when `BIFROST_LOG_SINK=bucket` retention is managed by bucket lifecycle policy

---

**Story 3.2 — Token Count Tracking for LLM APIs**
As an operator running LLM workloads, I want per-key token consumption tracked so that I can attribute costs to specific users or agents.

*Acceptance Criteria:*
1. Bifrost parses `usage.prompt_tokens`, `usage.completion_tokens`, `usage.total_tokens` from upstream LLM responses (OpenAI-compatible format)
2. Token counts are stored per request event and aggregated per key
3. `GET /v1/keys/{id}/usage` includes `total_tokens` in the response summary
4. Non-LLM responses (no `usage` field) are stored without token data — no error
5. Token tracking is opt-in via `BIFROST_TRACK_TOKENS=true`

---

**Story 3.3 — Token Budget Enforcement**
As an operator, I want to set a token budget per virtual key so that a single agent or user cannot consume unbounded LLM tokens.

*Acceptance Criteria:*
1. `POST /v1/keys` and `PUT /v1/keys/{id}` accept `token_budget` field (integer, total tokens)
2. Proxied requests that would exceed the budget are rejected with 429 `{"error":"token budget exceeded"}`
3. `GET /v1/keys/{id}` includes `token_budget`, `tokens_used`, and `tokens_remaining` fields
4. Budget enforcement only activates when `BIFROST_TRACK_TOKENS=true` and `token_budget > 0`
5. Budget reset is manual via `POST /v1/keys/{id}/reset-budget`

---

**Story 3.4 — Webhook Alerts**
As an operator, I want webhook notifications for key usage events so that I can react to anomalies without polling the API.

*Acceptance Criteria:*
1. Webhook URL configurable per org via `POST /v1/orgs/{id}/webhook`
2. Events fired: `key.rate_limited`, `key.expired`, `key.budget_exceeded`, `key.used` (one-shot keys)
3. Webhook payload is JSON with `event`, `key_id`, `org_id`, `timestamp`, and event-specific data
4. Failed webhook deliveries are retried up to 3 times with exponential backoff
5. Webhook delivery log accessible via `GET /v1/orgs/{id}/webhook/events`

---

### Epic 4 — Vault Backend

*Goal: Allow enterprises to store real API credentials in their existing secrets infrastructure instead of Bifrost's database.*

---

**Story 4.1 — Pluggable Secret Backend Interface**
As a developer, I want a clean interface for secret backends so that new providers can be added without changing core proxy logic.

*Acceptance Criteria:*
1. `SecretBackend` interface defined in `pkg/secrets/` with `Get(id string)`, `Set(id, value string)`, `Delete(id string)` methods
2. Existing DB-based root key storage reimplemented as `DatabaseBackend` satisfying the interface
3. Active backend selected via `BIFROST_SECRET_BACKEND` env var (`database`, `vault`, `aws`, `gcp`)
4. Proxy credential injection uses the interface — no backend-specific code in `routes/v1/proxy.go`
5. Both `MemoryBackend` (tests) and `DatabaseBackend` (default) pass the full existing test suite

---

**Story 4.2 — HashiCorp Vault Backend**
As an enterprise operator using Vault, I want Bifrost to retrieve real API keys from Vault at proxy time so that credentials never touch Bifrost's database.

*Acceptance Criteria:*
1. `VaultBackend` satisfies `SecretBackend` interface using Vault KV v2
2. Configured via `BIFROST_VAULT_ADDR`, `BIFROST_VAULT_TOKEN`, `BIFROST_VAULT_PATH`
3. Vault token is validated at startup — Bifrost refuses to start if Vault is unreachable and backend is `vault`
4. Vault errors during proxy requests return 502 — never expose Vault error details to callers
5. Integration test covers happy path + Vault unavailable scenario using a test Vault instance

---

**Story 4.3 — AWS Secrets Manager Backend**
As an enterprise operator on AWS, I want Bifrost to retrieve credentials from AWS Secrets Manager so that I can use native AWS IAM policies to control access.

*Acceptance Criteria:*
1. `AWSBackend` satisfies `SecretBackend` interface using `aws-sdk-go-v2`
2. Configured via standard AWS env vars or instance role
3. Secret name prefix configurable via `BIFROST_AWS_SECRET_PREFIX` (default: `bifrost/`)
4. AWS errors during proxy requests return 502 — never expose AWS error details to callers
5. Supports both string and JSON secrets (extracts `value` key from JSON automatically)

---

**Story 4.4 — GCP Secret Manager Backend**
As an enterprise operator on GCP, I want Bifrost to retrieve credentials from GCP Secret Manager so that I can use native GCP IAM to control access.

*Acceptance Criteria:*
1. `GCPBackend` satisfies `SecretBackend` interface using `cloud.google.com/go/secretmanager`
2. Configured via `BIFROST_GCP_PROJECT` and standard GCP application default credentials
3. Secret version defaults to `latest` — configurable via `BIFROST_GCP_SECRET_VERSION`
4. GCP errors during proxy requests return 502 — never expose GCP error details to callers
5. Integration test covers happy path using GCP Secret Manager emulator

---

### Epic 5 — Heimdall Integration & Distribution

*Goal: Officially support Heimdall as the companion management UI for Bifrost. Delivers a stable OpenAPI contract, Docker Compose co-deployment, and a validated end-to-end setup flow.*

---

**Story 5.1 — OpenAPI Contract Stability**
As a Heimdall developer, I want Bifrost's OpenAPI spec to have strict, typed schemas for all endpoints so that orval generates accurate TypeScript clients with no `unknown` types.

*Acceptance Criteria:*
1. All request bodies and response objects use named Go structs with swaggo annotations — no `object` or `interface{}` body types remain
2. `make swagger` is part of CI — PRs that change handler signatures without updating the spec are blocked
3. All error responses reference `ErrorResponse` — no inline error schemas
4. Breaking API changes increment the spec version and are documented in `CHANGELOG.md`
5. Heimdall's orval codegen produces zero `{ [key: string]: unknown }` types after regenerating

---

**Story 5.2 — End-to-End Setup Flow Validation**
As a new Heimdall user, I want the `/setup` flow to work out of the box against a fresh Bifrost instance so that I can go from zero to dashboard in under 5 minutes.

*Acceptance Criteria:*
1. Heimdall `/setup` calls `POST /v1/setup` with `name` and `email` fields
2. On success, `api_key` and `token` are stored in Zustand and user is redirected to `/dashboard`
3. A second visit to `/setup` when Bifrost is already initialized shows a clear "already set up" message (handles 409 gracefully)
4. End-to-end test covers: fresh Bifrost → `/setup` → credential storage → `/dashboard` loads
5. Heimdall orval client is regenerated from the updated spec before merging setup page changes

---

**Story 5.3 — Docker Compose Co-Deployment**
As an operator, I want a single `docker compose up` to start both Bifrost and Heimdall together so that I can run the full stack without manual configuration.

*Acceptance Criteria:*
1. `docker-compose.yml` includes a `heimdall` service pointing to the Heimdall image
2. `NEXT_PUBLIC_BIFROST_URL` is pre-configured to the Bifrost service name within the Compose network
3. `BIFROST_CORS_ORIGINS` is automatically set to Heimdall's internal URL
4. Stack starts in correct order: Postgres → Redis → Bifrost (healthy) → Heimdall
5. `docker compose up` from a fresh clone produces a working dashboard at `http://localhost:3001`

---

**Story 5.4 — Heimdall Key Management MVP**
As a Heimdall user, I want to create, list, and revoke virtual keys from the dashboard so that I can manage API access without using curl.

*Acceptance Criteria:*
1. Dashboard shows all virtual keys for the authenticated user's org with status (active, expired, revoked)
2. "Create Key" flow accepts service selection, expiry, rate limit, and optional token budget
3. Key revocation is a single click with a confirmation prompt — reflected immediately in the list
4. Keys created via MCP are visible in the dashboard with `source: mcp` label
5. All API calls use the orval-generated typed client — no hand-written fetch calls to Bifrost

---

## 7. Checklist Results Report

| Category | Status | Notes |
|---|---|---|
| 1. Problem Definition & Context | **PASS** | Competitive analysis present; personas defined; KPIs set |
| 2. MVP Scope Definition | **PASS** | In/out scope clear; MVP shipped (phases 0-2); epics cover next phase |
| 3. User Experience Requirements | **PARTIAL** | Bifrost is API-only; Heimdall UX deferred to separate PRD |
| 4. Functional Requirements | **PASS** | 12 FRs, 13 NFRs; all testable; ACs per story |
| 5. Non-Functional Requirements | **PASS** | Performance, security, HA, retention, token TTL all addressed |
| 6. Epic & Story Structure | **PASS** | 5 epics, logically sequential; stories vertical-sliced |
| 7. Technical Guidance | **PASS** | Full stack documented; trade-offs articulated |
| 8. Cross-Functional Requirements | **PARTIAL** | Retention policy added; monitoring plan deferred to Epic 1 ops story |
| 9. Clarity & Communication | **PARTIAL** | No formal stakeholder list; approval process undefined |

**Overall completeness:** 89%
**MVP scope:** Just Right
**Decision: READY FOR ARCHITECT** — Epics 1 and 5 can begin immediately. Epics 2 and 4 benefit from a brief architect spike before estimation.

---

## 8. Next Steps

### Immediate Actions
1. Share PRD with collaborators for review
2. Architect spike on MCP server implementation (Epic 2) before story estimation
3. Architect spike on pluggable secret backend interface (Epic 4, Story 4.1)
4. Heimdall team applies integration fixes (Story 5.2) and regenerates orval client
5. Begin Epic 1 implementation — Stories 1.3, 1.4, 1.5 are unblocked

### UX Expert Prompt
Bifrost is a headless API proxy — no UI required. Forward this PRD to the Heimdall team to drive UX work. Reference `docs/brief.md` for product context and `docs/swagger/swagger.json` for the current API contract.

### Architect Prompt
Review this PRD and produce an architecture document covering: (1) MCP server implementation approach for Epic 2, (2) pluggable secret backend interface design for Epic 4, (3) HA deployment topology for NFR11, (4) encryption key management strategy for Story 1.3. Use `doc/architecture.md` as the output target.
