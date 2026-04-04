# Project Brief: Bifrost

## Executive Summary

Bifrost is a proxy service that sits between users (or agents) and third-party APIs, issuing short-lived virtual keys instead of exposing real API credentials. When a request is made with a virtual key, Bifrost authenticates it, enforces access policies, and forwards the request with the real key injected — keeping actual credentials hidden at all times.

The core problem it solves is **shared API key management**: today, teams share raw API keys across services and employees, creating security gaps, no visibility into usage, and no way to revoke individual access without rotating the key for everyone. Bifrost gives each user or service account its own virtual key that can be managed, rate-limited, and revoked independently.

**Target users** include both individual developers (managing service accounts) and enterprises needing granular, auditable control over who uses which APIs, at what frequency, and at what cost — without touching cloud provider consoles.

**Key differentiator:** No existing solution makes API key proxying and lifecycle management this simple and accessible as a self-hostable service.

---

## Problem Statement

API keys are the de facto credential for accessing third-party services (OpenAI, Stripe, AWS, etc.), yet most teams manage them unsafely: keys are shared across employees and services, hardcoded in codebases, and rarely rotated. When a key is compromised or an employee leaves, the only option is to rotate the key — which breaks every service using it simultaneously.

The impact is significant:
- **No individual accountability** — shared keys make it impossible to audit who made which request
- **No granular revocation** — revoking one user's access means disrupting everyone
- **No usage visibility** — teams don't know which service is consuming how many tokens or dollars
- **Security exposure** — real credentials leak into logs, environment files, and agent prompts

Existing solutions (cloud IAM, API gateway products) are complex, expensive, or tightly coupled to a single cloud provider. There is no simple, provider-agnostic tool that makes per-user virtual key management accessible to small teams and enterprises alike.

---

## Proposed Solution

Bifrost is a self-hostable HTTP proxy that decouples API consumers from real API credentials by introducing a **virtual key layer**. Instead of distributing actual API keys, administrators issue short-lived virtual keys scoped to specific services, rate limits, and expiry windows. Requests made with a virtual key are authenticated by Bifrost, which injects the real credential before forwarding the request to the upstream provider.

**Core approach:**
- Each user, agent, or service gets its own virtual key — independently manageable and revocable
- Real API keys are stored server-side and never exposed to consumers
- All requests are proxied and logged, providing full usage visibility
- Access policies (rate limits, expiry, scope) are enforced per virtual key

**Key differentiators:**
- **Provider-agnostic** — works with any HTTP-based API (OpenAI, Stripe, custom services)
- **Self-hostable** — organizations keep full control over their data and credentials
- **Simple by design** — no cloud vendor lock-in, no complex IAM configuration
- **AI-era ready** — purpose-built for teams running agents and LLM workloads that need controlled API access

**Why this will succeed:** The rise of AI agents dramatically increases the surface area of API key exposure. Bifrost addresses a gap that cloud providers won't fill — lightweight, universal, self-hosted key proxying.

---

## Target Users

### Primary User Segment: Engineering Teams & Platform Engineers

- **Profile:** Small-to-mid engineering teams (5–50 engineers) building products that integrate multiple third-party APIs
- **Current behavior:** Manually managing API keys in `.env` files, secret managers, or shared Notion docs; rotating keys reactively after incidents
- **Pain points:** No visibility into per-user/service consumption, fear of key leakage in agent prompts or logs, painful key rotation cycles
- **Goals:** Secure, auditable API access without adding operational complexity

### Secondary User Segment: Enterprises with AI Agent Workloads

- **Profile:** Larger organizations deploying LLM agents (internal tools, customer-facing bots) that make frequent API calls
- **Current behavior:** Sharing a single OpenAI/Anthropic key across all agents, with no per-agent rate limiting or spend tracking
- **Pain points:** No granular control over agent API usage, compliance risk from unaudited credentials, inability to isolate a rogue agent without disrupting others
- **Goals:** Granular per-agent key lifecycle management, usage reporting for cost allocation, compliance-ready audit trails

---

## Goals & Success Metrics

### Business Objectives
- Establish Bifrost as the go-to open-source solution for API key proxying and lifecycle management
- Achieve adoption among teams running AI agent workloads as a first distribution channel
- Build a self-sustaining open-source community with contributors and enterprise interest within 12 months

### User Success Metrics
- Users can issue, revoke, and scope a virtual key in under 2 minutes
- Teams gain full per-key usage visibility without any changes to their existing API call code
- Zero real API key exposure incidents for teams using Bifrost in production

### Key Performance Indicators (KPIs)
- **GitHub stars:** 500 in 3 months, 2,000 in 12 months
- **Active deployments:** 50 self-hosted instances within 6 months
- **Key operations per day:** Average deployment managing 20+ virtual keys actively
- **Time-to-first-proxy-request:** < 10 minutes from `git clone` to first proxied call
- **Community:** 10+ external contributors within 12 months

