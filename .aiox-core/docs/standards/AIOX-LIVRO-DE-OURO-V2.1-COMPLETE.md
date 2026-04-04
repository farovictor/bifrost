# AIOX Framework - Golden Book v4.2 (Complete)

## The Definitive Operating System for AI Agent Orchestration

**Version:** 2.1.0
**Status:** Living Document
**Last Updated:** 2025-12-09
**Maintained By:** AIOX Framework Team + Community
**Main Repository:** `SynkraAI/aiox-core`

---

> **"Structure is Sacred. Tone is Flexible."**
> _— Philosophical foundation of AIOX_

---

## 📣 IMPORTANT: About This Document

This document is the **consolidated v4.2 version** that incorporates all changes from Sprints 2-5:

- ✅ **Modular Architecture** (4 modules: core, development, product, infrastructure)
- ✅ **Squad System** (new terminology, replacing "Squad")
- ✅ **Multi-Repo Strategy** (3 public + 2 private repositories)
- ✅ **Quality Gates 3 Layers** (Pre-commit, PR Automation, Human Review)
- ✅ **Story Template v2.0** (Cross-Story Decisions, CodeRabbit Integration)
- ✅ **npm Package Scoping** (@aiox/core, @aiox/squad-\*, @aiox/mcp-presets)

**Legacy References:**

- `AIOX-LIVRO-DE-OURO.md` - Base v2.0.0 (Jan 2025)
- `AIOX-LIVRO-DE-OURO-V2.1.md` - Partial delta
- `AIOX-LIVRO-DE-OURO-V2.1-SUMMARY.md` - Change summary

---

## 📜 Open Source vs. Service - Business Model v4.2

### What Changed from v2.0 to v4.0.4

**IMPORTANT: v4.0.4 fundamentally changed the business model!**

| Component                | v2.0        | v4.0.4          | Rationale                  |
| ------------------------ | ----------- | --------------- | -------------------------- |
| **11 Agents**            | ✅ Open     | ✅ Open         | Core functionality         |
| **Workers (97+)**        | ❌ Closed   | ✅ **OPEN**     | Commodity, network effects |
| **Service Discovery**    | ❌ None     | ✅ **BUILT-IN** | Community needs it         |
| **Task-First Arch**      | ⚠️ Implicit | ✅ **EXPLICIT** | Architecture clarity       |
| **Clones (DNA Mental™)** | 🔒 Closed   | 🔒 **CLOSED**   | True moat (IP)             |
| **Squads**               | 🔒 Closed   | 🔒 **CLOSED**   | Domain expertise           |

