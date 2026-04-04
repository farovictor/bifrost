# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make run          # start server with in-memory stores, console logs
make test         # go test ./... -coverprofile=coverage.out
make swagger      # regenerate docs/swagger/ from handler annotations
go test ./tests/  # run integration tests only
go test ./tests/ -run TestFoo  # run a single test
go build ./...    # verify compilation
```

## Architecture

Bifrost is an HTTP proxy that maps short-lived virtual keys to real API credentials. Requests to `/v1/proxy/*` are authenticated by the virtual key itself; all management endpoints (`/v1/keys`, `/v1/rootkeys`, `/v1/services`, `/v1/orgs`, etc.) require `X-API-Key` + `Authorization: Bearer <token>`.

### Key packages

- `routes/` — HTTP handlers as methods on `*routes.Server`. Each resource has its own file (`keys.go`, `rootkeys.go`, `services.go`, `users.go`, `orgs.go`). `server.go` holds the `Server` struct and the shared `writeError()` helper.
- `routes/v1/` — proxy handler (`proxy.go`) and the `Handler` struct. Has its own `writeError()` and `injectCredential()` helpers.
- `middlewares/` — `AuthMiddleware(store)`, `OrgCtxMiddleware(store)`, `RateLimitMiddleware(store)` all take their store as a constructor argument (no globals).
- `pkg/{keys,rootkeys,services,orgs,users}/` — domain types + `Store` interface with `MemoryStore` and `SQLStore` implementations. GORM-based SQL stores.
- `pkg/auth/` — HMAC-SHA256 token sign/verify.
- `cmd/bifrost/` — Cobra CLI wrapping the HTTP API.

### Request flow

```
POST /v1/keys
  → AuthMiddleware (validates X-API-Key via UserStore)
  → OrgCtxMiddleware (decodes bearer token, looks up membership)
  → Server.CreateKey

GET /v1/proxy/path?key=vk-1
  → RateLimitMiddleware (checks+increments counter in Redis or local)
  → Handler.Proxy (validates key, enforces scope+expiry, injects creds, reverse-proxy)
```

### Error responses

All handlers use `writeError(w, message, statusCode)` which writes `{"error":"message"}` with `Content-Type: application/json`.

See [doc/architecture.md](doc/architecture.md) for full architecture details and [doc/api.md](doc/api.md) for HTTP API reference.

## Test helpers (`tests/`)

- `newTestServer(t)` — `*routes.Server` with empty in-memory stores
- `newTestEnv(t)` — seeds one user, builds router, returns `*TestEnv{Server, Router, User, Token}`
- `env.Authorize(req)` — sets `X-API-Key` + `Authorization: Bearer` headers
- `errorBody(t, rr)` — decodes `{"error":"..."}` and returns the message string
- `setupRouter(s)` — builds a chi router matching `main.go` layout; used in `routes_test.go`

Tests never touch a real database or Redis. Rate-limit tests use timestamp-unique key IDs to avoid cross-run interference with a running Redis.

## Store interface rules

Every store interface follows this pattern:
```
Create, Get, Delete, [Update], List, [ListBy*]
```
When adding a new method to an interface, implement it on both `MemoryStore` and `SQLStore`.

## Phase roadmap (project memory)

- **Phase 0** ✅ Server struct DI, proxy decoupled from auth, TestEnv helpers
- **Phase 1** ✅ List endpoints, org/membership HTTP endpoints, JSON errors, configurable credential injection
- **Phase 2** ✅ 77.6% test coverage, SQL store integration tests, middleware tests, error-path coverage
- **Phase 2.5** ✅ Service update endpoint, token refresh, signing key hint, CI fix
- **Phase 3** Vault backend, management dashboard, OPA, Kubernetes operator

### Backlog (pre-Phase 3)
- **Configurable token TTL** — `POST /v1/users` and `POST /v1/token/refresh` should accept an optional `ttl` field (e.g. `"1h"`); fall back to `BIFROST_TOKEN_TTL` env var (default `24h`). Needed for ephemeral setups that require very short-lived tokens.
