# HTTP API Reference

All error responses are JSON: `{"error": "message"}`.

## Health / Version

```
GET /healthz          → 200 "ok"
GET /version          → 200 {"version":"..."}
```

## Authentication

Most `/v1` management endpoints require **both**:
- `X-API-Key: <api_key>` — the user's API key
- `Authorization: Bearer <token>` — the signed auth token returned by `user-add` or `POST /v1/users`

The token encodes `user_id`, `org_id`, and expiry (24 h). It is verified with HMAC-SHA256 using `BIFROST_SIGNING_KEY`.

**Exceptions:**
- `POST /v1/users`, `GET /v1/user`, `POST /v1/user/rootkeys` — bearer token only (no API key required)
- `GET /v1/proxy/*` — virtual key only (`X-Virtual-Key` header or `key` query param)
- `GET /healthz`, `GET /version` — no auth

In `test` mode or SQLite mode, any bearer token is accepted and `BIFROST_STATIC_API_KEY` is used instead of a user lookup.

## Users

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
Supply `org_id` to join an existing org, or `org_name` to create a new one.
`role`: `owner`, `admin`, or `member` (default: `member`).
Returns `201` with the user object and a signed `token` field.

## Organizations

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

## Root Keys

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

## Services

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

## Virtual Keys

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

## Proxy

```
GET|POST|... /v1/proxy/{path}
  X-Virtual-Key: <key>     (or ?key=<key> query param)
```

Bifrost validates the key, enforces scope and rate limit, strips the virtual key from the forwarded request, injects the root key credential (per `credential_header`), and proxies to the upstream service endpoint.

## Metrics

When `BIFROST_ENABLE_METRICS=true`:
```
GET /metrics     → Prometheus text format
```

Available metrics: `request_total`, `request_duration_seconds`, `key_usage_total`.
