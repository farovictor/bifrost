# Test Suite

This directory contains integration tests for the HTTP routes provided by the Bifrost server. The tests use Go's `net/http/httptest` package to start the router and validate each endpoint's behavior.

## Running

Run the suite with the provided Makefile target, which configures test mode and
an in-memory SQLite database:

```bash
make test
```

This invokes `go test ./...` with `BIFROST_MODE=test`, `BIFROST_DB=sqlite`, and
`POSTGRES_DSN=file::memory:?cache=shared` so that all packages share the same
ephemeral database. If you prefer to call `go test` directly, set these
environment variables yourself.

## Authentication Tokens

When the server runs with `BIFROST_MODE=test` or `BIFROST_DB=sqlite`, it uses a
static HMAC signing key of `test-signing-key`. Tests can rely on this value to
create and verify authentication tokens.
