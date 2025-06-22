# Bifrost – Secure, Delegated API Access for Cloud-Native Environments

Bifrost is a lightweight, extensible API proxy written in Go that enables secure delegation of API access through virtual keys. Instead of exposing long-lived secrets or API tokens to clients, Bifrost maps short-lived, scoped virtual keys to real credentials stored securely in Vault or other backends — and transparently proxies the request to the target API.

Built with Kubernetes in mind, Bifrost is designed to operate as a standalone proxy or as a Kubernetes Operator, making it easy to provision and manage virtual keys in cloud-native environments.

## Requirements

The project targets **Go 1.23.8**. To set up the required toolchain and run Bifrost locally, execute:

```bash
make setup
make run
```

## Running with Docker Compose

The repository provides a `docker-compose.yml` that starts Bifrost together with
Redis and PostgreSQL. Make sure both **Docker** and **Docker Compose** are installed, then run:

```bash
docker-compose up -d
```

This brings up the stack in the background on the default ports. When finished,
tear it down with:

```bash
docker-compose down
```

With the services running you can execute the Go tests against the live stack:

```bash
go test ./...
```

See `test-rate-limiting.sh` for a small script that exercises the rate limit
middleware against the Compose setup.

The Compose file sets `POSTGRES_DSN` for Bifrost to connect to the bundled
PostgreSQL service. If you run the server manually, configure the variable like
so:

```bash
export POSTGRES_DSN="postgres://bifrost:bifrost@localhost:5432/bifrost?sslmode=disable"
```

## Example Request

Once the server is running you can proxy a request with a virtual key:

```bash
curl -H "X-Virtual-Key: <key>" http://localhost:3333/v1/proxy/hello
```

You can also supply the key via the `key` query parameter instead of the header.

## Running Tests
Run the suite with:
```bash
go test ./...
```

### Configuration via Environment Variables

Bifrost can be configured through the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `BIFROST_PORT` | HTTP port to bind to | `3333` |
| `REDIS_ADDR` | address of the Redis instance | `localhost:6379` |
| `REDIS_PASSWORD` | password for Redis, if required | *(empty)* |
| `REDIS_DB` | numeric Redis DB index to use | `0` |
| `REDIS_PROTOCOL` | Redis protocol version | `3` |
| `POSTGRES_DSN` | connection string for PostgreSQL | *(empty)* |
| `BIFROST_LOG_LEVEL` | log level (`debug`, `info`, `warn`, `error`) | `info` |
| `BIFROST_LOG_FORMAT` | log output format (`json` or `console`) | `json` |
| `BIFROST_ENABLE_METRICS` | expose Prometheus metrics | `false` |
| `BIFROST_ADMIN_ID` | ID for the initial admin user | `admin` |
| `BIFROST_ADMIN_API_KEY` | API key for the admin user | random |

Use these variables to control the verbosity and choose between machine-readable JSON logs or a console-friendly format.

You can export these variables or prefix them when starting the server.

#### Example

```bash
BIFROST_PORT=8080 REDIS_ADDR=localhost:6379 make run
```

### Virtual Key Format

Virtual keys are represented by a small JSON object with the following fields:

- `id` – unique identifier of the virtual key.
- `scope` – allowed operations for the key (`read` or `write`).
- `target` – service identifier that the key grants access to.
- `expires_at` – RFC3339 timestamp when the key expires.
- `rate_limit` – maximum requests per minute permitted.

Example:

```json
{
  "id": "mykey",
  "scope": "read",
  "target": "svc",
  "expires_at": "2024-01-02T15:04:05Z",
  "rate_limit": 60
}
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
When `BIFROST_ENABLE_METRICS` is set, Bifrost exposes Prometheus metrics at
`/metrics`. Available metrics are:

- `request_total` – total HTTP requests processed.
- `request_duration_seconds` – histogram of request durations.
- `key_usage_total` – counter of virtual key usages.
## Golang First
Fast, type-safe, and built for performance and extensibility.

## Service Layer Diagram

```mermaid
flowchart TD
    Client --> Router
    Router --> LoggingMiddleware --> MetricsMiddleware --> AuthMiddleware --> OrgCtxMiddleware

    Router --> |"/v1/users"| UserStore
    Router --> |"/v1/keys"| KeyStore
    Router --> |"/v1/rootkeys"| RootKeyStore
    Router --> |"/v1/services"| ServiceStore

    Router --> |"/v1/proxy"| RateLimitMiddleware --> Proxy
    RateLimitMiddleware --> Redis
    Proxy --> KeyStore
    Proxy --> ServiceStore
    Proxy --> RootKeyStore
    Proxy --> UpstreamService

    OrgCtxMiddleware --> MembershipStore
    OrgCtxMiddleware --> OrgStore
    AuthMiddleware --> UserStore
    MetricsMiddleware --> Prometheus
```

## CLI Usage
You can manage virtual keys from the command line with the `bifrost` tool.
The commands interact with the running HTTP API by default. Before issuing
keys you must register a root key (or update it later) and then associate services with it using
`rootkey-add`, `rootkey-update` and `service-add`.

All management endpoints now require authentication. Create an API user first
using `bifrost user-add` (or the `/v1/users` HTTP route) to obtain an API key
**and an authentication token** for subsequent requests.

`user-add` accepts `--org-id` to join an existing organization or `--org-name`
to create a new one on the fly. Use `--role` to specify the membership role
(`owner`, `admin`, or `member`). When `--org-name` is supplied without
`--org-id`, the organization is created and the user is added with the chosen
role.

The `--target` flag on `issue` refers to the service ID provided when calling
`service-add`.
Use `--rate-limit` to specify the maximum requests per minute allowed for the issued key.

```bash
# register a root key
go run ./cmd/bifrost rootkey-add --id root --apikey SECRET

