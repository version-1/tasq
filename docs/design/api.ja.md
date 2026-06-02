# Tasq API

このドキュメントでは user-facing な issue-tracker API を扱います。所有境界と component responsibility は [architecture.ja.md](architecture.ja.md) を参照してください。local development と verification は [operations.ja.md](operations.ja.md) を参照してください。

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

Issue は必ず 1 つの project に属します。`POST /api/v1/issues` は `projectId` を必須とし、issue response は `projectId` と `projectKey` の両方を返します。`GET /api/v1/issues` は optional query parameter として `states` と `project_id` を受け取ります。`project_id` を省略した場合は全 project の issue を一覧表示します。

JSON success response は `{ "data": ..., "meta": {} }` を使います。JSON error response は `{ "error": { "code": "...", "message": "..." }, "meta": {} }` を使います。

`tq` CLI は issue CRUD endpoint を次の command で wrap します。

- `tq issue list [--project <project-key>]`
- `tq issue get <id>`
- `tq issue create --project <project-key> --title <title> [--description ...] [--status ...] [--priority ...] [--assignee ...]`
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
