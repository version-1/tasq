# Tasq API と運用

このドキュメントでは user-facing な issue-tracker API、local development environment、verification command、open design decision を扱います。所有境界と component responsibility は [architecture.ja.md](architecture.ja.md) を参照してください。

## API Surface

issue-tracker は user-facing API です。

現在の issue-tracker endpoint:

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

Attachment upload は `entity_type`、`entity_id`、`file` を持つ multipart form data を受け取ります。最初の実装では PNG、JPEG、GIF、WebP image file を 5 MiB までサポートします。Attachment bytes は `$TQ_HOME/system/data/attachments` 配下に保存し、SQLite には metadata と relative path を保存します。Issue と comment text は `![screenshot](attachment://att_...)` のような Markdown image link で attachment を参照します。

JSON success response は `{ "data": ..., "meta": {} }` を使います。JSON error response は `{ "error": { "code": "...", "message": "..." }, "meta": {} }` を使います。

`tq` CLI は issue CRUD endpoint を次の command で wrap します。

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

`tq` は default では human-readable output を使い、`--output json` が指定された場合は JSON output を使います。

orchestrator は `--port` または `server.port` で有効化したときに runtime inspection 用の optional loopback HTTP API を公開します。

## Development Environment

Docker Compose は issue-tracker を container port `8080`、web-ui を container port `3000`、orchestrator service を optional runtime inspection 用に実行します。

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

`make web-up` は issue-tracker、orchestrator、web-ui を起動します。Web UI は `/api/v1/...` を Compose network 内の issue-tracker に proxy します。

## Verification

現在の verification command:

```sh
go test ./...
```

```sh
cd web
npm run typecheck
npm run build
```

Manual verification:

1. issue-tracker と Web UI を起動する。
2. UI または `tq` で issue を作成・更新する。
3. issue-tracker summary が issue status change を反映することを確認する。
4. runtime inspection が必要な場合は `--port` 付きで orchestrator を起動する。

## Open Decisions

- external tracker sync を issue-tracker 内に置くか provider interface の behind に置くか。
- Production authentication、authorization、network exposure。
- large full-fidelity Codex transcript を SQLite に残すか filesystem artifact に移すか。
