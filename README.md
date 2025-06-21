# Bifrost – Secure, Delegated API Access for Cloud-Native Environments

Bifrost is a lightweight, extensible API proxy written in Go that enables secure delegation of API access through virtual keys. Instead of exposing long-lived secrets or API tokens to clients, Bifrost maps short-lived, scoped virtual keys to real credentials stored securely in Vault or other backends — and transparently proxies the request to the target API.

Built with Kubernetes in mind, Bifrost is designed to operate as a standalone proxy or as a Kubernetes Operator, making it easy to provision and manage virtual keys in cloud-native environments.

## Requirements

The project targets **Go 1.23.8**. To set up the required toolchain and run Bifrost locally, execute:

```bash
make setup
make run
```

If you want to exercise the rate‑limit middleware, ensure Redis is running (for example via `docker run -d --name redis-dev -p 6379:6379 redis:7-alpine`).

## Running Tests
Run the suite with:
```bash
go test ./...
```

### Configuration via Environment Variables

Bifrost can be configured through a handful of environment variables:

- `BIFROST_PORT` – HTTP port to bind to (defaults to `3333`).
- `REDIS_ADDR` – address of the Redis instance (defaults to `localhost:6379`).
- `REDIS_PASSWORD` – password for Redis, if required.
- `REDIS_DB` – numeric Redis DB index to use (defaults to `0`).
- `REDIS_PROTOCOL` – Redis protocol version (defaults to `3`).

You can export these variables or prefix them when starting the server.

#### Example

```bash
BIFROST_PORT=8080 REDIS_ADDR=localhost:6379 make run
```

# Core Features
## Virtual Key Mapping
Define ephemeral, revocable keys mapped to long-lived secrets or tokens.
## Secure Credential Injection
Inject real credentials into proxied requests without exposing them to the client.
## Policy Enforcement & Scoping
Apply granular access policies per virtual key: rate limits, expiration, scope control.
## Vault Integration
Retrieve and manage secrets securely with native Vault support.
## Kubernetes Native
Deploy Bifrost as a Kubernetes operator with CRD support for virtual key management.
## Audit & Observability
Log, trace, and monitor access by key, user, origin, or service.
## Golang First
Fast, type-safe, and built for performance and extensibility.

## CLI Usage
You can manage virtual keys from the command line with the `bifrost` tool.
The commands interact with the running HTTP API by default.

```bash
# issue a new key
go run ./cmd/bifrost issue --id mykey --target svc --scope read --ttl 10m

# revoke an existing key
go run ./cmd/bifrost revoke mykey
```

Use `--addr` to specify a custom API address if the server is not running on
`http://localhost:3333`.


# Planned Extensions
- Integration with Open Policy Agent (OPA) for dynamic authorization.
- Support for multiple target backends (OpenAI, Stripe, internal APIs).
- Web-based management dashboard for virtual keys.
- Optional JWT issuance or pass-through with verification hooks.

