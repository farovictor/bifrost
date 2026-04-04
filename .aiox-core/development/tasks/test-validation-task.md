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

## Task Definition (AIOX Task Format V2.0)

```yaml
task: testValidationTask()
responsible: Dex (Dev Agent)
executor_type: Agente
atomic_layer: Test

**Input:**
- campo: test_input
  type: string
  source: User Input
  required: false
  validation: Optional test input parameter

**Output:**
- campo: validation_result
  type: object
  destination: Memory
  persisted: false

- campo: success
  type: boolean
  destination: Return value
  persisted: false
```

---

## Pre-Conditions

**Purpose:** Validate prerequisites BEFORE task execution (blocking)

**Checklist:**

```yaml
pre-conditions:
  - [ ] Test environment available
    type: pre-condition
    blocker: true
    validation: |
      Verify test environment is available
    error_message: "Pre-condition failed: Test environment not available"
```

---

## Post-Conditions

**Purpose:** Validate execution success AFTER task completes

**Checklist:**

```yaml
post-conditions:
  - [ ] Validation completed successfully
    type: post-condition
    blocker: true
    validation: |
      Verify validation completed successfully
    error_message: "Post-condition failed: Validation did not complete successfully"
```

---

## Acceptance Criteria

**Purpose:** Definitive pass/fail criteria for task completion

**Checklist:**

```yaml
acceptance-criteria:
  - [ ] Task executed successfully
    type: acceptance-criterion
    blocker: true
    validation: |
      Assert task executed successfully
    error_message: "Acceptance criterion not met: Task did not execute successfully"
```

---

## Purpose

This is a test task created for validating the `create-task` task execution. It provides minimal functionality to test task creation workflow.

## Implementation

1. **Validate Inputs**
   - Check test input if provided
   - Validate environment

2. **Execute Validation**
   - Perform simple validation test
   - Return success status

3. **Output Result**
   - Return validation result
   - Log execution

## Error Handling

**Strategy:** abort

**Common Errors:**

1. **Error:** Test Environment Not Available
   - **Cause:** Test environment not configured
   - **Resolution:** Ensure test environment is available
   - **Recovery:** Log error and abort

---

## Performance

**Expected Metrics:**

```yaml
duration_expected: < 1 min
cost_estimated: $0.0001
token_usage: ~100-200 tokens
```

---

## Metadata

```yaml
story: STORY-6.1.7.2
version: 1.0.0
dependencies:
  - N/A
tags:
  - test
  - validation
updated_at: 2025-01-17
```

---

**Created By:** Dex (Dev Agent)  
**Created Date:** 2025-01-17  
**Purpose:** Test task for validating create-task task execution  
**Status:** Test/Validation Only

