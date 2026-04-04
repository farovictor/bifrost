# Contributor Guidelines for Agents

Automated agents working in this repository must run the following
checks before committing:

```bash
go fmt ./...
go test ./...
```

Only commit if both commands succeed.

Commit messages should be concise and written in the present tense.
Pull request descriptions must summarize the changes and mention test
results.

For local development you can use the provided Makefile:

```bash
make setup
make run
```

---

<!-- AIOX-MANAGED SECTIONS -->
<!-- These sections are managed by AIOX. Edit content between markers carefully. -->
<!-- Your custom content above will be preserved during updates. -->

<!-- AIOX-MANAGED-START: core -->
## Core Rules

1. Follow the Constitution at `.aiox-core/constitution.md`
2. Prioritize `CLI First -> Observability Second -> UI Third`
3. Work from stories in `docs/stories/`
4. Do not invent requirements outside existing artifacts
<!-- AIOX-MANAGED-END: core -->

<!-- AIOX-MANAGED-START: quality -->
## Quality Gates

- Run `go fmt ./...`
- Run `go build ./...`
- Run `go test ./...`
- Run `make test` (full test suite with coverage)
- Update story checklist and file list before completing
<!-- AIOX-MANAGED-END: quality -->

<!-- AIOX-MANAGED-START: codebase -->
## Project Map

- Core framework: `.aiox-core/`
- HTTP handlers: `routes/`
- Proxy handler: `routes/v1/`
- Middlewares: `middlewares/`
- Domain packages: `pkg/` (keys, rootkeys, services, orgs, users, auth, database)
- CLI entrypoint: `cmd/bifrost/`
- Integration tests: `tests/`
- API docs: `doc/`
- Swagger spec: `docs/swagger/`
<!-- AIOX-MANAGED-END: codebase -->

<!-- AIOX-MANAGED-START: commands -->
## Common Commands

- `make run` — start server with in-memory stores, console logs
- `make test` — run tests with coverage (`go test ./... -coverprofile=coverage.out`)
- `make swagger` — regenerate docs/swagger/ from handler annotations
- `go build ./...` — verify compilation
- `go fmt ./...` — format all Go source files
- `go test ./tests/` — run integration tests only
<!-- AIOX-MANAGED-END: commands -->

<!-- AIOX-MANAGED-START: shortcuts -->
## Agent Shortcuts

Activation preference in Codex CLI:
1. Use `/skills` and select `aiox-<agent-id>` from `.codex/skills` (e.g., `aiox-architect`)
2. Alternatively, use the shortcuts below (`@architect`, `/architect`, etc.)

Interpret the shortcuts below by loading the corresponding file in `.aiox-core/development/agents/` (fallback: `.codex/agents/`), render the greeting via `generate-greeting.js` and assume the persona until `*exit`:

- `@architect`, `/architect`, `/architect.md` -> `.aiox-core/development/agents/architect.md`
- `@dev`, `/dev`, `/dev.md` -> `.aiox-core/development/agents/dev.md`
- `@qa`, `/qa`, `/qa.md` -> `.aiox-core/development/agents/qa.md`
- `@pm`, `/pm`, `/pm.md` -> `.aiox-core/development/agents/pm.md`
- `@po`, `/po`, `/po.md` -> `.aiox-core/development/agents/po.md`
- `@sm`, `/sm`, `/sm.md` -> `.aiox-core/development/agents/sm.md`
- `@analyst`, `/analyst`, `/analyst.md` -> `.aiox-core/development/agents/analyst.md`
- `@devops`, `/devops`, `/devops.md` -> `.aiox-core/development/agents/devops.md`
- `@data-engineer`, `/data-engineer`, `/data-engineer.md` -> `.aiox-core/development/agents/data-engineer.md`
- `@ux-design-expert`, `/ux-design-expert`, `/ux-design-expert.md` -> `.aiox-core/development/agents/ux-design-expert.md`
- `@squad-creator`, `/squad-creator`, `/squad-creator.md` -> `.aiox-core/development/agents/squad-creator.md`
- `@aiox-master`, `/aiox-master`, `/aiox-master.md` -> `.aiox-core/development/agents/aiox-master.md`
<!-- AIOX-MANAGED-END: shortcuts -->
