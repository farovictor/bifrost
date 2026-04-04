---
name: Heimdall — Bifrost companion UI
description: Heimdall is the canonical enterprise management UI for Bifrost, living in a separate private repo
type: project
---

Bifrost is API-only by design (headless). Heimdall is the officially supported management UI, distributed as a companion service.

**Why:** "Bifrost is the open-core; Heimdall is the enterprise layer." Bifrost ships as a Go binary with no frontend. Heimdall is a separate Next.js app that talks to Bifrost over HTTP.

**How to apply:** Never suggest adding UI to Bifrost. Any management dashboard work belongs in Heimdall. When PRD/architecture decisions touch UI, reference Heimdall as the UI layer.

**Heimdall stack:** Next.js 16 App Router, TypeScript strict, Tailwind v4, shadcn/ui, TanStack Query, Zustand, orval (typed client codegen from Bifrost's OpenAPI spec)

**Repo:** `/Workspace/Personal/heimdall` (private, separate from Bifrost)

**Integration:** Heimdall generates a typed HTTP client from Bifrost's `/docs/openapi.json` via orval. Configured via `NEXT_PUBLIC_BIFROST_URL` (default: `localhost:3333`). Decoupled by design — Docker Compose co-deployment is trivial but not yet wired up.
