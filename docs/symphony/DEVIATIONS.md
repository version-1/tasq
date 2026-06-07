# Tasq Symphony Deviations

This document records intentional differences between Tasq and the upstream Symphony service specification in [SPEC.md](SPEC.md).

Tasq uses the Symphony specification as the direction for orchestration, workspace, workflow, agent-runner, tracker, and observability behavior. Some implementation choices differ because Tasq already has a local issue-tracker service that owns issue state.

## Tracker Adapter

Tasq does not implement the Linear tracker client described in Section 11 of the Symphony specification.

Instead, Tasq treats its local issue-tracker API as the tracker adapter boundary:

- The issue-tracker owns issue state and project data.
- The orchestrator owns workspace records, metadata, and lifecycle behavior.
- The issue-tracker exposes issue listing and issue-state query endpoints for tracker adapter reads.
- The orchestrator keeps historical run and runner-event data in its own SQLite store.

This keeps external tracker integrations out of the orchestrator and preserves the repository's existing service boundary.

## Workflow Front Matter Contract

Tasq's `WORKFLOW.md` front matter is intentionally a smaller, Tasq-specific subset of the Symphony
front matter schema. The canonical Tasq workflow contract is documented in
[WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md).

Supported Tasq fields:

- `tasq.task_work_prompt` is a Tasq extension that controls whether the orchestrator prepends the
  default `tq` issue-tracker synchronization instructions to the rendered prompt.
- `polling.interval_ms`, `workspace.root`, `agent.max_concurrent_agents`, `agent.max_turns`,
  `agent.continuation_turns_enabled`, `agent.max_retry_attempts`,
  `agent.max_retry_backoff_ms`, `codex.command`, `codex.read_timeout_ms`,
  `codex.turn_timeout_ms`, `codex.stall_timeout_ms`, `server.port`, and workspace lifecycle
  `hooks` are supported.
- `hooks.after_create`, `hooks.before_run`, `hooks.after_run`, `hooks.before_remove`, and
  `hooks.timeout_ms` are supported. Hook scripts run through `bash -lc` with the issue workspace as
  the working directory.
- `tracker.*` fields are still parsed for partial Symphony compatibility, but the current Tasq
  orchestrator reads work from the local issue-tracker API instead of using the Linear tracker
  client described by Symphony.
- Codex pass-through fields (`codex.approval_policy`, `codex.thread_sandbox`, and
  `codex.turn_sandbox_policy`) are parsed but are not part of the root `WORKFLOW.md` template used
  by this repository today.

Intentional differences from Symphony:

- `WORKFLOW.md` is loaded once at orchestrator process startup. Dynamic watch/reload and runtime
  re-application of changed settings are deferred.
- Unknown front matter fields are ignored for forward compatibility.
- `workspace.source` is not supported. Tasq creates issue workspaces with Git worktrees under
  `workspace.root`.
- Large transcript artifact paths and observability sinks are not configurable through workflow
  front matter; Tasq records runner progress, workspace metadata, workspace setup failures, and
  cleanup state in the orchestrator SQLite database.
- The local `tq project check` command validates the front matter fields required by Tasq's default
  project template rather than the full Symphony schema.

## Workspace Key

Symphony derives workspace keys from `issue.identifier`, such as `MT-649`.
Tasq intentionally uses `issue-<ID>` as its canonical issue identifier for workspace naming, for
example `issue-42`.

Rationale:

- Tasq's local issue-tracker owns issues with stable numeric IDs.
- Tasq does not currently model a separate human-readable external tracker identifier.
- Using `issue-<ID>` keeps workspace paths deterministic without adding Linear-specific fields to
  the local issue contract.

The same workspace key convention is used for workspace metadata, startup terminal cleanup, and
active-run reconciliation.

## Current Implementation Gap

The current orchestrator is moving toward Symphony conformance incrementally. It is not yet a complete Symphony implementation.

Implemented or in progress:

- Workflow file loading with a small supported subset of Symphony front matter.
- `WORKFLOW.md` is loaded at process startup only; runtime reload is intentionally deferred.
- Workspace root resolution and sanitized per-issue workspace directories.
- Workspace lifecycle hooks with `hooks.timeout_ms`.
- A runner interface with both simulated and Codex app-server subprocess implementations.
- SQLite runner event logging and workspace metadata records.
- Config-gated continuation turns on a live Codex app-server thread.
- In-process retry scheduling with capped exponential backoff.
- Active-run reconciliation for terminal/non-active issue states and stall handling.
- Git worktree workspace creation on first workspace creation.
- Terminal and failed/cancelled workspace cleanup with cleanup metadata.
- Operator-facing logs for workspace setup failures.

Not yet implemented:

- Dynamic `WORKFLOW.md` reload.
- Strict prompt rendering with full variable and filter checking.
- Token/rate-limit accounting.
- Full optional Symphony HTTP status/API surface.

Tasq supports the workflow front matter fields documented in
[WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md). Unknown fields are ignored for forward
compatibility.

## Workspace Creation Strategy

Tasq creates per-issue orchestrator workspaces with `git worktree add` instead of copying files from a
repository source directory.

Rationale:

- The coding agent needs the per-issue workspace to be a Git repository root.
- Copying files without `.git` makes repository inspection commands observe a different Git root and
  can cause edits to target the parent repository instead of the issue workspace.
- A worktree preserves Git metadata while keeping the workspace path deterministic under
  `workspace.root`.

The `workspace.source` workflow field is intentionally not supported. `workspace.root` must be
inside the target Git repository so the workspace manager can resolve the repository with
`git rev-parse --show-toplevel`.

Workspace branches use `agent/<workspace-key>`, for example `agent/issue-42`. Cleanup uses
`git worktree remove --force`, deletes the corresponding local branch best-effort, and runs
`git worktree prune` on orchestrator startup.

## Compatibility Notes

Where Symphony says "tracker" for scheduling, Tasq's orchestrator should read that as the local issue-tracker API unless a future design explicitly adds an external tracker adapter.

Where Symphony describes Linear-specific query semantics, Tasq records those requirements as not applicable to the orchestrator while the local issue-tracker boundary remains the selected tracker adapter.
