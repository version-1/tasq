# Tasq Operations

This document covers the local development environment, verification commands, and open design decisions. For ownership boundaries and component responsibilities, see [architecture.md](architecture.md). For the user-facing API surface, see [api.md](api.md).

## Development Environment

Docker Compose keeps local development in one long-lived `dev` container and a standalone OpenAPI UI container. The issue-tracker listens on container port `8080`, the orchestrator listens on container port `8081`, and the web-ui listens on container port `3000` inside `dev`.

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

1. Start the dev environment with `make dev-up`.
2. Create and update issues through the UI or `tq`.
3. Confirm the issue-tracker summary reflects issue status changes.
4. Confirm orchestrator runtime inspection is available through the printed orchestrator URL.

## Open Decisions

- Whether external tracker sync belongs inside issue-tracker or behind a provider interface.
- Production authentication, authorization, and network exposure.
- Whether large full-fidelity Codex transcripts should remain in SQLite or move to filesystem artifacts.