### Multi-Repo Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     SYNKRA ORGANIZATION                                 │
│                                                                         │
│   PUBLIC REPOSITORIES (3)                                               │
│   ═══════════════════════                                               │
│                                                                         │
│   ┌────────────────────────────────────────────────────────────────┐   │
│   │  SynkraAI/aiox-core (Commons Clause)                           │   │
│   │  • Core Framework & Orchestration Engine                       │   │
│   │  • 11 Base Agents (Dex, Quinn, Aria, etc.)                     │   │
│   │  • Task Runner & Workflow Engine                               │   │
│   │  • Quality Gates System                                        │   │
│   │  • Service Discovery                                           │   │
│   │  npm: @aiox/core                                               │   │
│   └────────────────────────────────────────────────────────────────┘   │
│                              ▲                                          │
│                              │ peerDependency                           │
│   ┌──────────────────────────┼──────────────────────────┐               │
│   ▼                          │                          ▼               │
│   ┌─────────────────────┐    │    ┌─────────────────────────────┐      │
│   │ SynkraAI/           │    │    │ SynkraAI/mcp-ecosystem      │      │
│   │ aiox-squads (MIT)   │    │    │ (Apache 2.0)                │      │
│   │ • ETL Squad         │    │    │ • Docker MCP Toolkit        │      │
│   │ • Creator Squad     │    │    │ • IDE Configurations        │      │
│   │ • MMOS Squad        │    │    │ • MCP Presets               │      │
│   │ npm: @aiox/squad-*  │    │    │ npm: @aiox/mcp-presets      │      │
│   └─────────────────────┘    │    └─────────────────────────────┘      │
│                              │                                          │
│   PRIVATE REPOSITORIES (2)                                              │
│   ════════════════════════                                              │
│   ┌─────────────────────┐         ┌─────────────────────────────┐      │
│   │ SynkraAI/mmos       │         │ SynkraAI/certified-partners │      │
│   │ (Proprietary + NDA) │         │ (Proprietary)               │      │
│   │ • MMOS Minds        │         │ • Premium Squads            │      │
│   │ • Cognitive Clones  │         │ • Partner Portal            │      │
│   │ • DNA Mental™       │         │ • Marketplace               │      │
│   └─────────────────────┘         └─────────────────────────────┘      │
└─────────────────────────────────────────────────────────────────────────┘
```

### Competitive Positioning

| Framework     | Open-Source Completeness | Unique Differentiator       |
| ------------- | ------------------------ | --------------------------- |
| LangChain     | ✅ Complete              | ❌ None (commodity)         |
| CrewAI        | ✅ Complete              | ❌ None (commodity)         |
| AutoGen       | ✅ Complete              | ❌ None (commodity)         |
| **AIOX v4.2** | ✅ **Complete**          | ✅ **Clones (DNA Mental™)** |

**Analogy:** Linux is open source, but Red Hat Enterprise Linux adds support and optimizations. Both are Linux, but the added value varies. AIOX works the same way.

---

## 📖 How to Use This Book

This is not a document to be read from start to finish. It is a **layered learning system**:

- 🚀 **Layer 0: DISCOVERY** - Find your path (5 min)
- 🎯 **Layer 1: UNDERSTANDING** - 5 essays that teach the mental model (75 min)
- 🎨 **Layer 2: COMPONENT LIBRARY** - Complete component catalog
- 📋 **Layer 3: USAGE GUIDE** - How to use AIOX v4.2 in your context
- 📚 **Layer 4: COMPLETE REFERENCE** - Full technical specification
- 🔄 **META: EVOLUTION** - How to contribute and evolve the framework

**Most people only need Layer 1.** The rest exists for when you need it.

---

# 🚀 LAYER 0: DISCOVERY ROUTER

## Welcome to AIOX v4.2 - Let's Find Your Path

### Available Learning Tracks

| Track                       | Time      | Best For                                |
| --------------------------- | --------- | --------------------------------------- |
| **Track 1: Quick Start**    | 15-30 min | Curious explorers, fast decision-makers |
| **Track 2: Deep Dive**      | 1.5-2h    | Active builders with real pain points   |
| **Track 3: Mastery Path**   | Weeks     | Framework developers, power users       |
| **Track 4: Decision Maker** | 30-45 min | Leaders evaluating adoption             |
| **Track 5: Targeted**       | Variable  | Need something specific                 |
| **Track 6: v2.0 Upgrade**   | 45-60 min | v2.0 users migrating                    |

---

# 🎯 LAYER 1: UNDERSTANDING

## Essay 1: Why AIOX Exists

### The Problem

Development with AI agents today is **chaotic**:

- Agents without coordination
- Inconsistent results
- No quality gates
- Context lost between sessions
- Every project reinvents the wheel

### The Solution

AIOX provides **structured orchestration**:

- 11 specialized agents with personalities
- Coordinated multi-agent workflows
- Quality Gates in 3 layers
- Task-First Architecture for portability
- Service Discovery for reuse

---

## Essay 2: Structure is Sacred

> "When information is always in the same positions, our brain knows where to look quickly."

**FIXED (Structure):**

- Template positions
- Section order
- Metric formats
- File structure
- Task workflows

**FLEXIBLE (Tone):**

- Status messages
- Vocabulary choices
- Emoji usage
- Agent personality
- Communication tone

---

## Essay 3: Business Model v4.2

### Why Workers Are Now Open-Source

1. **Workers are Commodity** - Any developer can write deterministic scripts
2. **Clones are Singularity** - DNA Mental™ takes years to develop
3. **Maximum Adoption Strategy** - Zero friction to start
4. **Network Effects** - More users → More contributors → Better Workers

### What Remains Proprietary?

- **Clones** - Cognitive emulation via DNA Mental™
- **Premium Squads** - Industry expertise (Finance, Healthcare, etc.)
- **Team Features** - Collaboration, shared memory
- **Enterprise** - Scale, support, SLAs

---

## Essay 4: Agent System

### The 11 Agents v4.2

| Agent     | ID              | Archetype    | Responsibility          |
| --------- | --------------- | ------------ | ----------------------- |
| **Dex**   | `dev`           | Builder      | Code implementation     |
| **Quinn** | `qa`            | Guardian     | Quality assurance       |
| **Aria**  | `architect`     | Architect    | Technical architecture  |
| **Nova**  | `po`            | Visionary    | Product backlog         |
| **Kai**   | `pm`            | Balancer     | Product strategy        |
| **River** | `sm`            | Facilitator  | Process facilitation    |
| **Zara**  | `analyst`       | Explorer     | Business analysis       |
| **Dara**  | `data-engineer` | Architect    | Data engineering        |
| **Felix** | `devops`        | Optimizer    | CI/CD and operations    |
| **Uma**   | `ux-expert`     | Creator      | User experience         |
| **Pax**   | `aiox-master`   | Orchestrator | Framework orchestration |

### Agent Activation

```bash
# Activate agent
@dev             # Activates Dex (Developer)
@qa              # Activates Quinn (QA)
@architect       # Activates Aria (Architect)
@aiox-master     # Activates Pax (Orchestrator)

