# Bifrost – Secure, Delegated API Access for Cloud-Native Environments

Bifrost is a lightweight, extensible API proxy written in Go that enables secure delegation of API access through virtual keys. Instead of exposing long-lived secrets or API tokens to clients, Bifrost maps short-lived, scoped virtual keys to real credentials stored in root keys — and transparently proxies the request to the target upstream service.

## Requirements

Go 1.23.8. To install the toolchain and dependencies:

```bash
make setup
```

## Quick Start

```bash
make run          # starts with in-memory stores, console log format
```

See [doc/configuration.md](doc/configuration.md) for database and environment options, including PostgreSQL and SQLite backends, and Docker Compose setup.

## Running tests

```bash
go test ./...
```

Tests use in-memory stores and require no external services. Generate a coverage report:

```bash
make test   # runs tests and writes coverage.out
go tool cover -html=coverage.out
```

## Documentation

| Document | Description |
|---|---|
| [doc/api.md](doc/api.md) | Full HTTP API reference — endpoints, auth, request/response formats |
| [doc/cli.md](doc/cli.md) | CLI commands and end-to-end usage example |
| [doc/configuration.md](doc/configuration.md) | Environment variables, database backends, Docker Compose |
| [doc/architecture.md](doc/architecture.md) | Code architecture, request flow, store patterns, roles |
