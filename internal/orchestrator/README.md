# Orchestrator Packages

The orchestrator owns agent run state, work assignment state, workspace setup, and durable delivery of run events to the issue-tracker.

`cmd/orchestrator` is the composition root. It wires the packages in this directory together and should stay responsible for process-level concerns such as flags, signal handling, and startup logging.

## Package Boundaries

### `run`

Defines the orchestrator run domain model.

Responsibilities:

- Define run status values.
- Define persisted run records.
- Define outbox event records.

This package should not depend on storage, HTTP clients, workers, or runner implementations.

### `runstore`

Persists orchestrator-owned state in SQLite.

Responsibilities:

- Open and migrate the orchestrator database.
- Create run records.
- Update run status.
- Enqueue outbox events whenever run state changes.
- List unsent outbox events.
- Mark outbox events as sent.

This package depends on `run` for domain types. It should not call the issue-tracker API or run agents.

### `tracker`

Adapts the local issue-tracker HTTP API for orchestrator use.

Responsibilities:

- Claim executable work items.
- Send orchestrator run events.
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
- Prevent workspace paths from escaping the configured root.

The current implementation creates directories and records their paths on runs. Repository population and cleanup policy are tracked as Symphony TODO items.

### `runner`

Defines the agent runner boundary.

Responsibilities:

- Define the runner interface.
- Define runner task input and result output.
- Provide a simulated runner for tests and a Codex app-server subprocess runner for real execution.

The Codex runner follows the contract documented in [../../docs/symphony/CODEX_APP_SERVER.md](../../docs/symphony/CODEX_APP_SERVER.md).

### `worker`

Coordinates polling, claiming, workspace creation, runner dispatch, and outbox delivery.

Responsibilities:

- Poll the issue-tracker work item queue.
- Claim work items with a lease.
- Respect configured concurrency and turn limits.
- Create a workspace for each claimed item.
- Create and update run records through `runstore`.
- Invoke the configured `runner.Runner`.
- Flush outbox events through `tracker`.

This package coordinates components but should avoid owning their internal details. Persistence belongs in `runstore`, tracker API behavior belongs in `tracker`, workspace path safety belongs in `workspace`, and agent execution belongs behind `runner`.

## Dependency Direction

Keep dependencies flowing inward to small contracts and outward only from coordinators.

```text
cmd/orchestrator
  └─ worker
       ├─ runstore ── run
       ├─ tracker ─── run
       ├─ runner ──── run, workspace
       ├─ workflow
       └─ workspace
```

Rules:

- `run` should remain dependency-light.
- `runstore`, `tracker`, `workflow`, `workspace`, and `runner` should not import `worker`.
- `worker` may coordinate other packages, but should not duplicate their responsibilities.
- `cmd/orchestrator` should wire packages together instead of hiding construction behind a broad root package.

## Current Gaps

Open Symphony-related implementation work is tracked in [../../docs/symphony/TODO.md](../../docs/symphony/TODO.md).
