# Tasq Design

Tasq is a local-first task system for managing issues and observing orchestrator run state.

The current architecture separates issue management from orchestration. The issue-tracker owns issue state and the user-facing API. The orchestrator owns historical run state and optional runtime inspection. UI clients talk to the issue-tracker only.

## Goals

- Keep issue state and run state as separate concepts with separate owners.
- Let web-ui, tui, and agent-facing CLI tools use one user-facing API surface.
- Keep the tracker API focused on issue, project, workspace, and summary data.
- Keep the orchestration runtime state local to the orchestrator.

## Non-goals

- Hosted multi-tenant operation.
- Production authentication and authorization.
- A complete Codex app-server runner in the first slice.
- External tracker integrations such as Linear in the first slice.
- Dispatch or worker scheduling semantics.

## Components

### web-ui

The web-ui is a Next.js client for issue operations.

Responsibilities:

- Request issue summaries from the issue-tracker.
- Display issue status, priority, and assignee.
- Move issues between issue statuses by calling the issue-tracker.
- Avoid direct calls to the orchestrator.

For Web UI structure and styling conventions, see [../web/docs/design.md](../web/docs/design.md).

### tui

The TUI is a Go terminal client for the same issue-tracker API.

Responsibilities:

- Fetch issue summaries from the issue-tracker.
- Render issue columns.
- Support one-shot and watch-mode rendering.
- Avoid direct calls to the orchestrator.

### tq

`tq` is a standalone Go CLI for agents and workflow tools that need to mutate issue state without embedding HTTP details.

Responsibilities:

- Create, read, list, and update issues through the issue-tracker API.
- Support human-readable output by default and JSON output for tool use.
- Resolve the issue-tracker API URL from `--api-url`, `TQ_API_URL`, or `http://localhost:8080`.
- Return machine-readable JSON errors on stderr and a non-zero exit code when a command fails.
- Avoid direct calls to the orchestrator.

### issue-tracker

The issue-tracker owns issue management and display aggregation.

Responsibilities:

- Store issues in SQLite.
- Create, edit, and list issues.
- Return issue states for orchestrator or tool reconciliation.
- Serve the UI/TUI summary API.

The issue-tracker is the source of truth for issue status, priority, title, description, assignee, projects, and workspaces.

### orchestrator

The orchestrator owns run state and runtime inspection.

Responsibilities:

- Create run records in its own SQLite database.
- Load the repository workflow contract used to configure orchestration.
- Create sanitized per-issue workspaces under the configured workspace root.
- Record runner events and workspace metadata.
- Expose optional loopback HTTP APIs for runtime state and issue-specific run details.

The orchestrator is the source of truth for run records, run attempts, runner events, and workspace metadata.

### agent

The future agent is a Codex app-server process controlled by the orchestrator.

Responsibilities:

- Receive tasks from the orchestrator.
- Execute the task inside a workspace.
- Report execution progress to the orchestrator over JSON-RPC.

The orchestrator starts Codex app-server through the runner boundary and records run progress in its local runstore.

### workspace

The workspace manager provides isolated execution environments for agents.

Responsibilities:

- Create and manage git workspaces.
- Support parallel execution and verification in isolated workspaces.
- Retain enough metadata for debugging and recovery.

The current workspace manager creates sanitized per-issue workspace directories, populates newly created workspaces from the configured repository source, and records cleanup/population metadata for recovery and debugging.

## Dependency Direction

User-facing clients and agent-facing workflow tools depend on the issue-tracker API only.

The orchestrator no longer uses issue-tracker work queue or event receiver endpoints. Historical run and runner-event data stays in the orchestrator SQLite store and is exposed by the optional orchestrator HTTP API.

```text
web-ui ─┐
tui ────┼─ issue-tracker ── SQLite: issues, projects, workspaces
tq ─────┘

        orchestrator ───── SQLite: runs, runner_events, workspace metadata
                │
                ├─ future: agent-runner ── Codex app-server over JSON-RPC
                └─ workspace manager ── git workspace / isolated runtime
```

## State Ownership

Issue status and run status are separate.

