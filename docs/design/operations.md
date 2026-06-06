# Tasq Operations

This document covers the local development environment, verification commands, and open design decisions. For ownership boundaries and component responsibilities, see [architecture.md](architecture.md). For the user-facing API surface, see [api.md](api.md).

## Development Environment

Docker Compose keeps local development in one long-lived `dev` container and a standalone OpenAPI UI container. The issue-tracker listens on container port `8080`, the orchestrator listens on container port `8081`, and the Go Web server listens on container port `3000` inside `dev`.

For host-only operation on a personal machine, `tq service start` runs issue-tracker and orchestrator as background processes. It uses fixed local ports `37651` and `37652`, writes discovery state to `$TQ_HOME/system/state.json`, and appends logs under `$TQ_HOME/system/log/`.

Recommended commands:

- `make run-issue-tracker`
- `make run-orchestrator`
- `make dev-up`
- `make run-web`
- `make run-tui`
- `make dev-ports`
- `make dev-codex-login`

CLI commands:

- `make run-tq ARGS="issue list"`
- `make run-tq ARGS="issue get 1"`
- `TQ_HOME=./.tasq go run ./cmd/tq service status`

`make dev-up` starts the OpenAPI UI and launches the issue-tracker, orchestrator, and web-ui inside the `dev` container. Runtime state is stored under `$TQ_HOME`, which defaults to `/workspace/.tasq` inside the container. `make dev-codex-login` uses device auth and persists Codex authentication in the `codex-home` Docker volume.

## Autonomous Approval Boundary

Tasq should run agents in an autonomous environment where routine repository work can proceed without a human in the loop, while requests that cross the documented safety boundary still become Codex `requestApproval` events.

The target posture is not to eliminate approvals. The target is to make approvals meaningful. Basic inspection, local edits inside the issue workspace, and normal verification commands should run inside the configured sandbox. Broader commands, filesystem access outside the workspace, credential access, network access, and other host-affecting actions should remain outside the autonomous boundary and surface as approval-required work.

For local development, the baseline environment is:

- Run Codex app-server inside the Tasq `dev` container.
- Start Codex with the documented workspace-write sandbox posture from [ADR-0006](../adr/0006-run-app-server-with-workspace-write-sandbox.md).
- Keep Codex authentication and personal state in the `codex-home` Docker volume, not in the repository.
- Mount repository-managed Codex rules read-only from `codex/rules/`.
- Keep Tasq runtime state under `$TQ_HOME`, separate from Codex credentials.
- Treat command-execution and file-change `requestApproval` events according to [ADR-0005](../adr/0005-block-issues-for-app-server-approval-decisions.md): cancel the request, fail the run with `approval_required`, and block the issue with the request details.

This gives Tasq a safe unattended default: the agent can make progress within the container and workspace boundary, but Tasq does not silently approve actions that require a wider trust decision.

Operators should review blocked issues out of band. If a request is acceptable, the operator can change the environment, add a narrower rule, or retry the issue after making an explicit approval decision. Tasq should not convert a blocked approval into success unless the requested action actually ran under a documented policy.

## Verification

Current verification commands:

```sh
go test ./...
```

```sh
cd cmd/web/frontend
npm run typecheck
npm run build
```

Manual verification:

1. Start the dev environment with `make dev-up`.
2. Create and update issues through the UI or `tq`.
3. Confirm the issue-tracker summary reflects issue status changes.
4. Confirm orchestrator runtime inspection is available through the printed orchestrator URL.

The Web server proxies `/tracker/*` to the issue-tracker and `/orchestrator/*` to the orchestrator. In Compose, `make run-web` starts the Web server with backend URLs pointing at `127.0.0.1:8080` and `127.0.0.1:8081` inside the dev container.

## Open Decisions

- Whether external tracker sync belongs inside issue-tracker or behind a provider interface.
- Production authentication, authorization, and network exposure.
- Whether large full-fidelity Codex transcripts should remain in SQLite or move to filesystem artifacts.
