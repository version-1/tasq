# Tasq Symphony Deviations

This document records intentional differences between Tasq and the upstream Symphony service specification in [SPEC.md](SPEC.md).

Tasq uses the Symphony specification as the direction for orchestration, workspace, workflow, agent-runner, tracker, and observability behavior. Some implementation choices differ because Tasq already has a local issue-tracker service that owns issue state.

## Tracker Adapter

Tasq does not implement the Linear tracker client described in Section 11 of the Symphony specification.

Instead, Tasq treats its local issue-tracker API as the tracker adapter boundary:

- The issue-tracker owns issue state, project data, and workspace records.
- The issue-tracker exposes issue listing and issue-state query endpoints for tracker adapter reads.
- The orchestrator keeps historical run and runner-event data in its own SQLite store.

This keeps external tracker integrations out of the orchestrator and preserves the repository's existing service boundary.

## Current Implementation Gap

The current orchestrator is moving toward Symphony conformance incrementally. It is not yet a complete Symphony implementation.

Implemented or in progress:

- Workflow file loading with a small supported subset of Symphony front matter.
- `WORKFLOW.md` is loaded at process startup only; runtime reload is intentionally deferred.
- Workspace root resolution and sanitized per-issue workspace directories.
- A runner interface with both simulated and Codex app-server subprocess implementations.
- SQLite runner event logging and workspace metadata records.
- Config-gated continuation turns on a live Codex app-server thread.
- In-process retry scheduling with capped exponential backoff.
- Active-run reconciliation for terminal/non-active issue states and stall handling.
- Repository-source workspace population on first workspace creation.
- Terminal and failed/cancelled workspace cleanup with cleanup metadata.
- Operator-facing logs for workspace setup failures.

Not yet implemented:

- Dynamic `WORKFLOW.md` reload.
- Strict prompt rendering with full variable and filter checking.
- Token/rate-limit accounting.
- Full optional Symphony HTTP status/API surface.

Tasq intentionally keeps workflow front matter parsing to the supported Tasq-specific subset
documented in [WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md). Full generic YAML compatibility is not a
current product requirement.

## Compatibility Notes

Where Symphony says "tracker" for scheduling, Tasq's orchestrator should read that as the local issue-tracker API unless a future design explicitly adds an external tracker adapter.

Where Symphony describes Linear-specific query semantics, Tasq records those requirements as not applicable to the orchestrator while the local issue-tracker boundary remains the selected tracker adapter.
