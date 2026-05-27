# Tasq Design

Tasq is a Symphony-compatible task queue orchestrator for running coding agents against tracker tasks.

The initial implementation focuses on the orchestration control plane: task state is stored in SQLite, exposed through a REST API, and consumed by both GUI and TUI clients. Agent execution and external tracker synchronization are intentionally left as later implementation slices.

## Goals

- Manage tasks and agent run state from one authoritative orchestrator.
- Provide both GUI and TUI clients over the same REST API.
- Keep the orchestrator independent from UI-specific state.
- Preserve a path toward Symphony compatibility for workflow-driven agent execution.
- Make implementation-defined policy explicit before it becomes hidden behavior.

## Non-goals

- A hosted multi-tenant service.
- A rich dashboard beyond the task operations UI.
- Direct built-in business rules for closing tracker tickets or linking pull requests.
- A complete Codex app-server runner implementation in the initial slice.
- A complete Linear synchronization implementation in the initial slice.

## Source Specification

Tasq is based on the Symphony `SPEC.md` contract.

Symphony defines the broad service shape: workflow loading, tracker polling, workspace management, agent running, retry/reconciliation behavior, structured logging, and optional status surfaces.

Symphony intentionally leaves several policies implementation-defined. Tasq must document these policies as they are added, especially for sandboxing, approvals, workspace synchronization, retry limits, observability exposure, and restart recovery.

## System Architecture

Tasq has three user-facing layers:

- GUI: React, TypeScript, and Next.js.
- TUI: Go terminal client.
- Orchestrator: Go service backed by SQLite.

The orchestrator is the only component that owns persistent system state. Both UIs communicate with the orchestrator through REST and do not read SQLite directly.

```text
GUI (Next.js)  ─┐
                ├─ REST API ─ Orchestrator (Go) ─ SQLite
TUI (Go)      ─┘
```

## Responsibility Boundaries

### Orchestrator

The orchestrator owns durable task state, global settings, API validation, and the canonical task summary.

Current responsibilities:

- Store task content and operational state.
- Store global settings.
- Serve task, summary, health, and settings endpoints.
- Normalize task board columns.
- Expose active agent status from task state.

Planned responsibilities:

- Load `WORKFLOW.md` definitions.
- Poll external trackers such as Linear.
- Create and manage workspaces.
- Dispatch agent runs.
- Record attempts, retries, terminal outcomes, and run errors.
- Reconcile state after restart.

### GUI

The GUI is the browser-based operational interface.

Current responsibilities:

- Show a Kanban board.
- Show active agent status.
- Show task details.
- Show and update global settings.
- Move tasks between statuses through the REST API.

The GUI must treat the REST API as the contract. It should not infer hidden orchestrator state.

### TUI

The TUI is the terminal-based operational interface.

Current responsibilities:

- Fetch the orchestrator summary endpoint.
- Render active agent state.
- Render tasks grouped as Kanban columns.
- Support one-shot rendering and watch mode.

Future TUI actions should use the same REST endpoints as the GUI.

## Persistent State

SQLite is the system-of-record for the orchestrator.

Current tables:

- `tasks`
- `settings`

The database is local to the orchestrator process. UIs must not connect to it.

Runtime SQLite files are ignored by Git:

- `*.sqlite`
- `*.sqlite-shm`
- `*.sqlite-wal`

## Task Model

A task represents a unit of work that can eventually be sourced from a tracker issue or created manually.

Current fields:

- `id`: internal numeric identifier.
- `title`: required human-readable task name.
- `description`: task body or context.
- `status`: Kanban/workflow state.
- `priority`: scheduling hint.
- `agentStatus`: current agent execution state.
- `assignee`: human or agent owner label.
- `source`: origin system, such as `linear` or `manual`.
- `sourceId`: tracker issue identifier when available.
- `workspace`: workspace path or identifier.
- `attempts`: number of run attempts.
- `lastError`: latest execution or orchestration error.
- `createdAt`: creation timestamp.
- `updatedAt`: last update timestamp.

## Status Model

Task statuses:

- `backlog`
- `ready`
- `running`
- `review`
- `done`
- `blocked`
- `failed`

Agent statuses:

- `idle`
- `queued`
- `running`
- `waiting_for_input`
- `succeeded`
- `failed`

Priority values:

- `low`
- `normal`
- `high`
- `urgent`

## Kanban Semantics

The Kanban board is a projection of task state.

