# Tasq API and Operations

This document covers the user-facing issue-tracker API, local development environment, verification commands, and open design decisions. For ownership boundaries and component responsibilities, see [architecture.md](architecture.md).

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
- `GET /api/v1/issues/{issueId}/comments`
- `POST /api/v1/issues/{issueId}/comments`
- `PATCH /api/v1/comments/{id}`
- `GET /api/v1/attachments`
- `POST /api/v1/attachments`
- `GET /api/v1/attachments/{id}/content`
- `DELETE /api/v1/attachments/{id}`

Attachment uploads accept multipart form data with `entity_type`, `entity_id`, and `file`. The first implementation supports PNG, JPEG, GIF, and WebP image files up to 5 MiB. Attachment bytes are stored below `$TQ_HOME/system/data/attachments`, while SQLite stores metadata and relative paths. Issue and comment text references attachments with Markdown image links such as `![screenshot](attachment://att_...)`.

JSON success responses use `{ "data": ..., "meta": {} }`. JSON error responses use `{ "error": { "code": "...", "message": "..." }, "meta": {} }`.

The `tq` CLI wraps issue CRUD endpoints with these commands:

- `tq issue list`
- `tq issue get <id>`
- `tq issue create --title <title> [--description ...] [--status ...] [--priority ...] [--assignee ...]`
- `tq issue update <id> [--title ...] [--description ...] [--status ...] [--priority ...] [--assignee ...]`
- `tq issue create ... --attach <image-path>`
- `tq issue update <id> ... --attach <image-path>`
- `tq issue close <id>`
- `tq issue ready <id>`
- `tq issue draft <id>`
- `tq issue rename <id> <title>`
- `tq issue edit <id> <description>`
- `tq comment add <issue-id> --body <body> [--attach <image-path>]`
- `tq comment list <issue-id>`

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
