# Tasq Symphony Deviations

This document records intentional differences between Tasq and the upstream Symphony service specification in [SPEC.md](SPEC.md).

Tasq uses the Symphony specification as the direction for orchestration, workspace, workflow, agent-runner, tracker, and observability behavior. Some implementation choices differ because Tasq already has a local issue-tracker service that owns issue state.

## Tracker Adapter

Tasq does not implement the Linear tracker client described in Section 11 of the Symphony specification.

Instead, Tasq treats its local issue-tracker API as the tracker adapter boundary:

- The issue-tracker owns issue state and project data.
- The issue-tracker owns issue dependency edges and computes derived queue state.
- The orchestrator owns workspace records, metadata, and lifecycle behavior.
- The issue-tracker exposes issue listing, issue-state query, and queue endpoints for tracker adapter reads.
- The orchestrator keeps historical run and runner-event data in its own SQLite store.

This keeps external tracker integrations out of the orchestrator and preserves the repository's existing service boundary.

Tasq's dispatch queue differs from Symphony worker scheduling by keeping `queued` and `pending` as derived issue-tracker API response states instead of persisted orchestrator-owned issue states. The orchestrator polls `GET /api/v1/queue`, dispatches only the `queued` array, and ignores `pending` issues until the issue-tracker classifies them as queued after dependency status changes.

Tasq also accepts initial issue dependencies at the local issue-tracker creation boundary:
`POST /api/v1/issues` may include `dependency_ids` when creating an issue. Symphony models
dependencies as normalized tracker data (`blocked_by`) derived from tracker relations and does not
define a local issue creation write contract. Tasq keeps this write-side dependency contract in the
issue-tracker service so dependency validation, cycle prevention, and derived queue classification
remain owned by the same boundary.

## Workflow Front Matter Contract

Tasq's `WORKFLOW.md` front matter is intentionally a smaller, Tasq-specific contract layered on top
of the Symphony front matter schema. [WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md) is the canonical
source for every supported field, validation rule, and runtime behavior.

This table records only differences from Symphony: Tasq extensions and fields that Tasq does not
support.

| Field | Symphony support | Tasq support | Difference |
| --- | --- | --- | --- |
| `workspace.source` | ✓ | ✗ | Tasq creates Git worktrees under `workspace.root` instead. |
| `agent.continuation_turns_enabled` | ✗ | ✓ | Tasq extension. |
| `agent.max_retry_attempts` | ✗ | ✓ | Tasq extension. |
| `tasq.task_work_prompt` | ✗ | ✓ | Tasq extension. |

Intentional differences from Symphony:

- The effective `WORKFLOW.md` is resolved per project when an issue is queued or dispatched.
  Dynamic watch/reload of already-running work is deferred.
- Workflow path selection does not use a process-level explicit workflow path or cwd default.
  Tasq resolves the effective workflow per project; see [Workflow Configuration](../site/docs/guides/workflow-configuration.md).
- Codex app-server orchestration is transport-neutral internally, and Tasq includes stdio and
  websocket transport packages. Production workflow execution still starts the stdio subprocess
  transport; runtime transport selection and real Codex websocket server integration are deferred.

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
- Effective workflow resolution happens per project at queue/dispatch time; runtime reload for
  already-running work is intentionally deferred.
- Workspace root resolution and sanitized per-issue workspace directories.
- Workspace lifecycle hooks with `hooks.timeout_ms`.
- A runner interface with both simulated and Codex app-server subprocess implementations.
- A Codex app-server transport contract with stdio and websocket transport implementations.
- SQLite runner event logging and workspace metadata records.
- Config-gated continuation turns on a live Codex app-server thread.
- Retry and resume across worker runs by starting a fresh app-server subprocess and reconnecting to
  persisted thread state.
- In-process retry scheduling with capped exponential backoff.
- Active-run reconciliation for terminal/non-active issue states and stall handling.
- Git worktree workspace creation on first workspace creation.
- Terminal and failed/cancelled workspace cleanup with cleanup metadata, including stale
  thread/rollout artifact cleanup.
- Operator-facing logs for workspace setup failures.

Resolved implementation gaps:

- The earlier Codex runner implementation created app-server threads with `ephemeral: true`, which
  prevented later `threadId` resume because the thread was not materialized on disk. Tasq now creates
  persistent Codex threads with `ephemeral: false`, persists the returned `thread_id`, and resumes
  eligible non-terminal retries with `thread/resume`.

Not yet implemented:

- Dynamic `WORKFLOW.md` reload.
- Strict prompt rendering with full variable and filter checking.
- Token/rate-limit accounting.
- Runtime selection between stdio and websocket Codex app-server transports.
- Integration verification against a real Codex websocket app-server.
- Full optional Symphony HTTP status/API surface.

## Workspace Creation Strategy

Tasq creates per-issue orchestrator workspaces with `git worktree add` instead of copying files from a
repository source directory.

Rationale:

- The coding agent needs the per-issue workspace to be a Git repository root.
- Copying files without `.git` makes repository inspection commands observe a different Git root and
  can cause edits to target the parent repository instead of the issue workspace.
- A worktree preserves Git metadata while keeping the workspace path deterministic under the
  issue project's `Project.Location` plus the configured relative `workspace.root`.

The `workspace.source` workflow field is intentionally not supported. `workspace.root` must be
inside the orchestrator project's Git repository so the workspace manager can resolve the relative
workspace-root suffix. For each dispatched issue, Tasq resolves `Issue.ProjectID` through the local
issue-tracker API and creates the worktree under the referenced project's location using that same
relative suffix, for example `<Project.Location>/.worktrees/agents/issue-42`.

Workspace branches use `agent/<workspace-key>`, for example `agent/issue-42`. Cleanup uses
`git worktree remove --force` and deletes the corresponding local branch best-effort.

## Codex Thread Resume Lifecycle

Symphony describes continuation turns on a live Codex app-server thread inside one worker run. Tasq
extends that lifecycle by persisting enough thread state to resume eligible non-terminal work across
separate worker runs.

When Tasq resumes a previous worker run, it starts a new app-server subprocess, reconnects using the
persisted `thread_id`, keeps the same workspace `cwd`, and sends continuation guidance instead of the
original issue prompt. The previous app-server subprocess is not treated as an issue-lifetime
process; worker exit always closes it.

Persisted `thread_id` values are reusable only while the issue remains non-terminal. Terminal issue
cleanup removes workspace-scoped coding-agent artifacts, including persisted thread/rollout state and
orchestrator-owned resume pointers. If the same issue is later reopened or recreated as active work,
dispatch starts without stale thread state.

This deviates from the following SPEC.md sections:

| Section | Symphony assumption | Tasq behavior |
| --- | --- | --- |
| §7.7, §8.1 | Terminal cleanup focuses on stale workspace directories | Terminal cleanup also removes thread/rollout artifacts and invalidates resume pointers |
| §10.2–10.3 | Continuation state is scoped to one live app-server subprocess inside a worker run | Retry/resume across worker runs starts a fresh subprocess and reconnects through persisted thread state |
| §14.3 | Restart recovery re-dispatches eligible work after workspace cleanup | Active issues may resume persisted Codex thread state; terminal issues are cleaned so they cannot resume stale state |
| §17.1, §17.6, §18 | Conformance covers workspace cleanup and live-thread continuation | Tasq additionally verifies cross-worker resume and terminal thread/rollout cleanup |

## Workflow Path Selection

Symphony defines a process-level workflow path selection model: the runtime can receive an explicit
workflow path, otherwise it defaults to `WORKFLOW.md` in the current process working directory. Tasq
does not expose the Symphony workflow-path flag, including the earlier `--workflow` flag shape.

Tasq resolves workflow configuration per project instead of per orchestrator process. The
orchestrator process cwd is not the default source of workflow behavior for dispatched issues.
The cwd still matters for operator commands and process startup, but issue dispatch uses the
project associated with the issue. For the effective-workflow source precedence, see
[Workflow Configuration](../site/docs/guides/workflow-configuration.md).

This deviates from the following SPEC.md sections:

| Section | Symphony assumption | Tasq behavior |
| --- | --- | --- |
| §5.2, §6.1 | Explicit runtime workflow path, otherwise cwd `WORKFLOW.md` | No `--workflow` path selection; resolve workflow per project |
| §17.7 | CLI accepts a positional workflow path and falls back to `./WORKFLOW.md` | Orchestrator CLI does not select one workflow file for all projects |
| §18 | Workflow path selection supports explicit runtime path and cwd default | Effective workflow is resolved at issue dispatch time |

## Project Deletion Cleanup

Symphony does not define a project deletion API. Tasq implements project deletion in the
issue-tracker API and performs best-effort cross-store cleanup of orchestrator runtime records for
the deleted project's issues.

The selected Tasq behavior is:

- The issue-tracker reads the project-owned issue IDs, asks the orchestrator runstore to delete
  `runner_events`, `workspace_setup_failures`, `workspace_metadata`, and `runs` for those issue IDs,
  then deletes the issue-tracker project descendants.
- The orchestrator runstore checks for `running` runs and deletes descendants in one SQLite
  transaction. If any target run is `running`, the delete returns `409 Conflict` and does not delete
  issue-tracker or orchestrator records.
- The cleanup order is intentionally orchestrator-first because the issue IDs are the cross-store
  join key. If orchestrator cleanup succeeds but issue-tracker deletion later fails, retrying the
  project deletion is idempotent for orchestrator records and can still delete the issue-tracker
  records.
- Tasq does not introduce a cross-store lock. A new run can theoretically be created after the
  running check and before issue-tracker deletion. The current orchestrator dispatch path reads
  eligible issues from the issue-tracker, so once issue-tracker deletion commits there is no durable
  issue for future dispatch; the narrow check-then-delete race is accepted as an implementation
  tradeoff rather than a distributed transaction.
- Project deletion removes persisted workspace metadata and setup failure records, but it does not
  delete workspace directories. Workspace filesystem cleanup remains owned by the workspace manager's
  normal lifecycle.

## Multi-Project Orchestration

Symphony assumes one process serves one project (one `WORKFLOW.md`, one `tracker.project_slug`,
one `workspace.root`). Tasq runs a single orchestrator process that serves multiple projects.

The Symphony model therefore treats workflow configuration as a single file selected for the
orchestrator instance. Tasq treats workflow configuration as project data: multiple projects can
have independent workflows, and the orchestrator resolves the relevant workflow only when it
dispatches a specific issue.

The orchestrator itself is project-unaware. It polls the local issue-tracker for all eligible
issues across projects in a single call, dispatches them through the same concurrency pool, and
manages runtime state (`claimed`, `running`, `retry_attempts`) in flat maps keyed by issue ID
without project partitioning.

Project-specific behavior is resolved per-issue at dispatch time:

- **Workspace path**: resolved via `Issue.ProjectID → Project.Location` plus the relative
  `workspace.root` suffix (documented in "Workspace Creation Strategy" above).
- **Prompt and hooks**: each project owns its own `WORKFLOW.md`. The orchestrator loads the
  project-local workflow file when building the prompt and executing hooks for a given issue.
- **Polling**: a single poll tick fetches candidates from all projects. Per-project
  `polling.interval_ms` is not supported; the orchestrator uses one global interval.
- **Concurrency**: `agent.max_concurrent_agents` is a global limit across all projects.
  Per-project concurrency limits are not currently supported.

This deviates from the following SPEC.md sections:

| Section | Symphony assumption | Tasq behavior |
| --- | --- | --- |
| §5.1–5.2 | One `WORKFLOW.md` per process | One `WORKFLOW.md` per project, resolved per-issue |
| §5.3.1 | `tracker.project_slug` (singular) | Not used; issue-tracker returns issues across projects |
| §4.1.8, §8.3 | Runtime state and concurrency are implicitly single-project scoped | Flat global state; no per-project partitioning |
| §8.1 | One poll interval from one workflow | One global poll interval; per-project intervals ignored |

## Compatibility Notes

Where Symphony says "tracker" for scheduling, Tasq's orchestrator should read that as the local issue-tracker API unless a future design explicitly adds an external tracker adapter.

Where Symphony describes Linear-specific query semantics, Tasq records those requirements as not applicable to the orchestrator while the local issue-tracker boundary remains the selected tracker adapter.