# Agent commands (prefix *)
*help            # Show available commands
*task <name>     # Execute specific task
*exit            # Deactivate agent
```

---

## Essay 5: Task-First Architecture

### The Philosophy

> **"Everything is a Task. Executors are attributes."**

### What This Means

**Traditional (Task-per-Executor):**

```yaml
# 2 separate implementations for the same task
agent_task.md:
  executor: Agent (Sage)

worker_task.js:
  executor: Worker (market-analyzer.js)
```

**Task-First (Universal Task):**

```yaml
# ONE task definition
task: analyzeMarket()
inputs: { market_data: object }
outputs: { insights: array }

# Executor is just a field
executor_type: Human    # Day 1
executor_type: Worker   # Week 10
executor_type: Agent    # Month 6
executor_type: Clone    # Year 2
```

### Instant Migration

- **Before:** 2-4 days (rewrite required)
- **After:** 2 seconds (change 1 field)

---

# 🎨 LAYER 2: COMPONENT LIBRARY

## Modular Architecture v4.2

### The 4 Modules

```
.aiox-core/
├── core/              # Framework foundations
│   ├── config/        # Configuration management
│   ├── registry/      # Service Discovery
│   ├── quality-gates/ # 3-layer QG system
│   ├── mcp/           # MCP global configuration
│   └── session/       # Session management
│
├── development/       # Development artifacts
│   ├── agents/        # 11 agent definitions
│   ├── tasks/         # 115+ task definitions
│   ├── workflows/     # 7 workflow definitions
│   └── scripts/       # Dev support utilities
│
├── product/           # User-facing templates
│   ├── templates/     # 52+ templates
│   ├── checklists/    # 11 checklists
│   └── data/          # PM knowledge base
│
└── infrastructure/    # System configuration
    ├── scripts/       # 55+ infrastructure scripts
    ├── tools/         # CLI, MCP, local configs
    └── integrations/  # PM adapters (ClickUp, Jira)
```

### Module Dependencies

```
┌─────────────────────────────────────────────────────┐
│                 CLI / Tools                          │
│                     │                                │
│      ┌──────────────┼──────────────┐                │
│      ▼              ▼              ▼                │
│  development    product    infrastructure           │
│      │              │              │                │
│      └──────────────┼──────────────┘                │
│                     ▼                                │
│                   core                               │
│           (no dependencies)                          │
└─────────────────────────────────────────────────────┘

