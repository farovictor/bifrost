# Task: Migration Dry-Run

**Purpose**: Execute migration inside BEGIN…ROLLBACK to catch syntax/ordering errors

**Elicit**: true

---

## Execution Modes

**Choose your execution mode:**

### 1. YOLO Mode - Fast, Autonomous (0-1 prompts)
- Autonomous decision making with logging
- Minimal user interaction
- **Best for:** Simple, deterministic tasks

### 2. Interactive Mode - Balanced, Educational (5-10 prompts) **[DEFAULT]**
- Explicit decision checkpoints
- Educational explanations
- **Best for:** Learning, complex decisions

### 3. Pre-Flight Planning - Comprehensive Upfront Planning
- Task analysis phase (identify all ambiguities)
- Zero ambiguity execution
- **Best for:** Ambiguous requirements, critical work

**Parameter:** `mode` (optional, default: `interactive`)

---

## Task Definition (AIOX Task Format V1.0)

```yaml
task: dbDryRun()
responsible: Dara (Sage)
executor_type: Agente
atomic_layer: Organism

inputs:
  - field: query
  type: string
  source: User Input
  required: true
  validation: Valid SQL query

- campo: params
  type: object
  source: User Input
  required: false
  validation: Query parameters

- campo: connection
  type: object
  source: config
  required: true
  validation: Valid PostgreSQL connection via Supabase

outputs:
  - field: query_result
  type: array
  destination: Memory
  persisted: false

- campo: records_affected
  type: number
  destination: Return value
  persisted: false

- campo: execution_time
  type: number
  destination: Memory
  persisted: false
```

---

## Pre-Conditions

**Purpose:** Validate prerequisites BEFORE task execution (blocking)

**Checklist:**

```yaml
pre-conditions:
  - [ ] Database connection established; query syntax valid
    type: pre-condition
    blocker: true
    validation: |
      Check database connection established; query syntax valid
    error_message: "Pre-condition failed: Database connection established; query syntax valid"
```

---

## Post-Conditions

**Purpose:** Validate execution success AFTER task completes

**Checklist:**

```yaml
post-conditions:
  - [ ] Query executed; results returned; transaction committed
    type: post-condition
    blocker: true
    validation: |
      Verify query executed; results returned; transaction committed
    error_message: "Post-condition failed: Query executed; results returned; transaction committed"
```

---

## Acceptance Criteria

**Purpose:** Definitive pass/fail criteria for task completion

**Checklist:**

```yaml
acceptance-criteria:
  - [ ] Data persisted correctly; constraints respected; no orphaned data
    type: acceptance-criterion
    blocker: true
    validation: |
      Assert data persisted correctly; constraints respected; no orphaned data
    error_message: "Acceptance criterion not met: Data persisted correctly; constraints respected; no orphaned data"
```

---

## Tools

**External/shared resources used by this task:**

- **Tool:** supabase
  - **Purpose:** PostgreSQL database connection via Supabase client
  - **Source:** @supabase/supabase-js

- **Tool:** query-validator
  - **Purpose:** SQL query syntax validation
  - **Source:** .aiox-core/utils/db-query-validator.js

---

## Scripts

**Agent-specific code for this task:**

- **Script:** db-query.js
  - **Purpose:** Execute PostgreSQL queries with error handling via Supabase
  - **Language:** JavaScript
  - **Location:** .aiox-core/scripts/db-query.js

---

## Error Handling

**Strategy:** retry

**Common Errors:**

1. **Error:** Connection Failed
   - **Cause:** Unable to connect to Neo4j database
   - **Resolution:** Check connection string, credentials, network
   - **Recovery:** Retry with exponential backoff (max 3 attempts)

2. **Error:** Query Syntax Error
   - **Cause:** Invalid Cypher query syntax
   - **Resolution:** Validate query syntax before execution
   - **Recovery:** Return detailed syntax error, suggest fix

3. **Error:** Transaction Rollback
   - **Cause:** Query violates constraints or timeout
   - **Resolution:** Review query logic and constraints
   - **Recovery:** Automatic rollback, preserve data integrity

---

## Performance

**Expected Metrics:**

```yaml
duration_expected: 5-15 min (estimated)
cost_estimated: $0.003-0.010
token_usage: ~3,000-10,000 tokens
```

**Optimization Notes:**
- Break into smaller workflows; implement checkpointing; use async processing where possible

---

## Metadata

```yaml
story: N/A
version: 1.0.0
dependencies:
  - N/A
tags:
  - database
  - infrastructure
updated_at: 2025-11-17
```

---


## Inputs

- `path` (string): Path to SQL migration file

---

## Process

### 1. Confirm Migration File

Ask user to confirm:
- Migration file path: `{path}`
- Purpose of this migration
- Expected changes (tables, functions, etc)

### 2. Execute Dry-Run

Run migration in transaction that will be rolled back:

```bash
psql "$SUPABASE_DB_URL" -v ON_ERROR_STOP=1 <<'SQL'
BEGIN;
\echo 'Starting dry-run...'
\i {path}
\echo 'Dry-run completed successfully - rolling back...'
ROLLBACK;
SQL
```

### 3. Report Results

**If successful:**
```
✓ Dry-run completed without errors
✓ Migration syntax is valid
✓ No dependency or ordering issues detected
```

**If failed:**
```
❌ Dry-run failed
Error: [error message]
Line: [line number if available]
Fix the migration and try again
```

---

## What This Validates

- ✅ SQL syntax correctness
- ✅ Object dependencies exist
- ✅ Execution order is valid
- ✅ No constraint violations
- ❌ Does NOT validate data correctness
- ❌ Does NOT check performance

---

## Next Steps After Success

1. Review migration one more time
2. Take snapshot: `*snapshot pre_migration`
3. Apply migration: `*apply-migration {path}`
4. Run smoke tests: `*smoke-test`

---

## Error Handling

Common errors and fixes:

**"relation does not exist"**
- Missing table/view dependency
- Check if you need to create dependent objects first

**"function does not exist"**
- Function called before creation
- Reorder: tables → functions → triggers

**"syntax error"**
- Check SQL syntax
- Verify PostgreSQL version compatibility
