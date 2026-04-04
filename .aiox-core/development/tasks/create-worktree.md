# create-worktree

**Task ID:** create-worktree
**Version:** 1.0
**Created:** 2026-01-28 (Story 1.3)
**Agent:** @devops (Gage)

---

## Execution Modes

**Choose your execution mode:**

### 1. YOLO Mode - Fast, Autonomous (0-1 prompts) **[DEFAULT]**

- Autonomous worktree creation
- Minimal user interaction
- **Best for:** Quick story setup

### 2. Interactive Mode - Balanced, Educational (2-3 prompts)

- Confirms story ID and options
- Shows worktree path before creation
- **Best for:** First-time users

**Parameter:** `mode` (optional, default: `yolo`)

---

## Task Definition (AIOX Task Format V1.0)

```yaml
task: createWorktree()
responsible: Gage (DevOps)
executor_type: Agente
atomic_layer: Atom

inputs:
  - campo: story_id
    type: string
    source: User Input
    required: true
    validation: Valid story identifier (e.g., 'STORY-42', '1.3', 'fix-auth')

  - campo: options
    type: object
    source: User Input
    required: false
    validation: Optional configuration overrides

outputs:
  - campo: worktree_info
    type: WorktreeInfo
    destination: Return value
    persisted: false

  - campo: worktree_path
    type: string
    destination: File system
    persisted: true
```

---

## Pre-Conditions

**Purpose:** Validate prerequisites BEFORE task execution (blocking)

```yaml
pre-conditions:
  - [ ] Current directory is a git repository
    type: pre-condition
    blocker: true
    validation: git rev-parse --is-inside-work-tree
    error_message: "Not a git repository. Initialize git first."

  - [ ] WorktreeManager is available
    type: pre-condition
    blocker: true
    validation: Script exists at .aiox-core/infrastructure/scripts/worktree-manager.js
    error_message: "WorktreeManager not found. Ensure AIOX is properly installed."

  - [ ] Max worktrees limit not reached
    type: pre-condition
    blocker: true
    validation: Current worktrees < maxWorktrees (default: 10)
    error_message: "Maximum worktrees limit reached. Remove stale worktrees first."
```

---

## Post-Conditions

**Purpose:** Validate execution success AFTER task completes

```yaml
post-conditions:
  - [ ] Worktree directory exists
    type: post-condition
    blocker: true
    validation: Directory exists at .aiox/worktrees/{storyId}
    error_message: "Worktree directory was not created."

  - [ ] Branch exists
    type: post-condition
    blocker: true
    validation: Branch auto-claude/{storyId} exists
    error_message: "Worktree branch was not created."
```

---

## Acceptance Criteria

```yaml
acceptance-criteria:
  - [ ] Worktree created with isolated git state
    type: acceptance-criterion
    blocker: true

  - [ ] Branch follows naming convention auto-claude/{storyId}
    type: acceptance-criterion
    blocker: true

  - [ ] Worktree appears in list
    type: acceptance-criterion
    blocker: true
```

---

## Tools

**External resources used by this task:**

- **Tool:** WorktreeManager
  - **Purpose:** Git worktree operations
  - **Source:** .aiox-core/infrastructure/scripts/worktree-manager.js

- **Tool:** git
  - **Purpose:** Version control operations
  - **Source:** System git installation

---

## Description

Creates an isolated Git worktree for developing a story in parallel. Each worktree has its own working directory and branch, enabling multiple stories to be worked on simultaneously without conflicts.

**Use cases:**

- Start working on a new story in isolation
- Enable Auto-Claude to develop stories autonomously
- Run parallel development tracks

---

## Inputs

| Parameter  | Type   | Required | Default | Description                                |
| ---------- | ------ | -------- | ------- | ------------------------------------------ |
| `story_id` | string | Yes      | -       | Story identifier (e.g., 'STORY-42', '1.3') |

---

## Elicitation

```yaml
elicit: false
```

This task runs autonomously. If story_id is not provided, prompt once.

---

## Steps

### Step 1: Validate Git Repository

**Action:** Verify current directory is a git repository

```bash
git rev-parse --is-inside-work-tree 2>/dev/null
```

**Exit Condition:** If not a git repo:

```
❌ Not a git repository.
   Initialize git first: git init
```

---

### Step 2: Parse Story ID

**Action:** Extract and validate story ID from input

**Validation:**

- Must be non-empty string
- Can contain alphanumeric, hyphens, dots, underscores
- Examples: `STORY-42`, `1.3`, `fix-auth-bug`

**If missing, prompt:**

```
📝 Enter story ID for the worktree:
   Example: STORY-42, 1.3, fix-auth-bug
```

