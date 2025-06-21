# Test Suite

This directory contains integration tests for the HTTP routes provided by the Bifrost server. The tests use Go's `net/http/httptest` package to start the router and validate each endpoint's behavior.

## Running

Execute all tests from the repository root using the standard Go tooling:

```bash
go test ./...
```

The command will build and run every `*_test.go` file, including the route tests found here.
