# Symphony TODO

This document tracks Symphony-related implementation work that is intentionally not complete yet.

## Orchestrator

- Add continuation turns on a live Codex app-server thread.
- Add retry scheduling, retry limits, and manual intervention thresholds for failed runs.
- Add active-run reconciliation and stall handling for long-running Codex turns.

## Workspace

- Populate per-issue workspaces from the repository instead of only creating directories.
- Define workspace cleanup and retention policy.
- Add terminal cleanup for completed, failed, and cancelled runs.
- Expand workspace metadata with repository population and cleanup state.

## Workflow Contract

- Expand workflow parsing if the project needs full YAML front matter compatibility.

## Tracker Integration

- Keep using the local issue-tracker API as the tracker adapter.
- Do not implement a Linear tracker client unless this product decision changes.
- Define any future external tracker sync behind issue-tracker or a provider boundary.

## Observability

- Add operator-facing visibility for stuck leases, outbox retries, and failed workspace setup.
- Decide whether large Codex transcript artifacts should stay in SQLite or move to filesystem artifacts.
