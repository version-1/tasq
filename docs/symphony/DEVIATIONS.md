# Tasq Symphony Deviations

This document records intentional differences between Tasq and the upstream Symphony service specification in [SPEC.md](SPEC.md).

Tasq uses the Symphony specification as the direction for orchestration, workspace, workflow, agent-runner, tracker, and observability behavior. Some implementation choices differ because Tasq already has a local issue-tracker service that owns issue state and exposes a work queue for orchestrators.

## Tracker Adapter

Tasq does not implement the Linear tracker client described in Section 11 of the Symphony specification.

Instead, Tasq treats its local issue-tracker API as the tracker adapter boundary:

- The issue-tracker owns issue state, work item eligibility, and lease-backed work item claims.
- The orchestrator polls `POST /api/v1/work-items/claim` instead of querying Linear directly.
- The orchestrator sends run facts to `POST /api/v1/orchestrator-events`.
- The issue-tracker applies issue status transitions from accepted run facts.

This keeps external tracker integrations out of the orchestrator and preserves the repository's existing service boundary.

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
- In-memory running and claimed issue tracking inside the worker.
- Existing issue-tracker work queue polling and durable outbox delivery.
- Claim-token-based lease renewal for active work item runs.
- SQLite runner event logging and workspace metadata records.
- Config-gated continuation turns on a live Codex app-server thread.
- In-process retry scheduling with capped exponential backoff.
- Active-run reconciliation for terminal/non-active issue states and stall handling.
- Repository-source workspace population on first workspace creation.
- Terminal and failed/cancelled workspace cleanup with cleanup metadata.
- Operator-facing logs for unsent outbox events and workspace setup failures.

Not yet implemented:

- Dynamic `WORKFLOW.md` reload.
- Strict prompt rendering with full variable and filter checking.
- Token/rate-limit accounting.
- Optional Symphony HTTP status/API surface.

Tasq supports the workflow front matter fields documented in
[WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md). Unknown fields are ignored for forward
compatibility.

## Compatibility Notes

Where Symphony says "tracker" for scheduling, Tasq's orchestrator should read that as the local issue-tracker API unless a future design explicitly adds an external tracker adapter.

Where Symphony describes Linear-specific query semantics, Tasq records those requirements as not applicable to the orchestrator while the local issue-tracker boundary remains the selected tracker adapter.
