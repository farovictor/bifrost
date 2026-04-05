# Deployment Guide

This guide covers everything you need to run Bifrost securely in production: environment variables, Docker Compose setup, reverse proxy configuration, and a security checklist.

---

## Environment Variables

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `BIFROST_PORT` | `3333` | No | HTTP listen port |
| `BIFROST_MODE` | `production` | No | `production` or `test` (test disables auth — never use in prod) |
| `BIFROST_DB` | `sqlite` | No | `sqlite` or `postgres` |
| `POSTGRES_DSN` | — | Yes (postgres) | PostgreSQL connection string, e.g. `postgres://user:pass@host:5432/bifrost?sslmode=require` |
| `BIFROST_SIGNING_KEY` | — | **Yes** | HMAC-SHA256 key for token signing. Generate with `openssl rand -base64 32` |
| `BIFROST_ENCRYPTION_KEY` | — | **Yes (prod)** | 32-byte AES-256-GCM key for root key encryption. Generate with `openssl rand -base64 32 \| head -c 32` |
| `BIFROST_CORS_ORIGINS` | `*` | No | Comma-separated allowed origins. Lock down in production, e.g. `https://app.example.com` |
| `REDIS_ADDR` | — | No | Redis address for rate limiting, e.g. `redis:6379`. Falls back to in-process counter if unset |
| `BIFROST_LOG_FORMAT` | `json` | No | `json` (structured, for log aggregation) or `console` (human-readable) |
| `BIFROST_METRICS` | `false` | No | Set to `true` to enable Prometheus `/metrics` endpoint |
| `BIFROST_TOKEN_TTL` | `24h` | No | Auth token lifetime. Use short values for ephemeral setups (e.g. `1h`) |

> **Generating secrets:**
> ```bash
> # Signing key (any length, base64-safe)
> openssl rand -base64 32
>
> # Encryption key (must be exactly 32 bytes)
> openssl rand -hex 16   # 32 hex chars = 16 bytes — use openssl rand -base64 32 | tr -d '\n=' | head -c 32
> python3 -c "import secrets; print(secrets.token_hex(16))"
> ```

---

## Docker Compose

### Quick start

Copy `.env.example` to `.env` and fill in the required values:

```bash
cp .env.example .env
```

Minimum `.env` for Docker Compose:

```dotenv
BIFROST_SIGNING_KEY=<output of openssl rand -base64 32>
BIFROST_ENCRYPTION_KEY=<32-char string>
ADMIN_EMAIL=admin@example.com
```

Start the stack:

```bash
docker compose up -d
```

Boot order: `postgres` (healthy) → `migrate` (schema) → `bifrost` + `setup-job` (parallel).

The `migrate` service runs `bifrost-server --migrate-only`, applies all GORM migrations, and exits. The `setup-job` runs `bifrost-cli init-admin` to create the first admin user and prints the bearer token to stdout.

Retrieve the admin token after first boot:

```bash
docker compose logs setup-job
```

### Customising the setup job

The `setup-job` uses environment variables for the admin account:

| Variable | Default | Description |
|----------|---------|-------------|
| `ADMIN_EMAIL` | _(required)_ | Admin user email |
| `ADMIN_NAME` | `admin` | Admin user display name |
| `ADMIN_ORG` | `Default` | Organisation name |

### Re-running migrations after upgrade

```bash
docker compose run --rm migrate
```

---

## Endpoints That Must Not Be Publicly Exposed

The following endpoints must be placed behind a VPN, firewall rule, or reverse proxy allowlist — they must **never** be reachable from the public internet:

| Endpoint | Reason |
|----------|--------|
| `POST /v1/setup` | One-shot bootstrap — exposes unauthenticated admin creation |
| `GET /metrics` | Leaks internal counters and resource usage |
| `GET /docs/openapi.json` | Reveals full API surface area |
| `GET /docs/openapi.yaml` | Same as above |

The proxy endpoint (`/v1/proxy/*`) and health check (`/healthz`) are safe to expose publicly.

---

## Reverse Proxy Configuration

Run Bifrost behind a reverse proxy to handle TLS termination and restrict sensitive endpoints.

### nginx

```nginx
server {
    listen 443 ssl;
    server_name api.example.com;

    ssl_certificate     /etc/ssl/certs/api.example.com.crt;
    ssl_certificate_key /etc/ssl/private/api.example.com.key;

    # Block sensitive management endpoints from the public internet
    location ~ ^/(v1/setup|metrics|docs) {
        deny all;
        return 403;
    }

    location / {
        proxy_pass         http://127.0.0.1:3333;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }
}
```

If you need management endpoints accessible internally only, serve them on a second `server` block bound to the internal interface.

### Caddy

```caddyfile
api.example.com {
    # Block sensitive endpoints
    @blocked path /v1/setup /metrics /docs/*
    respond @blocked 403

    reverse_proxy localhost:3333
}
```

For internal-only management access, add a second site bound to the internal address:

```caddyfile
internal.example.com {
    reverse_proxy localhost:3333
}
```

---

## Firewall Recommendations

- **Expose port 3333 (or your configured port) only via the reverse proxy** — do not bind Bifrost directly to a public interface.
- **Redis (6379)** — bind to `127.0.0.1` or the internal Docker network only. Never expose externally.
- **PostgreSQL (5432)** — same as Redis. Use `sslmode=require` in the DSN for any non-loopback connection.
- **Metrics port** — if you expose `/metrics` to a Prometheus scraper, use network-level ACLs or a scrape secret to prevent public access.

---

## Security Checklist

Run through this before exposing Bifrost to the internet:

- [ ] `BIFROST_ENCRYPTION_KEY` is set to a random 32-byte value — root keys are encrypted at rest
- [ ] `BIFROST_SIGNING_KEY` is set to a random value and not shared between environments
- [ ] `BIFROST_MODE` is **not** set to `test`
- [ ] `BIFROST_CORS_ORIGINS` is set to specific origins (not `*`)
- [ ] `/v1/setup` is blocked at the reverse proxy or firewall
- [ ] `/metrics` is blocked or restricted to internal scrapers
- [ ] `/docs/openapi.*` is blocked or restricted to internal consumers
- [ ] PostgreSQL is not reachable from the public internet; DSN uses `sslmode=require`
- [ ] Redis is not reachable from the public internet
- [ ] `.env` file is gitignored and not committed
- [ ] Container images are pulled from `ghcr.io/farovictor/bifrost` (signed releases) or built from a pinned commit

---

## Production Bootstrap Flow

```
1. docker compose up -d postgres
2. docker compose run --rm migrate        ← applies schema
3. docker compose up -d bifrost
4. docker compose run --rm setup-job      ← creates admin, prints token
5. Save the token — it will not be shown again
6. docker compose up -d                   ← start everything
```

After the first boot, `POST /v1/setup` returns `409 Conflict` and can be safely left in place (it is a no-op once an admin exists). Blocking it at the proxy layer is still recommended.
