# Getting Started

This guide walks you through running Bifrost locally and making your first proxied request in about 5 minutes. No database or Redis required.

## Prerequisites

- Go 1.23.8 — install with `make setup`, or check your version with `go version`
- `curl` (any version)

## 1. Start the server

Clone the repo and start Bifrost in **test mode**. Test mode uses in-memory stores and accepts a fixed API key (`secret` by default), so you can skip user management entirely.

```bash
BIFROST_MODE=test make run
```

You should see:

```
INF In-Memory Store set
INF Initializing Server ...
```

Verify the server is up:

```bash
curl http://localhost:3333/healthz
# → ok
```

> **What test mode does:** any `Authorization: Bearer <anything>` token is accepted, and the static API key (`secret`) is used instead of a real user lookup. Great for local exploration — not for production.

## 2. Register a root key

A **root key** is a real credential that Bifrost will inject into upstream requests on your behalf. For this tutorial we'll use a placeholder — replace it with a real API key when proxying an authenticated upstream.

```bash
curl -s -X POST http://localhost:3333/v1/rootkeys \
  -H "X-API-Key: secret" \
  -H "Authorization: Bearer tutorial" \
  -H "Content-Type: application/json" \
  -d '{"id": "my-root", "api_key": "sk-placeholder"}' | jq .
```

```json
{"id":"my-root","api_key":"sk-placeholder"}
```

## 3. Register a service

A **service** maps a name to an upstream endpoint and tells Bifrost which root key to use and how to inject it.

We'll point at [httpbin.org](https://httpbin.org) — a public HTTP echo service that's handy for testing:

```bash
curl -s -X POST http://localhost:3333/v1/services \
  -H "X-API-Key: secret" \
  -H "Authorization: Bearer tutorial" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "httpbin",
    "endpoint": "https://httpbin.org",
    "root_key_id": "my-root",
    "credential_header": "X-API-Key"
  }' | jq .
```

```json
{"id":"httpbin","endpoint":"https://httpbin.org","root_key_id":"my-root","credential_header":"X-API-Key"}
```

`credential_header` controls how the root key is forwarded:
- `"X-API-Key"` (default) → `X-API-Key: sk-placeholder`
- `"Authorization"` → `Authorization: Bearer sk-placeholder`
- any other string → `<that-header>: sk-placeholder`

## 4. Issue a virtual key

A **virtual key** is what you hand to clients. It is scoped, rate-limited, and time-bound — and it never exposes the real credential.

```bash
curl -s -X POST http://localhost:3333/v1/keys \
  -H "X-API-Key: secret" \
  -H "Authorization: Bearer tutorial" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "vk-demo",
    "target": "httpbin",
    "scope": "read",
    "expires_at": "2027-01-01T00:00:00Z",
    "rate_limit": 60
  }' | jq .
```

```json
{"id":"vk-demo","scope":"read","target":"httpbin","expires_at":"2027-01-01T00:00:00Z","rate_limit":60}
```

- `scope: "read"` — only GET/HEAD requests are allowed
- `rate_limit: 60` — max 60 requests per minute
- `expires_at` — the key is automatically rejected after this timestamp

## 5. Make a proxied request

Send a request using only the virtual key. Bifrost validates the key, enforces scope and rate limit, strips the virtual key, injects the root key, and forwards to the upstream.

```bash
curl -s http://localhost:3333/v1/proxy/get \
  -H "X-Virtual-Key: vk-demo" | jq .headers
```

You'll see httpbin echo back the headers Bifrost forwarded — including the injected `X-Api-Key`:

```json
{
  "Host": "httpbin.org",
  "X-Api-Key": "sk-placeholder",
  ...
}
```

The client only ever sees `vk-demo`. The real credential is never exposed.

You can also pass the key as a query param:

```bash
curl -s "http://localhost:3333/v1/proxy/get?key=vk-demo" | jq .headers
```

## 6. Verify the key constraints

**Scope enforcement** — a `read`-scoped key rejects write methods:

```bash
curl -s -X POST http://localhost:3333/v1/proxy/post \
  -H "X-Virtual-Key: vk-demo"
# → {"error":"forbidden: write scope required"}
```

**Expired key** — create a key that's already expired:

```bash
curl -s -X POST http://localhost:3333/v1/keys \
  -H "X-API-Key: secret" \
  -H "Authorization: Bearer tutorial" \
  -H "Content-Type: application/json" \
  -d '{"id":"vk-expired","target":"httpbin","scope":"read","expires_at":"2020-01-01T00:00:00Z","rate_limit":10}'

curl -s http://localhost:3333/v1/proxy/get \
  -H "X-Virtual-Key: vk-expired"
# → {"error":"key expired"}
```

## 7. Revoke a key

```bash
curl -s -X DELETE http://localhost:3333/v1/keys/vk-demo \
  -H "X-API-Key: secret" \
  -H "Authorization: Bearer tutorial"
# → 204 No Content

curl -s http://localhost:3333/v1/proxy/get \
  -H "X-Virtual-Key: vk-demo"
# → {"error":"key not found"}
```

---

## What's next

| Goal | Where to look |
|---|---|
| Use a real database | [configuration.md](configuration.md) — PostgreSQL or SQLite setup |
| Create real users and tokens | [api.md](api.md) — `POST /v1/users` |
| Run with Docker Compose | [configuration.md](configuration.md) — Docker Compose section |
| Full API reference | [api.md](api.md) |
| CLI usage | [cli.md](cli.md) |
| Code architecture | [architecture.md](architecture.md) |