Rules:
• core/ has no internal dependencies
• development/, product/, infrastructure/ depend ONLY on core/
• Circular dependencies are PROHIBITED
```

---

## Squad System (New in v4.2)

### Terminology

| Old Term      | New Term           | Description            |
| ------------- | ------------------ | ---------------------- |
| Squad         | **Squad**          | Modular AI agent teams |
| Squads/       | **squads/**        | Squads directory       |
| pack.yaml     | **squad.yaml**     | Squad manifest         |
| @expansion/\* | **@aiox/squad-\*** | npm scope              |

### Squad Structure

```
squads/
├── etl-squad/
│   ├── squad.yaml         # Manifest
│   ├── agents/            # Squad-specific agents
│   ├── tasks/             # Squad tasks
│   └── templates/         # Squad templates
├── creator-squad/
└── mmos-squad/
```

### Squad Manifest (squad.yaml)

```yaml
name: etl-squad
version: 1.0.0
description: Data pipeline and ETL automation squad
license: MIT

peerDependencies:
  '@aiox/core': '^2.1.0'

agents:
  - id: etl-orchestrator
    extends: data-engineer
  - id: data-validator
    extends: qa

tasks:
  - collect-sources
  - transform-data
  - validate-pipeline

exports:
  - agents
  - tasks
  - templates
```

---

## Quality Gates 3 Layers

### Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     QUALITY GATES 3 LAYERS                              │
│                                                                         │
│   LAYER 1: LOCAL (Pre-commit)                                           │
│   • Linting, formatting, vet checks                                     │
│   • Unit tests (fast)                                                   │
│   • Executor: Worker (deterministic)                                    │
│   • Blocking: Can't commit if fails                                     │
│   • Catches: 30% of issues instantly                                    │
│                                                                         │
│   LAYER 2: PR AUTOMATION (CI/CD)                                        │
│   • CodeRabbit AI review                                                │
│   • Integration tests, coverage                                         │
│   • Security scan, performance                                          │
│   • Executor: Agent (QA) + CodeRabbit                                  │
│   • Blocking: Required checks for merge                                 │
│   • Catches: Additional 50% (80% total)                                │
│                                                                         │
│   LAYER 3: HUMAN REVIEW (Strategic)                                     │
│   • Architecture alignment                                              │
│   • Business logic correctness                                          │
│   • Edge cases, documentation                                           │
│   • Executor: Human (Senior Dev / Tech Lead)                           │
│   • Blocking: Final approval required                                   │
│   • Catches: Final 20% (100% total)                                    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Story Template v2.0

### Complete Structure

````markdown
# Story X.X: [Title]

**Epic:** [Parent Epic]
**Story ID:** X.X
**Sprint:** [Number]
**Priority:** 🔴 Critical | 🟠 High | 🟡 Medium | 🟢 Low
**Points:** [Number]
**Status:** ⚪ Ready | 🔄 In Progress | ✅ Done

---

## 🔀 Cross-Story Decisions

| Decision        | Source     | Impact on This Story        |
| --------------- | ---------- | --------------------------- |
| [Decision Name] | [Story ID] | [How it affects this story] |

---

## 📋 User Story

**As** [persona],
**I want** [action],
**So that** [benefit].

---

## ✅ Tasks

### Phase 1: [Name]

- [ ] **1.1** [Task description]
- [ ] **1.2** [Task description]

---

## 🎯 Acceptance Criteria

```gherkin
GIVEN [context]
WHEN [action]
THEN [expected result]
```

---

## 🤖 CodeRabbit Integration

### Story Type Analysis

| Attribute         | Value             | Rationale |
| ----------------- | ----------------- | --------- |
| Type              | [Type]            | [Why]     |
| Complexity        | [Low/Medium/High] | [Why]     |
| Test Requirements | [Type]            | [Why]     |

### Agent Assignment

| Role      | Agent | Responsibility |
| --------- | ----- | -------------- |
| Primary   | @dev  | [Task]         |
| Secondary | @qa   | [Task]         |

---

## 🧑‍💻 Dev Agent Record

### Execution Log

| Timestamp | Phase | Action | Result |
| --------- | ----- | ------ | ------ |

---

## 🧪 QA Results

### Test Execution Summary

| Check | Status | Notes |
| ----- | ------ | ----- |
````

---

## npm Package Scoping

### Package Structure

| Package               | Registry   | Depends On  | License        |
| --------------------- | ---------- | ----------- | -------------- |
| `@aiox/core`          | npm public | -           | Commons Clause |
| `@aiox/squad-etl`     | npm public | @aiox/core  | MIT            |
| `@aiox/squad-creator` | npm public | @aiox/core  | MIT            |
| `@aiox/squad-mmos`    | npm public | @aiox/core  | MIT            |
| `@aiox/mcp-presets`   | npm public | -           | Apache 2.0     |

---

# 📋 LAYER 3: USAGE GUIDE

## Quick Start v4.2

### Installation (5 minutes)

```bash
# New project (Greenfield)
$ npx @SynkraAI/aiox@latest init

