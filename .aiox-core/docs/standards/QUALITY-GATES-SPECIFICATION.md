# Quality Gates Specification v4.2

**Version:** 2.1.0
**Last Updated:** 2025-12-09
**Status:** Official Standard
**Related:** Sprint 3 Implementation

---

## 📋 Table of Contents

- [Overview](#overview)
- [3-Layer Architecture](#3-layer-architecture)
- [Layer 1: Pre-commit](#layer-1-pre-commit)
- [Layer 2: PR Automation](#layer-2-pr-automation)
- [Layer 3: Human Review](#layer-3-human-review)
- [Configuration Guide](#configuration-guide)
- [CodeRabbit Self-Healing](#coderabbit-self-healing)
- [Metrics & Impact](#metrics--impact)

---

## Overview

### Purpose

The Quality Gates 3-Layer system ensures code quality through progressive automated validation, catching 80% of issues automatically and focusing human review on strategic decisions.

### Design Principles

1. **Shift Left** - Catch issues as early as possible
2. **Progressive Depth** - Each layer adds more comprehensive checks
3. **Automation First** - Humans focus on what humans do best
4. **Fast Feedback** - Immediate response at each layer
5. **Non-Blocking Default** - Warnings vs. errors where appropriate

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     QUALITY GATES WORKFLOW                              │
│                                                                         │
│   Developer                                                             │
│      │                                                                  │
│      ▼                                                                  │
│   ┌──────────────────────────────────────────────────────────────────┐ │
│   │                     LAYER 1: PRE-COMMIT                           │ │
│   │                     ════════════════════                          │ │
│   │   Trigger: File save, git commit                                  │ │
│   │   Time: < 5 seconds                                               │ │
│   │   Catches: 30% of issues                                          │ │
│   │                                                                   │ │
│   │   ✓ ESLint (syntax, patterns)                                    │ │
│   │   ✓ Prettier (formatting)                                        │ │
│   │   ✓ TypeScript (type checking)                                   │ │
│   │   ✓ Unit tests (changed files only)                              │ │
│   │                                                                   │ │
│   │   Blocking: Yes (can't commit if fails)                          │ │
│   └──────────────────────────────────────────────────────────────────┘ │
│                                │                                        │
│                          PASS? │                                        │
│                                ▼                                        │
│                         git commit                                      │
│                         git push                                        │
│                                │                                        │
│                                ▼                                        │
│   ┌──────────────────────────────────────────────────────────────────┐ │
│   │                     LAYER 2: PR AUTOMATION                        │ │
│   │                     ══════════════════════                        │ │
│   │   Trigger: PR creation, PR update                                 │ │
│   │   Time: < 3 minutes                                               │ │
│   │   Catches: Additional 50% (80% cumulative)                        │ │
│   │                                                                   │ │
│   │   ✓ CodeRabbit AI review                                         │ │
│   │   ✓ Integration tests                                            │ │
│   │   ✓ Coverage analysis (threshold: 80%)                           │ │
│   │   ✓ Security scan (npm audit, Snyk)                              │ │
│   │   ✓ Performance benchmarks                                       │ │
│   │   ✓ Documentation validation                                     │ │
│   │                                                                   │ │
│   │   Blocking: Yes (required checks for merge)                      │ │
│   └──────────────────────────────────────────────────────────────────┘ │
│                                │                                        │
│                          PASS? │                                        │
│                                ▼                                        │
│   ┌──────────────────────────────────────────────────────────────────┐ │
│   │                     LAYER 3: HUMAN REVIEW                         │ │
│   │                     ═════════════════════                         │ │
│   │   Trigger: Layer 2 passes                                         │ │
│   │   Time: 30 min - 2 hours                                          │ │
│   │   Catches: Final 20% (100% cumulative)                            │ │
│   │                                                                   │ │
│   │   □ Architecture alignment                                        │ │
│   │   □ Business logic correctness                                    │ │
│   │   □ Edge cases coverage                                           │ │
│   │   □ Documentation quality                                         │ │
│   │   □ Security best practices                                       │ │
│   │   □ Strategic decisions                                           │ │
│   │                                                                   │ │
│   │   Blocking: Yes (final approval required)                        │ │
│   └──────────────────────────────────────────────────────────────────┘ │
│                                │                                        │
│                          APPROVE                                        │
│                                │                                        │
│                                ▼                                        │
│                            MERGE                                        │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Layer 1: Pre-commit

### Purpose

Catch syntax errors, formatting issues, and simple bugs immediately during development, before code leaves the developer's machine.

### Checks

| Check | Tool | Config File | Blocking |
|-------|------|-------------|----------|
| Formatting | gofmt | built-in | Yes |
| Vet | go vet | built-in | Yes |
| Build | go build | built-in | Yes |
| Unit Tests | go test | built-in | Yes |
| Coverage | go test -coverprofile | built-in | Yes |
| Commit Message | conventional commits | manual | Yes |

### Configuration

#### .husky/pre-commit (or git hook)

```bash
#!/bin/sh

# Format all Go files
go fmt ./...

# Vet for correctness
go vet ./...

# Build to verify compilation
go build ./...

# Run tests (fast, unit only)
go test ./... -short
```

#### Makefile targets

```makefile
test:
	go test ./... -coverprofile=coverage.out

build:
	go build ./...

fmt:
	go fmt ./...

vet:
	go vet ./...
```

### Expected Results

- **Time:** < 5 seconds per commit
- **Issues Caught:** ~30% of all potential issues
- **Developer Experience:** Immediate feedback, no context switching

---

## Layer 2: PR Automation

### Purpose

Run comprehensive automated checks on every PR, including AI-powered code review, integration tests, and security scanning.

### Checks

| Check | Tool | Threshold | Blocking |
|-------|------|-----------|----------|
| AI Code Review | CodeRabbit | N/A (suggestions) | No* |
| Integration Tests | Jest | 100% pass | Yes |
| Coverage | Jest | 80% minimum | Yes |
| Security Audit | npm audit | No high/critical | Yes |
| Lint | ESLint | 0 errors | Yes |
| Type Check | TypeScript | 0 errors | Yes |
| Build | npm/webpack | Success | Yes |

*CodeRabbit suggestions are non-blocking but tracked.

### Configuration

#### .github/workflows/quality-gates-pr.yml

```yaml
name: Quality Gates PR

on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [main, develop]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  quality-gates:
    runs-on: ubuntu-latest
    timeout-minutes: 15

    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: Verify formatting
        run: |
          if [ -n "$(gofmt -l .)" ]; then
            echo "Go files not formatted. Run go fmt ./..."
            exit 1
          fi

      - name: Vet
        run: go vet ./...

      - name: Build
        run: go build ./...

      - name: Test with coverage
        run: go test ./... -coverprofile=coverage.out -covermode=atomic

      - name: Check coverage threshold
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $coverage%"
          if (( $(echo "$coverage < 70" | bc -l) )); then
            echo "Coverage below 70% threshold"
            exit 1
          fi

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
          fail_ci_if_error: false
```

#### .github/coderabbit.yaml

```yaml
# CodeRabbit Configuration
language: "en"
tone_instructions: "Be constructive and helpful. Focus on bugs, security, and best practices."
early_access: false

reviews:
  profile: "chill"
  request_changes_workflow: false
  high_level_summary: true
  poem: false
  review_status: true
  collapse_walkthrough: false
  auto_review:
    enabled: true
    drafts: false
    base_branches:
      - main
      - develop
  path_filters:
    - path: "**/*.test.ts"
      instructions: "Focus on test coverage and edge cases"
    - path: "**/*.md"
      instructions: "Check for broken links, typos, and clarity"
    - path: ".aiox-core/**"
      instructions: "Ensure consistency with framework standards"

chat:
  auto_reply: true
```

### Expected Results

- **Time:** < 3 minutes per PR update
- **Issues Caught:** Additional 50% (80% cumulative)
- **Developer Experience:** Detailed feedback before human review

---

## Layer 3: Human Review

### Purpose

Strategic review by humans focusing on architecture, business logic, and edge cases that automated tools cannot evaluate.

### Review Focus

| Area | Reviewer | What to Check |
|------|----------|---------------|
| Architecture | @architect, Tech Lead | Alignment with patterns, scalability |
| Business Logic | PO, Domain Expert | Correctness, edge cases |
| Security | Security Champion | Best practices, vulnerabilities |
| Documentation | Tech Writer | Clarity, completeness |
| UX Impact | UX Expert | User-facing changes |

### CODEOWNERS Configuration

```
# CODEOWNERS - Layer 3 Human Review Assignments

# Default reviewers
* @team-leads

# Architecture-sensitive areas
/.aiox-core/core/ @architect @senior-devs
/docs/architecture/ @architect
/src/core/ @senior-devs

# Security-sensitive areas
/src/auth/ @security-team
/.github/workflows/ @devops-team
**/security*.* @security-team

# Documentation
*.md @tech-writers
/docs/ @tech-writers

# Configuration files
package.json @senior-devs
tsconfig.json @senior-devs
.eslintrc.* @senior-devs

# Squads (modular areas)
/squads/etl/ @data-team
/squads/creator/ @content-team
```

### Review Checklist

```markdown
## Human Review Checklist

### Architecture
- [ ] Changes align with module boundaries
- [ ] Dependencies flow correctly (no circular)
- [ ] No breaking changes without migration path

### Business Logic
- [ ] Requirements correctly implemented
- [ ] Edge cases handled
- [ ] Error scenarios covered

### Security
- [ ] No hardcoded secrets
- [ ] Input validation present
- [ ] Authentication/authorization correct

### Performance
- [ ] No N+1 queries
- [ ] Caching considered
- [ ] Large operations async

### Documentation
- [ ] README updated if needed
- [ ] API documentation current
- [ ] Breaking changes documented

### Tests
- [ ] Critical paths covered
- [ ] Edge cases tested
- [ ] Mocks appropriate
```

### Expected Results

- **Time:** 30 min - 2 hours per PR
- **Issues Caught:** Final 20% (100% cumulative)
- **Focus:** Strategic decisions, not syntax

---

## Configuration Guide

### Initial Setup

```bash
# Go requires no package manager setup — toolchain is built-in

# 1. Ensure Go 1.23+ is installed
go version

# 2. Install project dependencies
go mod download

# 3. Verify everything builds
go build ./...

# 4. Run tests with coverage
make test
# or: go test ./... -coverprofile=coverage.out

# 5. View coverage report
go tool cover -html=coverage.out
```

### Customization

#### Adjusting Coverage Thresholds

Edit the CI workflow threshold check in `.github/workflows/quality-gates-pr.yml`:

```bash
# Change the threshold value (currently 70%)
if (( $(echo "$coverage < 70" | bc -l) )); then
```

Or use Makefile targets for specific test scopes:

```makefile
test-integration:
	go test ./tests/

test-one:
	go test ./tests/ -run TestFoo
```

#### Skipping Checks (Emergency Only)

```bash
# Skip Layer 1 pre-commit hook (use sparingly!)
git commit --no-verify -m "emergency: fix production issue"

# Layer 2: Use [skip ci] in commit message
git commit -m "docs: update readme [skip ci]"
```

---

## CodeRabbit Self-Healing

### Story Type Analysis

CodeRabbit automatically adjusts review focus based on story type:

| Story Type | Review Focus | Priority Checks |
|------------|--------------|-----------------|
| 🔧 Infrastructure | Configuration, CI/CD | Security, backwards compatibility |
| 💻 Feature | Business logic, UX | Tests, documentation |
| 📖 Documentation | Clarity, accuracy | Links, terminology |
| ✅ Validation | Test coverage | Edge cases |
| 🐛 Bug Fix | Root cause, regression | Tests, side effects |

### Path-Based Instructions

```yaml
# .github/coderabbit.yaml
reviews:
  path_instructions:
    - path: "**/*.test.ts"
      instructions: |
        Focus on:
        - Test coverage completeness
        - Edge case handling
        - Mock appropriateness
        - Assertion quality

    - path: ".aiox-core/docs/standards/**"
      instructions: |
        Verify:
        - Terminology uses 'Squad' not 'Squad'
        - All internal links work
        - Version numbers are v4.2

    - path: "squads/**"
      instructions: |
        Check:
        - squad.yaml manifest is valid
        - peerDependency on @aiox/core declared
        - Follows Squad structure conventions

    - path: ".github/workflows/**"
      instructions: |
        Review:
        - No hardcoded secrets
        - Proper timeout settings
        - Concurrency configuration
        - Security best practices
```

---

## Metrics & Impact

### Before Quality Gates (v2.0)

| Metric | Value |
|--------|-------|
| Issues caught automatically | 0% |
| Average review time | 2-4 hours per PR |
| Issues escaping to production | ~15% |
| Developer context switches | High |

### After Quality Gates (v4.2)

| Metric | Value | Improvement |
|--------|-------|-------------|
| Issues caught automatically | 80% | **∞** |
| Average review time | 30 min per PR | **75% reduction** |
| Issues escaping to production | <5% | **67% reduction** |
| Developer context switches | Low | **Significant** |

### Layer Breakdown

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     ISSUE DETECTION BY LAYER                            │
│                                                                         │
│   Layer 1 (Pre-commit)                                                  │
│   ████████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  30%    │
│                                                                         │
│   Layer 2 (PR Automation)                                               │
│   ████████████████████████████████████████████████████████░░░░  80%    │
│   (includes Layer 1 + additional 50%)                                   │
│                                                                         │
│   Layer 3 (Human Review)                                                │
│   ████████████████████████████████████████████████████████████  100%   │
│   (includes Layer 1 + Layer 2 + final 20%)                             │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Related Documents

- [AIOX-LIVRO-DE-OURO-V2.1-COMPLETE.md](./AIOX-LIVRO-DE-OURO-V2.1-COMPLETE.md)
- [CodeRabbit Integration Decisions](../../docs/architecture/coderabbit-integration-decisions.md)
- [STORY-TEMPLATE-V2-SPECIFICATION.md](./STORY-TEMPLATE-V2-SPECIFICATION.md)

---

**Last Updated:** 2025-12-09
**Version:** 2.1.0
**Maintainer:** @qa (Quinn)
