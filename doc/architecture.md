# Architecture

## Overview

Bifrost is an HTTP proxy that maps short-lived virtual keys to real API credentials. Requests to `/v1/proxy/*` are authenticated by the virtual key itself; all management endpoints (`/v1/keys`, `/v1/rootkeys`, `/v1/services`, `/v1/orgs`, etc.) require `X-API-Key` + `Authorization: Bearer <token>`.

## Router layout

```
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

## Key packages

- `routes/` — HTTP handlers as methods on `*routes.Server`. Each resource has its own file (`keys.go`, `rootkeys.go`, `services.go`, `users.go`, `orgs.go`). `server.go` holds the `Server` struct and the shared `writeError()` helper.
- `routes/v1/` — proxy handler (`proxy.go`) and the `Handler` struct. Has its own `writeError()` and `injectCredential()` helpers.
- `middlewares/` — `AuthMiddleware(store)`, `OrgCtxMiddleware(store)`, `RateLimitMiddleware(store)` — all take their store as a constructor argument (no globals).
- `pkg/{keys,rootkeys,services,orgs,users}/` — domain types + `Store` interface with `MemoryStore` and `SQLStore` implementations (GORM).
- `pkg/database/` — `Connect(dbType, dsn)` and `IsDuplicateError(err)` (handles both PostgreSQL and SQLite duplicate-key errors).
- `pkg/auth/` — HMAC-SHA256 token sign/verify.
- `cmd/bifrost/` — Cobra CLI wrapping the HTTP API.

## Request flow

```
POST /v1/keys
  → AuthMiddleware (validates X-API-Key via UserStore)
  → OrgCtxMiddleware (decodes bearer token, looks up membership)
  → Server.CreateKey

GET /v1/proxy/path?key=vk-1
  → RateLimitMiddleware (checks+increments counter in Redis or local)
  → Handler.Proxy (validates key, enforces scope+expiry, injects creds, reverse-proxy)
```

## Store interface pattern

Every store interface follows:
```
Create, Get, Delete, [Update], List, [ListBy*]
```
Each domain package has:
- `MemoryStore` — in-memory implementation (tests / no-DSN mode)
- `SQLStore` — GORM-based SQL implementation

When adding a method to an interface, implement it on both.

## Proxy credential injection

`Service.CredentialHeader` controls how the root key is forwarded upstream:
- `""` / `"X-API-Key"` → `X-API-Key: <key>` (default)
- `"Authorization"` → `Authorization: Bearer <key>`
- anything else → `<header>: <key>`

## Roles

- `owner` — full control over the organization and its members
- `admin` — manage resources, cannot demote owners
- `member` — limited to explicitly granted resources

## Planned Extensions

- Open Policy Agent (OPA) integration for dynamic authorization
- Multiple target backend presets (OpenAI, Stripe, etc.)
- Web-based management dashboard
- JWT pass-through with verification hooks
- Vault secret backend
- Kubernetes operator with CRD support