---

## MVP Scope

### Core Features (Must Have)
- **Virtual key issuance:** Create virtual keys scoped to a specific upstream service, with configurable expiry and rate limits
- **Request proxying:** Authenticate incoming requests by virtual key, inject real credentials, forward to upstream API
- **Key revocation:** Instantly revoke a virtual key without affecting other keys or rotating the real credential
- **Multi-user/org support:** Organize keys under users and organizations with role-based access
- **Usage tracking:** Log request counts per virtual key for basic visibility
- **Management API:** REST endpoints to create, list, and delete keys, users, orgs, and services

### Out of Scope for MVP
- Management dashboard (UI) — API-first, CLI/curl usage for MVP
- Vault/secrets manager backend integration
- OPA (Open Policy Agent) policy enforcement
- Kubernetes operator
- Billing / spend tracking with cost attribution
- SSO / enterprise identity provider integration

### MVP Success Criteria

A team can self-host Bifrost, register a real API key for a service, issue virtual keys to individual users or agents, proxy requests through Bifrost, and revoke access for a specific key — all without ever exposing the real credential to consumers. The full flow works end-to-end in under 10 minutes from a fresh deployment.

---

## Post-MVP Vision

### Phase 2 Features
- **Management Dashboard:** Web UI for key lifecycle management, usage graphs, and org administration — making Bifrost accessible to non-technical users
- **Vault Backend:** Integration with HashiCorp Vault and cloud secret managers (AWS Secrets Manager, GCP Secret Manager) for enterprise-grade credential storage
- **Spend Tracking:** Cost attribution per virtual key based on token usage (critical for LLM workloads)
- **Webhook Alerts:** Notify on rate limit breaches, expiry, or unusual usage patterns
- **Protocol Expansion:** gRPC and WebSocket proxying to support streaming APIs and broader service coverage
- **Security Innovations:** Key fingerprinting (IP/user-agent binding), one-shot ephemeral keys for CI/CD, automated credential rotation workflows
- **AI Agent Native** ⭐ *(Priority)*: MCP server mode to expose Bifrost as a Model Context Protocol server for on-demand key issuance by agents; agent identity header injection for upstream auditability
- **Intelligence Layer:** Request inspection for prompt injection detection, token budget enforcement per key, semantic rate limiting

### Long-term Vision (1-2 years)
- Bifrost becomes the standard infrastructure layer for AI agent API access — every agent framework (LangChain, AutoGPT, CrewAI) recommends it as the default key management solution
- Enterprise SaaS offering: hosted Bifrost with SSO, compliance reports, and SLA guarantees
- OPA integration enabling policy-as-code for fine-grained access control
- Kubernetes operator for cloud-native deployments

### Expansion Opportunities
- **Cost marketplace:** Bifrost as a metering layer enabling internal API cost chargebacks across teams
- **Key broker:** Act as a credential broker for SaaS platforms that want to offer multi-tenant API access to their own customers
- **Compliance module:** Auto-generate SOC2/ISO27001 evidence from Bifrost audit logs
- **Observability** *(lower priority)*: Grafana dashboard template, anomaly detection webhooks, per-request trace IDs via `X-Bifrost-Trace-ID`

---

## Technical Considerations

### Platform Requirements
- **Target Platforms:** Linux server / Docker container (self-hosted); macOS supported for local dev
- **Go Version:** 1.23.8
- **Browser/OS Support:** API-only for MVP; no browser requirements
- **Performance Requirements:** Proxy overhead target < 10ms per request; Redis rate limiting for high-throughput deployments

### Technology Preferences
- **Language:** Go 1.23.8
- **HTTP Router:** `go-chi/chi v5` — lightweight, idiomatic, middleware-friendly
- **Database:** PostgreSQL via `gorm.io/driver/postgres` (pgx v5 driver) + SQLite via `gorm.io/driver/sqlite` for lightweight deployments
- **ORM:** GORM v1.30
- **Cache / Rate Limiting:** Redis via `redis/go-redis/v9`; falls back to in-memory when Redis is unavailable
- **CLI:** Cobra v1.9.1
- **Logging:** `rs/zerolog` — structured, zero-allocation JSON logging
- **Metrics:** `prometheus/client_golang v1.17` — `/metrics` endpoint built in
- **API Docs:** `swaggo/swag` — OpenAPI spec auto-generated from handler annotations
- **Auth:** HMAC-SHA256 token signing (`pkg/auth/`)
- **ID generation:** `google/uuid v1.6`

