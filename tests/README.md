# Test Suite

This directory contains integration tests for the HTTP routes provided by the Bifrost server. The tests use Go's `net/http/httptest` package to start the router and validate each endpoint's behavior.

## Running

Run the suite with the provided Makefile target, which configures test mode and
an in-memory SQLite database:

```bash
make test
```

This invokes `go test ./...` with `BIFROST_MODE=test`, `BIFROST_DB=sqlite`, and
`DATABASE_DSN=file::memory:?cache=shared` so that all packages share the same
ephemeral database. If you prefer to call `go test` directly, set these
environment variables yourself.

## Authentication Tokens

When the server runs with `BIFROST_MODE=test` or `BIFROST_DB=sqlite`, token
signatures are not verified and any bearer token is accepted. The server still
uses the static HMAC signing key `test-signing-key` when generating tokens so
tests can create and verify them deterministically.
