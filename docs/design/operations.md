# Tasq Operations

This document covers the local development environment, verification commands, and open design decisions. For ownership boundaries and component responsibilities, see [architecture.md](architecture.md). For the user-facing API surface, see [api.md](api.md).

## Development Environment

Docker Compose keeps local development in one long-lived `dev` container and a standalone OpenAPI UI container. The issue-tracker listens on container port `8080`, the orchestrator listens on container port `8081`, and the Go Web server listens on container port `3000` inside `dev`.

For host-only operation on a personal machine, `tq service start` runs issue-tracker, orchestrator, and web as background processes. It prefers fixed local ports `37651`, `37652`, and `37653`; when any is occupied, it proposes one OS-selected loopback port for each service and starts only after interactive confirmation (or `-y`). It rechecks the proposed ports after confirmation and fails rather than reselecting if one was claimed. The selected addresses are written to `$TQ_HOME/system/state.json`, and logs are appended under `$TQ_HOME/system/log/`.

Run `tq migrate` before starting services when a database is new or migrations are pending. `tq service start` checks the issue-tracker and orchestrator databases before launching any service process and exits with guidance to run `tq migrate` if pending migrations exist. The services also fail fast with the same guidance instead of applying schema changes automatically.

Recommended commands:

- `make run-issue-tracker`
- `make run-orchestrator`
- `make dev-up`
- `make run-web`
- `make run-tui`
- `make dev-ports`
- `make dev-codex-login`

CLI commands:

- `make build-tq`
- `make run-migrate`
- `TQ_HOME=./.tasq go run ./cmd/tq migrate`
- `make run-tq ARGS="issue list"`
- `make run-tq ARGS="issue get 1"`
- `TQ_HOME=./.tasq go run ./cmd/tq service status`
- `tq tui` (aliases: `tq console`, `tq c`)

`tq tui` requires an interactive terminal and text output. The issue-tracker URL follows the normal CLI resolution order. `--orchestrator-url` overrides the orchestrator address from service state. A missing or unavailable orchestrator degrades only the Run tab; tracker failures show a retry screen.

`make build-tq` builds the host `tq` binary at `./bin/tq` and injects the current short commit hash into `tq version` through the shared build-info ldflags variable used by release builds.

`make dev-up` starts the OpenAPI UI and launches the issue-tracker, orchestrator, and web-ui inside the `dev` container. Runtime state is stored under `$TQ_HOME`, which defaults to `/workspace/.tasq` inside the container. The `run-all` step applies migrations explicitly before starting services. `make dev-codex-login` uses device auth and persists Codex authentication in the `codex-home` Docker volume.

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
