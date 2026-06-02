# Tasq Operations

This document covers the local development environment, verification commands, and open design decisions. For ownership boundaries and component responsibilities, see [architecture.md](architecture.md). For the user-facing API surface, see [api.md](api.md).

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
