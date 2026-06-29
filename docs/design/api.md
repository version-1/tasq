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
- `PUT /api/v1/projects/{id}/workflow`
- `POST /api/v1/projects/{id}/check`
- `DELETE /api/v1/projects/{id}/workflow`
- `GET /api/v1/issues`
- `POST /api/v1/issues`
- `POST /api/v1/issues/states`
- `GET /api/v1/queue`
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

Issues belong to exactly one project. `POST /api/v1/issues` requires `projectId` and accepts `dependency_ids` as the initial dependency set. Issue responses include both `projectId` and `projectKey`. Issue responses include `dependency_ids`; issues without dependencies return an empty array. `GET /api/v1/issues` accepts optional `states`, `project_id`, `project_ids`, `priorities`, `assignee`, `search`, `limit`, `offset`, `sort_by`, and `sort_direction` query parameters. Omitting project filters lists issues across all projects. `project_id` limits the list to one project, while `project_ids` accepts a comma-separated set for table filters. `priorities` accepts comma-separated issue priorities. `search` matches issue IDs exactly and issue titles with case-insensitive partial matching; numeric search text matches either the exact ID or a title containing that text. Empty or whitespace-only `search` values are ignored. Search is combined with other filters before sorting and pagination. `sort_by` is limited to `id`, `priority`, `created_at`, and `updated_at`; `sort_direction` is limited to `asc` and `desc`.

`GET /api/v1/summary` returns issue board columns. Each issue summary includes `queueStatus`, which is the issue state from the system-wide queue perspective. `queueStatus=backlog` means the issue status is `backlog`. `queueStatus=pending` means the issue is `ready` but still has at least one active dependency. `queueStatus=queued` means the issue is `ready` and has no active dependencies. `queueStatus=processing` means the issue status is `in_progress`. `queueStatus=completed` means the issue status is `done`. `queueStatus=inactive` means the issue is outside the queue flow, including `review`, `blocked`, `failed`, `cancelled`, and `duplicate`. See [status.md](status.md) for status definitions and expected transitions.

### `POST /api/v1/issues`

Creates an issue in a project. `dependency_ids` is optional and sets the initial dependency issue IDs during the same create operation. Omit `dependency_ids` or pass an empty array to create an issue without dependencies. The API rejects missing dependency issues, self-dependencies, duplicate dependency IDs, and dependency cycles.

Request:

```json
{
  "projectId": 1,
  "title": "Document create dependencies",
  "description": "Update API and schema docs.",
  "status": "ready",
  "priority": "normal",
  "assignee": "docs",
  "dependency_ids": [12, 18]
}
```

Response:

```json
{
  "data": {
    "id": 42,
    "projectId": 1,
    "projectKey": "tasq",
    "title": "Document create dependencies",
    "description": "Update API and schema docs.",
    "status": "ready",
    "priority": "normal",
    "assignee": "docs",
    "dependency_ids": [12, 18],
    "createdAt": "2026-06-24T10:00:00Z",
    "updatedAt": "2026-06-24T10:00:00Z"
  },
  "meta": {}
}
```

`PATCH /api/v1/issues/{id}` accepts `dependency_ids` as an optional full replacement field. When omitted, existing dependencies are preserved. Passing an empty array removes all dependencies. The API rejects missing dependency issues, self-dependencies, duplicate dependency IDs, and updates that would create a dependency cycle.

`GET /api/v1/queue` returns ready issues split into `queued` and `pending` arrays. `queued` issues are ready and have no active dependencies; `pending` issues are ready but still have at least one active dependency. Active dependency statuses are `backlog`, `ready`, `in_progress`, and `review`; satisfied dependency statuses are `done`, `cancelled`, `duplicate`, `failed`, and `blocked`. Each array is sorted by priority descending (`urgent`, `high`, `normal`, `low`) and then ID ascending. The endpoint accepts the same `project_id` filter semantics as issue listing. Pending items include `blocked_dependency_ids` for active dependencies that keep the issue pending.

JSON success responses use `{ "data": ..., "meta": {} }`. JSON error responses use `{ "error": { "code": "...", "message": "..." }, "meta": {} }`.

The `tq` CLI wraps issue CRUD endpoints with these commands:

- `tq issue list [--project <project-key>]`
- `tq issue get <id>`
- `tq issue create --project <project-key> --title <title> [--description ...] [--status ...] [--priority ...] [--assignee ...] [--dependency <ids>]`
- `tq issue update <id> [--title ...] [--description ...] [--status ...] [--priority ...] [--assignee ...] [--dependency <ids>] [--clear-dependencies]`
- `tq issue create ... --attach <image-path>`
- `tq issue update <id> ... --attach <image-path>`
- `tq issue close <id>`
- `tq issue cancel <id>`
- `tq issue ready <id>`
- `tq issue draft <id>`
- `tq issue rename <id> <title>`
- `tq issue edit <id> <description>`
- `tq comment add <issue-id> --body <body> [--attach <image-path>]`
- `tq comment list <issue-id>`
- `tq workflow add --project <project-key> (--file <path> | --body <text>)`
- `tq workflow remove --project <project-key>`
- `tq workflow show --project <project-key> [--json]`

`tq` uses human-readable output by default and JSON output when `--output json` is set.

The orchestrator exposes an optional loopback HTTP API for runtime inspection when enabled with `--port` or `server.port`. Its issue runtime detail response includes historical run summaries; each run may include `thread_id` once the Codex app-server thread has been persisted.
