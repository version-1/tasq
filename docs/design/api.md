# Tasq API

This document covers the user-facing issue-tracker API. For ownership boundaries and component responsibilities, see [architecture.md](architecture.md). For local development and verification, see [operations.md](operations.md).

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
- `GET /api/v1/projects/{id}/workflow`
- `POST /api/v1/projects/{id}/check`
- `DELETE /api/v1/projects/{id}/workflow`
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

Issues belong to exactly one project. `POST /api/v1/issues` requires `projectId`, and issue responses include both `projectId` and `projectKey`. `GET /api/v1/issues` accepts optional `states` and `project_id` query parameters. Omitting `project_id` lists issues across all projects.

JSON success responses use `{ "data": ..., "meta": {} }`. JSON error responses use `{ "error": { "code": "...", "message": "..." }, "meta": {} }`.

The `tq` CLI wraps issue CRUD endpoints with these commands:

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

`tq` uses human-readable output by default and JSON output when `--output json` is set.

The orchestrator exposes an optional loopback HTTP API for runtime inspection when enabled with `--port` or `server.port`.