### Architecture Considerations
- **Repository Structure:** Monorepo — single Go module (`github.com/farovictor/bifrost`)
- **Service Architecture:** Monolith — single binary serving management API + proxy handler
- **Dual store pattern:** Every domain (`keys`, `rootkeys`, `services`, `orgs`, `users`) has a `MemoryStore` (tests/dev) and `SQLStore` (production) implementing the same interface
- **Integration:** Any HTTP upstream via reverse-proxy; credential injection is configurable per service
- **Security:** HMAC-SHA256 tokens, virtual key scoping, rate limiting per key; real credentials never leave the server
- **Observability:** Prometheus metrics + zerolog structured logs out of the box

---

## Constraints & Assumptions

### Constraints
- **Budget:** Open-source project — no dedicated budget; development driven by maintainer time and community contributions
- **Timeline:** MVP is already functionally complete (Phase 0-2 done); Phase 3 features are next
- **Resources:** Small team / solo maintainer; contributions welcome but not guaranteed
- **Technical:** Go monolith only for MVP — no polyglot services; PostgreSQL or SQLite required for production; Redis optional but recommended for rate limiting at scale

### Key Assumptions
- Users are comfortable self-hosting via Docker or direct binary deployment
- Target users have basic familiarity with REST APIs and HTTP proxying concepts
- Upstream APIs are standard HTTP/HTTPS — no proprietary protocols in MVP scope
- PostgreSQL is available in production environments; SQLite acceptable for small/dev deployments
- The AI agent ecosystem (LangChain, AutoGPT, CrewAI, etc.) will continue adopting MCP as a standard, validating the MCP server mode investment
- Open-source distribution is the primary adoption channel; no paid marketing budget

---

## Risks & Open Questions

### Key Risks
- **Security misconfiguration:** Bifrost holds real API credentials — a misconfigured deployment (exposed management API, weak root key) could lead to credential compromise. *Impact: critical; mitigation: secure defaults, deployment hardening docs, mandatory auth on all management endpoints*
- **Performance bottleneck:** Every API call routes through Bifrost — if it becomes a latency or availability bottleneck, teams will bypass it. *Impact: high; mitigation: benchmark proxy overhead, ensure < 10ms target, document HA deployment patterns*
- **MCP ecosystem bet:** AI agent native features depend on MCP becoming the dominant agent tool protocol. *Impact: medium; mitigation: keep MCP mode additive, not core — Bifrost works without it*
- **Adoption chicken-and-egg:** Open-source tools need community to grow but need growth to attract community. *Impact: medium; mitigation: target AI agent communities (LangChain Discord, Hacker News) where key management pain is acute*
- **Maintenance burden:** Solo/small team maintaining security-critical infrastructure is risky. *Impact: medium; mitigation: keep codebase minimal, well-tested (77.6% coverage already), and dependency-light*

### Open Questions
- Should Bifrost offer a hosted/SaaS option, or remain purely self-hosted?
- What is the licensing model — MIT, Apache 2.0, or source-available for enterprise features?
- How should credential storage be handled at rest — encrypt in DB or delegate entirely to Vault?
- Is there a target cloud marketplace listing (AWS/GCP/Azure) to accelerate enterprise adoption?
- Should the management dashboard be a separate service or bundled into the same binary?

### Areas Needing Further Research
- Competitive landscape: LiteLLM proxy, Portkey — deeper differentiation analysis
- MCP adoption rate and which agent frameworks are committing to it
- Enterprise compliance requirements (SOC2, GDPR) that self-hosters will ask about
- Benchmark data: proxy latency overhead under realistic load (100–1000 req/s)

---

## Competitive Landscape

| Competitor | Overlap | Bifrost Advantage |
|---|---|---|
| **Kong AI Gateway** | Credential injection, virtual keys, rate limiting | Zero licensing cost; single binary; first-class key TTL/scope; no Kubernetes required |
| **LiteLLM Proxy** | LLM proxy, multi-provider routing | Bifrost is provider-agnostic (not LLM-only); stronger key lifecycle management |
| **Portkey** | LLM observability, key management | Bifrost is fully self-hosted; no SaaS dependency; open-source |

**Kong AI Gateway deep dive:** Kong solves credential isolation at the $50K–$200K+/year end of the market for teams already running Kong infrastructure. Key features like token-based rate limiting, RBAC, and advanced analytics are Enterprise-only. Bifrost solves the same core problem for $0, in minutes, for any developer.

**Positioning statement:** Bifrost is the developer-first, self-hosted alternative to enterprise AI gateways — delivering credential isolation, per-key lifecycle management, and usage visibility without the infrastructure burden or licensing cost.

---

## Next Steps

### Immediate Actions
1. Share this Project Brief with collaborators/stakeholders for review
2. Research LiteLLM and Portkey in more depth to sharpen competitive positioning
3. Decide on licensing model (MIT vs Apache 2.0) before enterprise conversations begin
4. Proceed to PRD creation using this brief as input

### PM Handoff

This Project Brief provides the full context for Bifrost. Please start in **PRD Generation Mode**, review the brief thoroughly to work with the user to create the PRD section by section as the template indicates, asking for any necessary clarification or suggesting improvements.
