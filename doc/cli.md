# CLI Reference

All CLI commands are run via `go run ./cmd/bifrost` (or the compiled binary).
Use `--addr` to point to a non-default server address (default: `http://localhost:3333`).

## Commands

```bash
# Seed an admin user + org (run once against a fresh database)
go run ./cmd/bifrost init-admin

# Create a user (returns api_key and token)
go run ./cmd/bifrost user-add --name Alice --email alice@acme.com --org-name acme --role owner

# Register a root key
go run ./cmd/bifrost rootkey-add --id my-root --apikey sk-secret

# Update a root key
go run ./cmd/bifrost rootkey-update --id my-root --apikey sk-new-secret

# Register a service
go run ./cmd/bifrost service-add --id my-svc --endpoint https://api.example.com --rootkey my-root

# Issue a virtual key
go run ./cmd/bifrost issue --id vk-1 --target my-svc --scope read --ttl 10m --rate-limit 60

# Revoke a virtual key
go run ./cmd/bifrost revoke vk-1

# Delete a service
go run ./cmd/bifrost service-delete my-svc

# Delete a root key
go run ./cmd/bifrost rootkey-delete my-root

# Check server health
go run ./cmd/bifrost check
```

## End-to-End Example

```bash
# 1. Start the server
make run

# 2. Create a root key
go run ./cmd/bifrost rootkey-add --id demo-root --apikey my-secret-key

# 3. Register an upstream service (default X-API-Key injection)
go run ./cmd/bifrost service-add --id demo-svc --endpoint http://localhost:8081 --rootkey demo-root

# 4. Create a user and org (save the printed token)
go run ./cmd/bifrost user-add --name Admin --email admin@example.com --org-name demo --role owner
# → prints: api_key=<key>  token=<token>

# 5. Issue a virtual key (requires API key + token)
curl -X POST http://localhost:3333/v1/keys \
  -H "X-API-Key: <api_key>" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"id":"demo-key","target":"demo-svc","scope":"read","expires_at":"2027-01-01T00:00:00Z","rate_limit":60}'

# 6. Make a proxied request using the virtual key
curl -H "X-Virtual-Key: demo-key" http://localhost:3333/v1/proxy/hello
```