The orchestrator returns all board columns in a stable order, even when a column has no tasks. Empty task lists must be encoded as empty arrays, not `null`, so both GUI and TUI can use the same contract without special cases.

The current column order is:

1. Backlog
2. Ready
3. Running
4. Review
5. Blocked
6. Failed
7. Done

## REST API

The REST API is the public contract between orchestrator and UIs.

Current endpoints:

- `GET /api/v1/health`
- `GET /api/v1/summary`
- `GET /api/v1/tasks`
- `POST /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PATCH /api/v1/tasks/{id}`
- `GET /api/v1/settings`
- `PUT /api/v1/settings`

### Error Shape

Errors are returned as JSON objects with an `error` field.

The API should keep this shape stable unless a versioned API is introduced.

### CORS

The orchestrator currently allows local browser origins that start with:

- `http://localhost:`
- `http://127.0.0.1:`

This is a development-oriented policy. Production exposure requires a separate decision for authentication, allowed origins, and network binding.

## Settings

Current global settings:

- `pollIntervalSeconds`
- `maxConcurrentRuns`
- `workspaceRoot`
- `trackerProvider`
- `agentCommand`

Settings are stored in SQLite and exposed through REST.

At this stage, settings are operational metadata. Not every setting is fully enforced by a running scheduler yet.

## Workspace Policy

Current implementation stores a workspace string on each task but does not create, sync, clean, or delete workspaces.

Future workspace management must define:

- How a workspace is populated.
- Whether workspaces are Git worktrees, clones, or plain directories.
- How existing workspace paths are handled.
- What happens when a path exists but is not usable.
- When terminal task workspaces are cleaned up.
- What artifacts are retained for debugging.

## Agent Runner Policy

The initial implementation models agent status but does not run agents.

Future agent execution must define:

- Which Codex app-server protocol version is supported.
- What command starts an agent session.
- How approvals are surfaced.
- What sandbox policy is used.
- How user input requests are represented.
- How rate-limit and usage signals affect scheduling.
- Which terminal states update task status.

## Tracker Policy

The initial implementation does not poll Linear or any other tracker.

Future tracker integration must define:

- Which tracker fields map to Tasq task fields.
- How tracker labels or states determine eligibility.
- How duplicate tracker issues are reconciled.
- Whether Tasq writes comments or state transitions back to the tracker.
- Whether tracker updates are agent-owned, orchestrator-owned, or workflow-owned.

By default, Symphony expects ticket business logic to live in workflow prompts and agent tooling rather than in the orchestrator.

## Retry And Recovery Policy

The initial implementation stores `attempts` and `lastError` but does not schedule retries.

Future retry behavior must define:

- Maximum retry count.
- Backoff cap.
- Which errors are retryable.
- When a task becomes `failed` or `blocked`.
- Whether operator intervention is required after repeated failures.

Restart recovery must define what is reconstructed from SQLite, filesystem workspaces, and tracker state. Symphony does not require restoring in-memory retry timers or live worker sessions after process restart.

## Observability

Current observability:

- HTTP request logging through the standard logger.
- `GET /api/v1/health`.
- `GET /api/v1/summary`.

Future observability must define:

- Structured log fields.
- Log sink and retention.
- Redaction rules.
- Snapshot API stability.
- Metrics and alerting scope.

## Security And Trust

Current trust posture is development-only.

The REST API has no authentication. Local CORS is allowed. Agent execution is not yet implemented.

Before running untrusted tasks or exposing the service beyond localhost, Tasq must define:

- Authentication and authorization.
- Allowed API origins.
- Network bind policy.
- Agent sandbox policy.
- Approval policy.
- Secret handling and redaction.
- Workspace filesystem boundaries.

## Development Environment

The repository provides a Dev Container with:

- Go 1.22.
- Node.js 22.
- GitHub CLI.
- Forwarded ports `8080` and `3000`.

After container creation, dependencies are installed with:

- `go mod download`
- `npm install` in `web`

## Verification

Current verification commands:

```sh
go test ./...
```

```sh
cd web
npm run typecheck
npm run build
npm audit --audit-level=moderate
```

## Open Decisions

The following decisions are intentionally not final yet:

- Exact Symphony compliance level for the first release.
- Workflow file schema coverage and validation strictness.
- Linear integration ownership and write-back policy.
- Workspace implementation strategy.
- Agent runner protocol version and sandbox defaults.
- Retry limits and manual intervention thresholds.
- API authentication and production exposure model.
- Whether the GUI should remain a single tabbed page or split into routed pages.

These should be resolved before implementing the corresponding subsystem.
