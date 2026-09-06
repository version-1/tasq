---
name: tq-orchestrator
description: Poll ready issues from the tq issue tracker with the Monitor tool and delegate them to subagents
---

# Purpose

Poll ready issues from the Tasq issue tracker and delegate each issue to a subagent that follows the project's workflow.

# Constraints

- Run this skill only inside a Claude Code agent. Return an error in other environments.
- The orchestrator must delegate issue resolution to subagents and must not resolve issues itself.
- A project's `WORKFLOW.md` takes precedence. Use the default delegated workflow below only for behavior that the project workflow does not specify.

# Orchestration

1. Ask the user for the polling interval in seconds (default: 30).
2. Use the **Monitor tool** to run `tq issue watch --interval <seconds>`.
3. For each `issue-ready` event, read the issue id from `body.id` and fetch the issue with `tq issue get <issue_id>`.
4. Delegate the issue to a subagent using the project `WORKFLOW.md` and fill only its unspecified behavior from the default delegated workflow below.
5. After a successful hand-off:
   - Change the issue status with `tq issue update <issue_id> --status in_progress`.
   - Record the subagent name with `tq comment add <issue_id> --author claude-code --body "<comment>"`.

# Default Delegated Workflow

1. Confirm the scope from the issue title and description.
2. Before editing, create an isolated task branch and worktree from the latest default branch:
   - Run `git fetch origin`.
   - Create the task branch from `origin/main` or the repository's default branch.
   - Create the worktree at `<project_root>/.worktrees/tasq/<issue_id>-<issue-summary-title>`.
   - Record the worktree and branch names with `tq comment add <issue_id> --author claude-code --body "<comment>"`.
3. Make focused changes and run the narrowest useful verification first, broadening checks when shared behavior is affected.
4. Commit the changes and create or update a pull request.
5. After the pull request is available:
   - Record its title and URL with `tq comment add <issue_id> --author claude-code --body "<comment>"`.
   - Register its URL with `tq artifact set <issue_id> --type pull_request <pr_url>`.
   - Leave a comment for any unresolved items or hand-off notes.
   - Change the issue status with `tq issue update <issue_id> --status review`.
6. Delete the task worktree and branch only after confirming that the pull request exists and the issue is in review. Otherwise, preserve the worktree.

If the task cannot be completed for any reason, including an error or missing command permission:

1. Explain the blocker with `tq comment add <issue_id> --author claude-code --body "<comment>"`.
2. Change the issue status with `tq issue update <issue_id> --status blocked`.
3. Preserve the task worktree for follow-up.

# Output Protocol

> [!NOTE]
> `tq issue watch` is experimental and intentionally isolated from the rest of the CLI.

`tq issue watch` ignores the global `--output` flag and always writes one JSON object per line. Each stdout line becomes one Monitor notification.

- `{"type":"error","body":"<message>"}` — a transient polling failure; the loop continues.
- `{"type":"event","eventType":"issue-ready","body":<issue>}` — a ready issue; `body` contains the full issue payload.
- `{"type":"info","body":"<message>"}` — startup configuration or a polling summary; emitted only with `--verbose`.

Flags:

- `--interval <seconds>`: polling interval (default: 30).
- `--seen-ttl <seconds>`: suppress duplicate issue events for this duration (default: 900; must be greater than `--interval`).
- `--verbose`: emit info envelopes.

The loop runs until stopped by SIGINT, kill, or the Monitor timeout; it does not exit on its own.
