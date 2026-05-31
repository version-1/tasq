# Symphony TODO

This document tracks Symphony-related implementation work that is intentionally not complete yet.

## Orchestrator

- Replace the simulated runner with a real Codex app-server runner.
- Define and implement the exact Codex app-server JSON-RPC contract.
- Add runner process lifecycle management, cancellation, and timeout handling.
- Add long-running lease renewal for active agent runs.
- Persist enough runner progress and logs for debugging and recovery.
- Define retry limits and manual intervention thresholds for failed runs.

## Workspace

- Populate per-issue workspaces from the repository instead of only creating directories.
- Define workspace cleanup and retention policy.
- Add terminal cleanup for completed, failed, and cancelled runs.
- Preserve enough workspace metadata to support recovery and debugging.

## Workflow Contract

- Decide whether `WORKFLOW.md` should be reloaded dynamically or only at process startup.
- Expand workflow parsing if the project needs full YAML front matter compatibility.
- Document supported Tasq-specific workflow fields once the contract stabilizes.

## Tracker Integration

- Keep using the local issue-tracker API as the tracker adapter.
- Do not implement a Linear tracker client unless this product decision changes.
- Define any future external tracker sync behind issue-tracker or a provider boundary.

## Observability

- Decide whether run logs live in SQLite or filesystem artifacts.
- Add operator-facing visibility for stuck leases, outbox retries, and failed workspace setup.
