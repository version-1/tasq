# Tasq Design

Tasq is a local-first task execution system for managing issues, assigning executable work to coding agents, and observing agent run state from both a web UI and a TUI.

The current architecture separates issue management from orchestration. The issue-tracker owns issue state and the user-facing API. The orchestrator owns agent run state and work assignment state. UI clients talk to the issue-tracker only.

## Goals

- Keep issue state and run state as separate concepts with separate owners.
- Let web-ui and tui use one user-facing API surface.
- Make work assignment safe for parallel orchestrator instances.
- Preserve run state changes when the issue-tracker is temporarily unavailable.
- Keep the first implementation slice small enough to verify before adding a real Codex app-server runner.

## Non-goals

- Hosted multi-tenant operation.
- Production authentication and authorization.
- A complete Codex app-server runner in the first slice.
- Full workspace lifecycle management in the first slice.
- External tracker integrations such as Linear in the first slice.
- A distributed queue beyond the SQLite-backed issue-tracker work item queue.

## Components

### web-ui

The web-ui is a Next.js client for issue operations.

Responsibilities:

- Request issue summaries from the issue-tracker.
- Display issue status, priority, assignee, and latest run state.
- Move issues between issue statuses by calling the issue-tracker.
- Avoid direct calls to the orchestrator.

For Web UI structure and styling conventions, see [../web/docs/design.md](../web/docs/design.md).

### tui

The TUI is a Go terminal client for the same issue-tracker API.

Responsibilities:

- Fetch issue summaries from the issue-tracker.
- Render issue columns and latest run state.
- Support one-shot and watch-mode rendering.
- Avoid direct calls to the orchestrator.

### issue-tracker

The issue-tracker owns issue management and display aggregation.

Responsibilities:

- Store issues in SQLite.
- Create, edit, and list issues.
- Decide when an issue becomes executable.
- Create work items when an issue is ready to run.
- Expose a lease-backed work item claim API for orchestrators.
- Receive orchestrator run events idempotently.
- Apply issue status transitions based on run facts.
- Serve the UI/TUI summary API.

The issue-tracker is the source of truth for issue status, priority, title, description, assignee, work item claim state, and received orchestrator event ids.

### orchestrator

The orchestrator owns agent assignment and run state.

Responsibilities:

- Poll the issue-tracker work item queue.
- Claim executable work items with a lease.
- Create run records in its own SQLite database.
- Emit run state changes through a durable outbox.
- Retry outbox delivery to the issue-tracker until accepted.
- In the MVP, simulate the run lifecycle enough to verify the boundary.

The orchestrator is the source of truth for run records, run attempts, claim tokens attached to runs, and outbox delivery state.

### agent

The future agent is a Codex app-server process controlled by the orchestrator.

Responsibilities:

- Receive tasks from the orchestrator.
- Execute the task inside a workspace.
- Report execution progress to the orchestrator over JSON-RPC.

The MVP does not start a real agent process. It records a minimal run lifecycle so the issue-tracker/orchestrator contract can be verified first.

### workspace

The future workspace manager provides isolated execution environments for agents.

Responsibilities:

- Create and manage git workspaces.
- Support parallel execution and verification in isolated workspaces.
- Retain enough metadata for debugging and recovery.

The MVP stores workspace identifiers on run state but does not create or clean real workspaces.

## Dependency Direction

User-facing clients depend on the issue-tracker API only.

The orchestrator depends on the issue-tracker work queue and event receiver APIs. The issue-tracker does not poll the orchestrator. Instead, it stores the latest run snapshots from orchestrator push events.

```text
web-ui ─┐
        ├─ issue-tracker ── SQLite: issues, work_items, received_events, run_snapshots
tui ────┘       ▲
                │ claim work item / push run event
                │
        orchestrator ───── SQLite: runs, outbox_events
                │
                ├─ future: agent-runner ── Codex app-server over JSON-RPC
                └─ future: workspace manager ── git workspace / isolated runtime
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

The orchestrator never directly changes issue status. It emits run facts. The issue-tracker receives those facts and applies issue status rules.

## Work Item Queue

An issue becomes executable when its issue status is changed to `ready`.

When an issue becomes `ready`, the issue-tracker creates a pending work item. The orchestrator does not scan all issues and does not decide whether an issue is executable. It polls the work item queue only.

Re-running the same issue creates a new work item. This keeps claim tokens, run attempts, and results tied to a single execution request.

## Claim And Lease

Work item claim is lease-backed.

When an orchestrator claims a work item, the issue-tracker records:

- `claimed_by`
- `claim_token`
- `lease_until`
- incremented attempt count

If the orchestrator dies or stops renewing in a later implementation, the work item becomes claimable again after `lease_until`.

The claim token is the generation marker for one work item claim. Orchestrator run events are applied only when their claim token matches the current claim token for the work item. Late events from expired claims are recorded for idempotency but are not allowed to update issue state.

## Run Events And Outbox

The orchestrator writes run events to its SQLite outbox before sending them to the issue-tracker.

The issue-tracker accepts run events idempotently:

- Each event has a unique `eventId`.
- Processed event ids are stored in SQLite.
- Duplicate event ids are treated as already accepted.

This allows the orchestrator to retry delivery without double-applying state transitions.

## Current MVP Behavior

The current implementation slice includes:

- `cmd/issue-tracker`
- `cmd/orchestrator`
- issue-tracker SQLite tables for issues, work items, received orchestrator events, and run snapshots.
- orchestrator SQLite tables for runs and outbox events.
- issue-tracker summary API consumed by web-ui and tui.
- lease-backed work item claim API.
- idempotent run event receiver.
- orchestrator polling, claim, run creation, and outbox delivery.
- a minimal simulated run lifecycle: `queued -> running -> succeeded`.

The simulated lifecycle is intentionally temporary. Its purpose is to prove the service boundary before adding the Codex app-server runner and real workspace manager.

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
- `GET /api/v1/workspaces`
- `POST /api/v1/workspaces`
- `GET /api/v1/workspaces/{id}`
- `PATCH /api/v1/workspaces/{id}`
- `DELETE /api/v1/workspaces/{id}`
- `GET /api/v1/issues`
- `POST /api/v1/issues`
- `GET /api/v1/issues/{id}`
- `PATCH /api/v1/issues/{id}`
- `POST /api/v1/work-items/claim`
- `POST /api/v1/orchestrator-events`

Errors are returned as JSON objects with an `error` field.

The orchestrator currently has no user-facing HTTP API. Its external dependency is the issue-tracker API.

## Development Environment

Docker Compose runs the issue-tracker on container port `8080`, the web-ui on container port `3000`, and the orchestrator as a background worker.

Recommended commands:

- `make issue-tracker-up`
- `make orchestrator-up`
- `make dev-up`
- `make dev-up-forward`
- `make web-up`
- `make tui-up`
- `make dev-status`

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

Manual MVP verification:

1. Start issue-tracker and orchestrator.
2. Create an issue with status `ready`.
3. Confirm the issue-tracker creates a work item.
4. Confirm the orchestrator claims it and emits run events.
5. Confirm the issue-tracker summary shows the issue in `review` with latest run status `succeeded`.

## Open Decisions

- Exact Codex app-server JSON-RPC contract.
- Agent runner process lifecycle and cancellation behavior.
- Lease renewal cadence for long-running real agents.
- Workspace implementation strategy and cleanup policy.
- Retry limits and manual intervention thresholds.
- Whether external tracker sync belongs inside issue-tracker or behind a provider interface.
- Production authentication, authorization, and network exposure.
- Whether run logs are stored directly in SQLite or referenced as filesystem artifacts.
