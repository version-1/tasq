# Orchestrator Packages

The orchestrator owns agent run state, runner events, workspace setup metadata, and optional local runtime inspection.

`cmd/orchestrator` is the composition root. It wires the packages in this directory together and should stay responsible for process-level concerns such as flags, signal handling, and startup logging.

## Package Boundaries

### `run`

Defines the orchestrator run domain model.

Responsibilities:

- Define run status values.
- Define persisted run records.

This package should not depend on storage, HTTP clients, or runner implementations.

### `runstore`

Persists orchestrator-owned state in SQLite.

Responsibilities:

- Open and migrate the orchestrator database.
- Create run records.
- Update run status.
- Record runner events and workspace metadata.
- Query active runs and issue-specific run details for the HTTP API.

This package depends on `run` for domain types. It should not call the issue-tracker API or run agents.

### `tracker`

Adapts the local issue-tracker HTTP API for orchestrator use.

Responsibilities:

- Fetch issues and issue states.
- Decode the standard API envelope used by issue-tracker.
- Surface issue-tracker error responses with their error code and message.

This package is the tracker boundary for the orchestrator. Tasq currently uses the local issue-tracker API here and does not implement a Linear client.

### `workflow`

Loads the repository workflow contract.

Responsibilities:

- Load `WORKFLOW.md`.
- Parse supported front matter fields.
- Provide orchestration configuration defaults.
- Resolve workspace root paths relative to the workflow file.

This package owns configuration parsing only. It should not start workers, create workspaces, or call external APIs.

### `workspace`

Creates and validates orchestrator workspaces.

Responsibilities:

- Manage the configured workspace root.
- Create sanitized per-issue workspace directories.
- Populate newly created workspaces from the configured source.
- Remove terminal, failed, and cancelled workspaces when cleanup is requested.
- Prevent workspace paths from escaping the configured root.

Workspace population and cleanup state are persisted through `runstore` metadata.

### `runner`

Defines the agent runner boundary.

Responsibilities:

- Define the runner interface.
- Define runner task input and result output.
- Provide a simulated runner for tests and a Codex app-server subprocess runner for real execution.

The Codex runner follows the contract documented in [../../docs/symphony/CODEX_APP_SERVER.md](../../docs/symphony/CODEX_APP_SERVER.md).

## Dependency Direction

Keep dependencies flowing inward to small contracts and outward only from coordinators.

```text
cmd/orchestrator
  ├─ httpserver ── runstore ── run
  └─ workflow
```

Rules:

- `run` should remain dependency-light.
- `runstore`, `tracker`, `workflow`, `workspace`, `runner`, and `httpserver` should remain focused on their own contracts.
- `cmd/orchestrator` should wire packages together instead of hiding construction behind a broad root package.