Issue status belongs to the issue-tracker:

- `backlog`
- `ready`
- `in_progress`
- `review`
- `done`
- `blocked`
- `failed`

Run status belongs to the orchestrator:

- `queued`
- `starting`
- `running`
- `waiting_for_input`
- `succeeded`
- `failed`
- `cancelled`

The orchestrator does not directly change issue status. Issue status changes go through the issue-tracker issue APIs.

## Current MVP Behavior

The current implementation slice includes:

- `cmd/issue-tracker`
- `cmd/tq`
- `cmd/orchestrator`
- issue-tracker SQLite tables for issues, projects, and workspaces.
- orchestrator SQLite tables for runs, runner events, workspace metadata, and workspace setup failures.
- issue-tracker summary API consumed by web-ui and tui.
- issue CRUD API consumed by `tq`.
- Codex runner lifecycle: app-server startup, live-thread turns, continuation turns when enabled, and terminal run status reporting.

The simulated runner remains available for narrow tests, but production wiring uses the Codex app-server runner.

## API Surface

The issue-tracker is the user-facing API.

Current issue-tracker endpoints:

- `GET /api/v1/health`
- `GET /api/v1/summary`
- `GET /api/v1/projects`
- `POST /api/v1/projects`
- `GET /api/v1/projects/{id}`
- `PATCH /api/v1/projects/{id}`
- `DELETE /api/v1/projects/{id}`
- `POST /api/v1/projects/{id}/check`
- `GET /api/v1/workspaces`
- `POST /api/v1/workspaces`
- `GET /api/v1/workspaces/{id}`
- `PATCH /api/v1/workspaces/{id}`
- `DELETE /api/v1/workspaces/{id}`
- `GET /api/v1/issues`
- `POST /api/v1/issues`
- `POST /api/v1/issues/states`
- `GET /api/v1/issues/{id}`
- `PATCH /api/v1/issues/{id}`

JSON success responses use `{ "data": ..., "meta": {} }`. JSON error responses use `{ "error": { "code": "...", "message": "..." }, "meta": {} }`.

The `tq` CLI wraps issue CRUD endpoints with these commands:

- `tq issue list`
- `tq issue get <id>`
- `tq issue create --title <title> [--description ...] [--status ...] [--priority ...] [--assignee ...]`
- `tq issue update <id> [--title ...] [--description ...] [--status ...] [--priority ...] [--assignee ...]`
- `tq issue close <id>`
- `tq issue ready <id>`
- `tq issue draft <id>`
- `tq issue rename <id> <title>`
- `tq issue edit <id> <description>`

`tq` uses human-readable output by default and JSON output when `--output json` is set.

The orchestrator exposes an optional loopback HTTP API for runtime inspection when enabled with `--port` or `server.port`.

## Development Environment

Docker Compose runs the issue-tracker on container port `8080`, the web-ui on container port `3000`, and the orchestrator service for optional runtime inspection.

Recommended commands:

- `make issue-tracker-up`
- `make orchestrator-up`
- `make dev-up`
- `make dev-up-forward`
- `make web-up`
- `make tui-up`
- `make dev-status`

Host commands:

- `go run ./cmd/tq --api-url http://localhost:8080 issue list`
- `TQ_API_URL=http://localhost:8080 go run ./cmd/tq issue get 1`

`make web-up` starts the issue-tracker, orchestrator, and web-ui. The web UI proxies `/api/v1/...` to the issue-tracker inside the Compose network.

## Verification

Current verification commands:

```sh
go test ./...
```

```sh
cd web
npm run typecheck
npm run build
```

Manual verification:

1. Start issue-tracker and web UI.
2. Create and update issues through the UI or `tq`.
3. Confirm the issue-tracker summary reflects issue status changes.
4. Start orchestrator with `--port` when runtime inspection is needed.

## Open Decisions

- Whether external tracker sync belongs inside issue-tracker or behind a provider interface.
- Production authentication, authorization, and network exposure.
- Whether large full-fidelity Codex transcripts should remain in SQLite or move to filesystem artifacts.