# Existing project (Brownfield)
$ npx @SynkraAI/aiox migrate v2.0-to-v4.0.4
```

### First Steps

```bash
$ aiox agents list
$ aiox squads list
$ aiox stories create
$ aiox task develop-story --story=1.1
```

---

## Workflows

### Available Workflows

| Workflow                 | Use Case                | Agents Involved   |
| ------------------------ | ----------------------- | ----------------- |
| `greenfield-fullstack`   | New full-stack project  | All agents        |
| `brownfield-integration` | Add AIOX to existing    | dev, architect    |
| `fork-join`              | Parallel task execution | Multiple          |
| `organizer-worker`       | Delegated execution     | po, dev           |
| `data-pipeline`          | ETL workflows           | data-engineer, qa |

---

# 📚 LAYER 4: COMPLETE REFERENCE

## Key Metrics Comparison

### Installation

| Metric          | v2.0       | v4.2      | Improvement     |
| --------------- | ---------- | --------- | --------------- |
| Time to install | 2-4 hours  | 5 minutes | **96% faster**  |
| Steps required  | 15+ manual | 1 command | **93% simpler** |
| Success rate    | 60%        | 98%       | **+38%**        |

### Development Speed

| Metric                | v2.0     | v4.2       | Improvement       |
| --------------------- | -------- | ---------- | ----------------- |
| Find reusable Worker  | N/A      | 30 seconds | **∞**             |
| Quality issues caught | 20%      | 80%        | **4x**            |
| Executor migration    | 2-4 days | 2 seconds  | **99.99% faster** |

### Quality

| Metric              | v2.0       | v4.2          |
| ------------------- | ---------- | ------------- |
| Quality Gate Layers | 1 (manual) | 3 (automated) |
| Auto-caught issues  | 0%         | 80%           |
| Human review time   | 2-4h/PR    | 30min/PR      |

---

## Version History

| Version | Date       | Changes                                           |
| ------- | ---------- | ------------------------------------------------- |
| 2.0.0   | 2025-01-19 | Initial v2.0 release                              |
| 2.1.0   | 2025-12-09 | Modular arch, Squads, Multi-repo, QG3, Story v2.0 |

---

## Related Documents

- [QUALITY-GATES-SPECIFICATION.md](./QUALITY-GATES-SPECIFICATION.md)
- [STORY-TEMPLATE-V2-SPECIFICATION.md](./STORY-TEMPLATE-V2-SPECIFICATION.md)
- [STANDARDS-INDEX.md](./STANDARDS-INDEX.md)
- [BIFROST-PROJECT-CONTEXT.md](./BIFROST-PROJECT-CONTEXT.md)

---

**Last Updated:** 2025-12-09
**Version:** 2.1.0-complete
**Maintained By:** AIOX Framework Team
