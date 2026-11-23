# Shared orchestration workflow

Read this after the service-status prerequisite in `SKILL.md` passes. Use the selected CLI everywhere; `<cli>` below is a placeholder, not a shell variable. Use the comment author from the environment reference.

## Dispatch

1. Reuse a polling interval already specified in the conversation; otherwise ask for seconds (default 30). Read the project's `WORKFLOW.md` and the applicable `tasq-cli` resource references.
2. Start `<cli> issue watch --interval <seconds>` using the environment's managed monitoring procedure. Keep monitoring until the user's stop request or a reported blocker; an empty queue is not completion.
3. On each `issue-ready` event, fetch `<cli> issue get <id>`. Recheck the service prerequisite, current `ready` status, dependency-aware queue eligibility, and existing workers before delegating. Do not promote backlog issues automatically.
4. Track only running and pending IDs. Suppress duplicates across reads and watcher restarts, retain pending IDs while capacity is full, and recheck eligibility before using a free slot. Release an ID after worker results and tracker updates are handled, or when it becomes ineligible. A later eligible ready event may be delegated again.
5. Delegate with the issue description and project workflow. After successful hand-off, set `in_progress` and add a progress comment naming the worker. Reserve or release capacity for required review; a worker waiting for a reviewer can yield its turn and resume after review.
6. Process worker completions while watching. Preserve failures and follow-up notes in the tracker. Do not promise unattended event handling after ending the monitoring turn.

## Fallback delegated task workflow

Use this only for behavior not specified by the project workflow.

1. Confirm scope and fetch origin. Use the base specified by the issue or the target repository's branch rules, falling back to the repository's default branch. Create a fresh isolated branch/worktree from that base and target it in the PR. Follow the repository's naming rules, with `<project_root>/.worktrees/tasq/<id>-<summary>` as the fallback. Record the branch and path in a progress comment.
2. Make focused changes and run appropriate checks. Read component instructions and preserve unrelated changes. Follow repository rules for commits, base updates, review, push, and PR creation.
3. After the PR exists, record its title, URL, verification and review results in a hand-off comment; register `<cli> artifact set <id> --type pull_request <url>` and set the issue to `review`.
4. Do not create a change or empty PR just to satisfy this flow when investigation finds nothing to fix. Record scope, evidence and checks, and report the outcome; use the user's or project's agreed no-change completion policy.
5. Remove only owned task worktrees/branches, only after verifying the PR exists and the issue is in review, and only when repository safety checks permit removal. Otherwise preserve them; do not force-delete an unmerged branch.
6. If blocked, add a blocker comment, set `blocked`, and preserve the worktree. Do not interpret a service conflict affecting this orchestrator as failure of every delegated issue.

Do not add standalone reports or other deliverables beyond the issue's scope unless requested or required by the project. Keep investigation evidence in tracker comments when that satisfies the task.

## Event protocol and recovery

`issue watch` ignores `--output` and emits NDJSON:

- `type=event`, `eventType=issue-ready`: the full issue is in `body`; use `body.id`.
- `type=error`: transient error text in `body`; report it and let the watcher retry.
- `type=info`: informational text, emitted with `--verbose` only.

Parse each complete line once and buffer partial lines. Do not parse stderr as events. Set `--seen-ttl` strictly above the polling interval (default 900 seconds).

If the watcher exits unexpectedly, report it and restart only for recoverable causes while monitoring remains requested, after rechecking the service prerequisite. Keep running/pending IDs across restart and avoid tight retries. Permission/configuration errors require a blocker report rather than repeated restarts.

## Stop

On `stop` or equivalent cancellation, stop dispatching immediately, terminate only the owned watcher through its managing tool, and verify exit. Preserve unread event evidence as needed. Report any running workers and their issue/worktree paths; watcher cancellation does not itself authorize cancelling workers or deleting work.