---

### Step 3: Check Existing Worktree

**Action:** Verify worktree doesn't already exist

```javascript
const WorktreeManager = require('./.aiox-core/infrastructure/scripts/worktree-manager.js');
const manager = new WorktreeManager();
const exists = await manager.exists(storyId);
```

**If exists:**

```
⚠️  Worktree for '{storyId}' already exists.
    Path: .aiox/worktrees/{storyId}
    Branch: auto-claude/{storyId}

    Use *list-worktrees to see all worktrees.
```

---

### Step 4: Check Worktree Limit

**Action:** Ensure we haven't reached max worktrees

```javascript
const count = await manager.getCount();
if (count.total >= manager.maxWorktrees) {
  // Show error with stale worktrees to clean up
}
```

**If limit reached:**

```
❌ Maximum worktrees limit (10) reached.

   Current worktrees: 10
   Stale worktrees: {count.stale}

   Run *cleanup-worktrees to remove stale worktrees, or
   Run *remove-worktree {storyId} to remove a specific one.
```

---

### Step 5: Create Worktree

**Action:** Create the worktree using WorktreeManager

```javascript
const worktreeInfo = await manager.create(storyId);
```

**Creates:**

- Directory: `.aiox/worktrees/{storyId}/`
- Branch: `auto-claude/{storyId}`

---

### Step 6: Display Success

**Action:** Show creation confirmation

```
╔══════════════════════════════════════════════════════════════╗
║  ✅ Worktree Created Successfully                            ║
╚══════════════════════════════════════════════════════════════╝

Story:    {storyId}
Path:     .aiox/worktrees/{storyId}
Branch:   auto-claude/{storyId}
Status:   active

Next Steps:
  • cd .aiox/worktrees/{storyId}  - Navigate to worktree
  • git status                    - Check worktree state
  • *list-worktrees               - See all worktrees
  • *merge-worktree {storyId}     - Merge back when done
```

---

## Outputs

### Return Value

```typescript
interface WorktreeInfo {
  storyId: string; // 'STORY-42'
  path: string; // '/abs/path/.aiox/worktrees/STORY-42'
  branch: string; // 'auto-claude/STORY-42'
  createdAt: Date; // Creation timestamp
  uncommittedChanges: number; // 0 (new worktree)
  status: 'active' | 'stale'; // 'active'
}
```

### File System

- `.aiox/worktrees/{storyId}/` - Isolated worktree directory

---

## Validation

- [ ] Worktree directory exists and is accessible
- [ ] Git branch `auto-claude/{storyId}` exists
- [ ] Worktree appears in `git worktree list`
- [ ] Worktree is clean (no uncommitted changes)

---

## Error Handling

### Not a Git Repository

**Error:**

```
❌ Not a git repository.
```

**Resolution:** Run `git init` first.

### Worktree Already Exists

**Error:**

```
⚠️  Worktree for '{storyId}' already exists.
```

**Resolution:** Use existing worktree or remove it first.

### Max Worktrees Reached

**Error:**

```
❌ Maximum worktrees limit (10) reached.
```

**Resolution:** Run `*cleanup-worktrees` or `*remove-worktree`.

### Git Worktree Command Failed

**Error:**

```
❌ Failed to create worktree: {error.message}
```

**Resolution:** Check git status and ensure no conflicts.

---

## Rollback

To remove a created worktree:

```bash
*remove-worktree {storyId}
```

Or manually:

```bash
git worktree remove .aiox/worktrees/{storyId}
git branch -d auto-claude/{storyId}
```

---

## Performance Notes

- **Creation time:** ~500ms-2s (depends on repo size)
- **Disk usage:** Same as shallow clone (hardlinks for objects)
- **Branch overhead:** Minimal (just ref pointer)

---

## Dependencies

### Scripts

- `.aiox-core/infrastructure/scripts/worktree-manager.js` - Core manager

### NPM Packages

- `execa` - Git command execution
- `chalk` - Terminal colors

### Git Commands Used

- `git worktree add` - Create worktree
- `git branch` - Create/manage branches

---

## Related

- **Story:** 1.3 - CLI Commands for Worktree Management
- **Script:** `.aiox-core/infrastructure/scripts/worktree-manager.js`
- **Tasks:** `list-worktrees.md`, `remove-worktree.md`, `merge-worktree.md`

---

## Command Registration

This task is exposed as CLI command `*create-worktree` in @devops agent:

```yaml
commands:
  - 'create-worktree {storyId}': Create isolated worktree for story development
```

---

**Status:** ✅ Production Ready
**Tested On:** Windows, Linux, macOS
**Git Requirement:** git >= 2.5 (worktree support)
