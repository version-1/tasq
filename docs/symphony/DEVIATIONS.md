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

Tasq's `WORKFLOW.md` front matter is intentionally a smaller, Tasq-specific contract layered on top
of the Symphony front matter schema. The canonical Tasq workflow contract is documented in
[WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md). The table below maps the Symphony `SPEC.md` fields to
Tasq front matter behavior.

| Top-level key | Child field | `SPEC.md` support | Tasq support | Tasq front matter / behavior |
| --- | --- | --- | --- | --- |
| `tracker` | `kind` | Core field; required for Symphony dispatch. | Parsed, not used for dispatch. | Current orchestration reads from the local issue-tracker API instead of a Linear client. |
| `tracker` | `endpoint` | Core field. | Parsed, not used by the local tracker path. | Parsed for partial compatibility with the Symphony tracker shape. |
| `tracker` | `api_key` | Core field; required for Linear dispatch after `$VAR` resolution. | Parsed, not used by the local tracker path. | Supports `$VAR` resolution. |
| `tracker` | `project_slug` | Core field; required when `tracker.kind` is `linear`. | Parsed, not used by the local tracker path. | Required only when `tracker.kind` is set. |
| `tracker` | `active_states` | Core field. | Parsed, not authoritative. | Tasq dispatch eligibility is driven by local issue-tracker states. |
| `tracker` | `terminal_states` | Core field. | Parsed, not authoritative. | Tasq cleanup/reconciliation uses local issue-tracker states. |
| `polling` | `interval_ms` | Core field. | Supported. | Orchestrator polling interval. |
| `workspace` | `root` | Core field. | Supported with Tasq constraint. | Resolved relative to the selected `WORKFLOW.md`; must be inside the target Git repository for worktree management. |
| `workspace` | `source` | Not in the current core table; referenced by workspace population designs. | Not supported. | Tasq creates Git worktrees under `workspace.root` instead. |
| `hooks` | `after_create` | Core field. | Supported. | Runs through `bash -lc` in the issue workspace only for newly created workspaces. |
| `hooks` | `before_run` | Core field. | Supported. | Runs before each agent attempt. |
| `hooks` | `after_run` | Core field. | Supported. | Runs after each agent attempt; failures are logged and ignored. |
| `hooks` | `before_remove` | Core field. | Supported. | Runs before workspace cleanup when the workspace directory exists; failures are logged and cleanup continues. |
| `hooks` | `timeout_ms` | Core field. | Supported. | Applies to all workspace hooks and must be positive. |
| `agent` | `max_concurrent_agents` | Core field. | Supported. | Global concurrent run limit. |
| `agent` | `max_turns` | Core field. | Supported with Tasq gate. | Maximum Codex turns for one run; continuation turns still require `agent.continuation_turns_enabled`. |
| `agent` | `max_retry_backoff_ms` | Core field. | Supported. | Retry backoff cap. |
| `agent` | `max_concurrent_agents_by_state` | Core field. | Parsed. | Normalized map; usefulness depends on the local issue-tracker state model. |
| `agent` | `continuation_turns_enabled` | Not supported by `SPEC.md` core schema. | Tasq extension. | Gates continuation turns even when `agent.max_turns` is greater than one. |
| `agent` | `max_retry_attempts` | Not supported by `SPEC.md` core schema. | Tasq extension. | Sets the number of run retry attempts. |
| `codex` | `command` | Core field. | Supported. | Repository root `WORKFLOW.md` uses `codex --sandbox workspace-write app-server`. |
| `codex` | `approval_policy` | Core Codex pass-through field. | Parsed, optional. | Repository root `WORKFLOW.md` does not set it. |
| `codex` | `thread_sandbox` | Core Codex pass-through field. | Parsed, optional. | Repository root `WORKFLOW.md` does not set it. |
| `codex` | `turn_sandbox_policy` | Core Codex pass-through field. | Parsed, optional. | Repository root `WORKFLOW.md` does not set it. |
| `codex` | `turn_timeout_ms` | Core field. | Supported. | Codex turn timeout. |
| `codex` | `read_timeout_ms` | Core field. | Supported. | Codex app-server read timeout. |
| `codex` | `stall_timeout_ms` | Core field; `<= 0` disables stall detection in Symphony. | Supported with stricter validation. | Tasq validates it as positive. |
| `server` | `port` | Supported extension field. | Supported extension. | Optional HTTP extension port; CLI `--port` can override it. |
| `tasq` | `task_work_prompt` | Not supported by `SPEC.md` core schema. | Tasq extension. | Controls whether the orchestrator prepends default `tq` issue-tracker synchronization instructions to the rendered prompt. |

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
