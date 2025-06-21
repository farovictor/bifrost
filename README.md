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


# Planned Extensions
- Integration with Open Policy Agent (OPA) for dynamic authorization.
- Support for multiple target backends (OpenAI, Stripe, internal APIs).
- Web-based management dashboard for virtual keys.
- Optional JWT issuance or pass-through with verification hooks.

