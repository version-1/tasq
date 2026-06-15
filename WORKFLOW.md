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
  read_timeout_ms: 15000
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
2. Before branch creation, run `git fetch origin`, then create the task branch from `origin/main` or the repository's default branch.
3. Create or switch to an isolated task branch/worktree before editing.
4. Make focused changes that satisfy the issue.
5. Run the narrowest useful verification first, then broaden checks when shared behavior is affected.
6. Commit the change and create or update a pull request.
7. Leave a progress or handoff comment on the issue.
8. Move the issue to `review` when the pull request is ready for human review.

## Complete the issue

1. Run a review and resolve every High or higher severity finding.
2. Leave an issue comment that summarizes the review result, including whether any High or higher severity findings were found and how they were handled.
3. Create a pull request.
4. After steps 1 through 3 are complete, run `tq issue update {{ issue.id }} --status review`
