# Tasq Architecture

Tasq is a local-first task system for managing issues and observing orchestrator run state.

The current architecture separates issue management from orchestration. The issue-tracker owns issue state and the user-facing API. The orchestrator owns historical run state and optional runtime inspection. UI clients primarily talk to the issue-tracker. The Web UI server also exposes an orchestrator proxy path for future run-state views.

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

The web-ui is a Go-served Vite + React single-page app for issue operations.

Responsibilities:

- Request issue summaries from the issue-tracker.
- Display issue status, priority, and assignee.
- Move issues between issue statuses by calling the issue-tracker.
- Serve browser routes through SPA fallback.
- Proxy `/tracker/*` to the issue-tracker and `/orchestrator/*` to the orchestrator.

For Web UI structure and styling conventions, see [web.md](web.md).

### tui

The TUI is a Go terminal client for the same issue-tracker API.

Responsibilities:

- Fetch issue summaries from the issue-tracker.
- Render issue columns.
- Support one-shot and watch-mode rendering.
- Avoid direct calls to the orchestrator.

This legacy `cmd/tasq-tui` component remains a summary-oriented one-shot/watch client. The canonical interactive terminal entry point is `tq tui`.

### tq

`tq` is a standalone Go CLI for agents and workflow tools that need to mutate issue state without embedding HTTP details.

Responsibilities:

- Create, read, list, and update issues through the issue-tracker API.
- Upload image attachments for issue descriptions and comments.
- Support human-readable output by default and JSON output for tool use.
- Resolve the issue-tracker API URL from `--api-url`, `TQ_API_URL`, `$TQ_HOME/system/state.json`, or `http://localhost:37651`.
- Manage host-local issue-tracker and orchestrator processes through `tq service`.
- Render text errors as colored `Error: <message>` on stderr by default; preserve the machine-readable `{"error":"<message>"}` envelope when `--output json` is selected, always with a non-zero exit code.
- Avoid direct calls to the orchestrator.
- Provide the experimental read-only `tq tui` console. It reads issue data from the issue-tracker and may read runtime inspection from the orchestrator, but never calls mutation or refresh endpoints.

### issue-tracker

The issue-tracker owns issue management and display aggregation.

Responsibilities:

- Store issues in SQLite.
- Require each issue to belong to one project.
- Create, edit, and list issues.
- Store attachment metadata in SQLite and attachment bytes under `$TQ_HOME`.
- Return issue states for orchestrator or tool reconciliation.
- Compute ready-issue queue state from issue dependencies; `queued` and `pending` are derived states and are not persisted.
- Serve the UI/TUI summary API.

The issue-tracker is the source of truth for issue status, priority, title, description, assignee, comments, attachments, and projects.
It is also the source of truth for issue dependency edges and queue eligibility. The orchestrator reads `GET /api/v1/queue` and dispatches only the returned `queued` issues; it does not duplicate dependency resolution.
Projects cannot be deleted while linked issues exist.

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

User-facing clients and agent-facing workflow tools depend on the issue-tracker API. The optional `tq tui` Run tab also has a narrowly scoped read-only dependency on the orchestrator inspection API.

The orchestrator no longer uses issue-tracker event receiver endpoints. Historical run and runner-event data stays in the orchestrator SQLite store and is exposed by the optional orchestrator HTTP API. Dispatch eligibility is still read from the issue-tracker through the derived queue endpoint because dependency resolution belongs to issue state ownership.

```text
web-ui ─┐
tui ────┼─ issue-tracker ── SQLite: issues, issue_dependencies, comments, attachments, projects
tq ─────┘
                 │
                 └─ $TQ_HOME/system/data/attachments

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
- issue-tracker SQLite tables for issues and projects.
- issue-tracker attachment metadata in SQLite and image bytes under `$TQ_HOME`.
- orchestrator SQLite tables for runs, runner events, workspace metadata, and workspace setup failures.
- issue-tracker summary API consumed by web-ui and tui.
- issue CRUD API consumed by `tq`.
- Markdown issue descriptions and comment bodies with `attachment://<id>` image references.
- Codex runner lifecycle: app-server startup, live-thread turns, continuation turns when enabled, and terminal run status reporting.

The simulated runner remains available for narrow tests, but production wiring uses the Codex app-server runner.