# update the root key without downtime
go run ./cmd/bifrost rootkey-update --id root --apikey NEWSECRET

# register a service
go run ./cmd/bifrost service-add --id svc --endpoint http://localhost:8081 --rootkey root

# issue a new key for that service
go run ./cmd/bifrost issue --id mykey --target svc --scope read --ttl 10m --rate-limit 60

# revoke an existing key
go run ./cmd/bifrost revoke mykey

# delete a service
go run ./cmd/bifrost service-delete svc

# delete the root key
go run ./cmd/bifrost rootkey-delete root

# create an API user and organization
go run ./cmd/bifrost user-add --id admin --org-name demo-org --role owner
```

Use `--addr` to specify a custom API address if the server is not running on
`http://localhost:3333`.

The generated API key must be provided in the `X-API-Key` or `Authorization`
header when calling any `/v1` endpoint. Along with the key, include the
**auth token** returned by `user-add` in the `Authorization` header as a Bearer
token. The token expires after 24 hours and is used to resolve organization
context for each request.

### End-to-End Example
Below is a minimal workflow showing how to register a service, create a user and
then proxy a request using the issued key.

```bash
# add a root key
go run ./cmd/bifrost rootkey-add --id demo-root --apikey SECRET

# add the upstream service
go run ./cmd/bifrost service-add --id demo --endpoint http://localhost:8081 --rootkey demo-root

# create a user and organization (prints API key and auth token)
go run ./cmd/bifrost user-add --id admin --org-name demo-org --role owner

# issue a virtual key using the auth token
curl -X POST http://localhost:3333/v1/keys \
  -H "X-API-Key: <api_key>" \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"id":"demo-key","target":"demo","ttl":"5m"}'

# make a proxied request
curl -H "X-Virtual-Key: demo-key" http://localhost:3333/v1/proxy/hello
```

## HTTP API
The CLI commands above are thin wrappers around a simple HTTP interface. The
server listens on `http://localhost:3333` by default.

All routes under `/v1` (except `/healthz` and `/version`) require a valid API
key supplied via the `X-API-Key` or `Authorization` header.

### Issue a key
POST `/v1/keys` with a JSON body describing the key. The response echoes the
same JSON and returns status `201 Created`.

Example request:

```bash
curl -X POST http://localhost:3333/v1/keys \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: <api_key>' \
  -d '{"id":"mykey","scope":"read","target":"svc","expires_at":"2024-01-02T15:04:05Z","rate_limit":60}'
```

Expected JSON format:

```json
{
  "id": "mykey",
  "scope": "read",
  "target": "svc",
  "expires_at": "2024-01-02T15:04:05Z",
  "rate_limit": 60
}
```

### Revoke a key
DELETE `/v1/keys/<id>` to revoke a previously issued key. A successful request
returns status `204 No Content`.

```bash
curl -X DELETE http://localhost:3333/v1/keys/mykey \
  -H 'X-API-Key: <api_key>'
```

### Add a service
POST `/v1/services` with JSON describing the service. Returns `201 Created`.

Example:

```bash
curl -X POST http://localhost:3333/v1/services \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: <api_key>' \
  -d '{"id":"svc","endpoint":"http://localhost:8081","root_key_id":"root"}'
```

### Delete a service
DELETE `/v1/services/<id>` to remove a service. Include your API key header.

### Create a user
POST `/v1/users` with a JSON body containing the user ID. Optional `org_id` and
`org_name` fields associate the user with an organization. Provide a `role`
(`owner`, `admin`, or `member`) when joining or creating an organization. The
response returns the API key **and** an authentication token valid for 24 hours.

Example:

```bash
curl -X POST http://localhost:3333/v1/users \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: <api_key>' \
  -H 'Authorization: Bearer <token>' \
  -d '{"id":"admin","org_name":"demo-org","role":"owner"}'
```

### Metrics endpoint
When metrics are enabled, GET `/metrics` returns Prometheus-formatted metrics.

### Proxy using a virtual key
Send any request to `/v1/proxy/<path>` with either the `X-Virtual-Key` header or
the `key` query parameter containing a previously issued key. Bifrost injects
the real API key, strips the virtual key from the forwarded request and forwards
the request to the mapped service.

## Organizations and Memberships

Phase 1 introduces database-ready models for multi-tenant setups. An
**organization** groups related services and users, while a **membership** links
a user to an organization with a specific role. API endpoints for managing these
objects are not available yet but SQL migrations and in-memory stores are in
place.

### Roles

- **Owner** – full control over the organization and its members.
- **Admin** – manage resources but cannot remove or demote owners.
- **Member** – limited to resources explicitly granted.

The auth token contains the `org_id` used to look up the user's membership.
If a membership exists, the corresponding role is attached to the request
context so handlers can enforce organization-specific policies.


# Planned Extensions
- Integration with Open Policy Agent (OPA) for dynamic authorization.
- Support for multiple target backends (OpenAI, Stripe, internal APIs).
- Web-based management dashboard for virtual keys.
- Optional JWT issuance or pass-through with verification hooks.

