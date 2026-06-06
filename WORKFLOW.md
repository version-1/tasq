---
polling:
  interval_ms: 30000
workspace:
  root: .worktrees/agents
agent:
  max_concurrent_agents: 5
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
Write agent plan files, pull request summary drafts, and other temporary
scratch files under the repository `.tmp/` directory. Do not place these
temporary artifacts under `~/codex`, `$CODEX_HOME`, or other external home
directories.

## Required Flow

1. Confirm the task scope from the issue title and description above.
2. Create or switch to an isolated task branch/worktree before editing.
3. Make focused changes that satisfy the issue.
4. Run the narrowest useful verification first, then broaden checks when shared behavior is affected.
5. Commit the change and create or update a pull request.
6. Leave a progress or handoff comment on the issue.
7. Move the issue to `review` when the pull request is ready for human review.
