---
polling:
  interval_ms: 30000
workspace:
  root: .worktrees
agent:
  max_concurrent_agents: 1
  max_turns: 20
  continuation_turns_enabled: false
  max_retry_attempts: 3
  max_retry_backoff_ms: 300000
codex:
  command: codex --sandbox workspace-write app-server
  read_timeout_ms: 5000
  turn_timeout_ms: 3600000
  stall_timeout_ms: 300000
server:
  port: 8081
---

# Task

Issue ID: {{ issue.id }}
Title: {{ issue.title }}
Status: {{ issue.status }}
Priority: {{ issue.priority }}
Assignee: {{ issue.assignee }}
Attempt: {{ attempt }}

## Description

{{ issue.description }}

## Repository Guidance

Follow the development workflow in [docs/development.md](docs/development.md).
Read the component workflow document for the area you change before editing.

## Required Flow

1. Confirm the task scope from the issue title and description above.
2. Create or switch to an isolated task branch/worktree before editing.
3. Make focused changes that satisfy the issue.
4. Run the narrowest useful verification first, then broaden checks when shared behavior is affected.
5. Commit the change and create or update a pull request.
6. Leave a progress or handoff comment on the issue.
7. Move the issue to `review` when the pull request is ready for human review.
8. If work is blocked after the issue has been moved to `in_progress`, add a
   `blocker` comment and update the issue status to `blocked` before ending the
   run. Do this even when the local implementation or commit is complete but
   push, pull request creation, verification, or another handoff step failed.

Use `tq` to keep the issue tracker synchronized:

```sh
tq issue update {{ issue.id }} --status in_progress
tq comment add {{ issue.id }} --type progress --body "Started work."
tq issue update {{ issue.id }} --status review
tq comment add {{ issue.id }} --type blocker --body "Blocked because <reason>."
tq issue update {{ issue.id }} --status blocked
```

Run the installed `tq` binary from `PATH`. Do not use `go run ./cmd/tq` for
tracker synchronization.
