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

## Schema Contract

Tasq's persisted schema is intentionally not a direct implementation of the Symphony normalized
domain model. The canonical Tasq field contract is documented in
[../design/schema.md](../design/schema.md).

Issue-tracker differences:

- Tasq issues use a local numeric `id` as the stable tracker ID and expose `issue-<ID>` as the
  orchestrator-facing identifier convention.
- Issues are owned by a local `project_id` and expose `project_key` from the referenced Project.
  Symphony's `tracker.project_slug` is a query scope for an external tracker, not a persisted
  Project entity.
- Issue `description` is stored as a non-null Markdown string with `""` as the empty value, rather
  than a nullable string.
- Issue `status` is a Tasq enum (`backlog`, `ready`, `in_progress`, `review`, `done`, `blocked`,
  `failed`) instead of arbitrary tracker state names such as Linear workflow states.
- Issue `priority` is a Tasq enum (`low`, `normal`, `high`, `urgent`) instead of Symphony's
  normalized integer-or-null priority.
- Tasq does not persist Symphony issue fields such as `branch_name`, `url`, `labels`, or
  `blocked_by` in the issue row. Blocking and PR metadata can be represented in issue text,
  comments, workflow policy, or future first-class fields when needed.
- Comments and image attachments are first-class issue-tracker records. They are Tasq API features,
  not Symphony core-domain entities.

Orchestrator persistence differences:

- Symphony treats scheduling state as authoritative in memory and does not require a persistent
  database for restart recovery. Tasq additionally persists run attempts, runner events, workspace
  metadata, and workspace setup failures in SQLite for operator visibility and local recovery.
- Tasq run records store `run_id`, numeric `issue_id`, `status`, `workspace`, `attempt`, `error`,
  and `orchestrator_id`. They are a persisted audit and reconciliation surface for run attempts,
  not the full live session state described by Symphony.
- Tasq runner events store `event_type`, `message`, optional JSON payload, and occurrence time as an
  append-only observability log.
- Tasq workspace metadata records include cleanup bookkeeping fields managed by store methods. These
  fields support worktree lifecycle cleanup and are not exposed as workflow input contract fields.

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
