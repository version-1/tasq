# Tasq API

This document describes the issue-tracker API's behavior and ownership boundaries. The [issue-tracker OpenAPI document](../openapi/issue-tracker.yml) is the normative source for paths, methods, parameters, and schemas. For component ownership, see [architecture.md](architecture.md); for local operation and verification, see [operations.md](operations.md).

## Contract

The issue-tracker is Tasq's user-facing API. Successful JSON responses use `{ "data": ..., "meta": {} }`; JSON errors use `{ "error": { "code": "...", "message": "..." }, "meta": {} }`.

API changes must update the OpenAPI document and generated clients as described in [development.md](../development.md). This document records cross-endpoint behavior that is easier to understand outside individual schema definitions.

## Projects

Every issue belongs to exactly one project. Project responses identify the project by both numeric ID and key, while commands generally accept the project key.

`DELETE /api/v1/projects/{id}` deletes the project and all issue-tracker descendants it owns:

- issues and dependency edges that reference those issues;
- comments and change requests;
- attachment records and files under `$TQ_HOME/system/data/attachments`; and
- stored project workflow overrides.

The operation also deletes orchestrator runtime data owned by those issues: runs, runner events, workspace metadata, and workspace setup failures. It deletes orchestrator data before issue-tracker records so a partial failure can be retried while the issue IDs still exist.

If an owned issue has a `running` orchestrator run, deletion returns `409 Conflict` with `projects.delete.running_runs` and changes nothing. Project deletion never deletes or modifies the directory in `project.location` or its worktrees.

## Issues and dependencies

`POST /api/v1/issues` requires `projectId`. The optional `dependency_ids` field sets the initial dependency set in the same operation. Omitting it or passing an empty array creates an issue without dependencies.

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

`PATCH /api/v1/issues/{id}` treats `dependency_ids` as a full replacement when present. Omitting the field preserves existing dependencies; passing an empty array removes all dependencies. Create and update both reject missing dependency issues, self-dependencies, duplicate IDs, and dependency cycles.

Issue responses include `projectId`, `projectKey`, `dependency_ids`, and `artifacts`. Dependency and artifact lists are required arrays and are `[]` when empty. Artifacts are sorted by `type`.

### Artifacts

`PUT /api/v1/issues/{issue_id}/artifacts/{type}` creates or updates one artifact for an issue and type. The request body contains only `data_value`; the server determines `data_type`. Both creation and update return `200` with the artifact's public `type`, `data_type`, and `data_value` fields.

`DELETE /api/v1/issues/{issue_id}/artifacts/{type}` removes the artifact and returns an empty `204` response. A missing issue or artifact returns `404`; an unsupported type, invalid body, or invalid URL returns `400`.

The initial supported type is `pull_request`, with `data_type` `url`. The URL is trimmed before validation and must be an absolute `http` or `https` URL with a host, without userinfo, and no more than 4,096 UTF-8 bytes. Apart from trimming, the API does not normalize the URL.

### Listing and search

`GET /api/v1/issues` supports `states`, `project_id`, `project_ids`, `priorities`, `assignee`, `search`, `limit`, `offset`, `sort_by`, and `sort_direction`.

- Omitting project filters lists issues across every project.
- `project_id` selects one project; `project_ids` accepts a comma-separated set.
- `priorities` accepts comma-separated priorities.
- `search` matches an issue ID exactly or a title with a case-insensitive substring match. Numeric text may match either.
- Empty or whitespace-only `search` values are ignored.
- Search and filters are applied before sorting and pagination.
- `sort_by` accepts `id`, `priority`, `created_at`, or `updated_at`; `sort_direction` accepts `asc` or `desc`.

## Queue and summary

`GET /api/v1/queue` divides ready issues into `queued` and `pending` arrays:

- `queued` issues have no blocking dependencies.
- `pending` issues have at least one blocking dependency and include `blocked_dependency_ids`.

Dependencies in `done`, `cancelled`, or `duplicate` are satisfied; every other status remains blocking. Both arrays sort by priority (`urgent`, `high`, `normal`, `low`) and then by ascending issue ID. The endpoint accepts the same `project_id` filter as issue listing.

`GET /api/v1/summary` exposes the board view. Each issue summary includes a queue-oriented `queueStatus`:

| `queueStatus` | Meaning |
|---|---|
| `backlog` | Issue status is `backlog`. |
| `pending` | Issue is `ready` with a blocking dependency. |
| `queued` | Issue is `ready` without a blocking dependency. |
| `processing` | Issue status is `in_progress`. |
| `completed` | Issue status is `done`. |
| `inactive` | Issue is outside the queue flow, including `review`, `blocked`, `failed`, `cancelled`, and `duplicate`. |

See [status.md](status.md) for status definitions and transition expectations.

## Comments and change requests

Comments record discussion. Change requests represent additional user or reviewer work and carry workflow state.

Creating a change request under an issue sets its status to `open`. Open requests may be edited or moved to `in_progress`; in-progress requests may be resolved or canceled. The complete transition set is:

- `open -> in_progress`
- `open -> canceled`
- `in_progress -> resolved`
- `in_progress -> canceled`

Resolved and canceled requests are immutable. Cancellation uses `POST /api/v1/change-requests/{id}/cancel`; there is no physical delete endpoint.

When the orchestrator continues an issue with a previous run, it fetches up to 20 open requests in chronological order, moves the included requests to `in_progress`, and adds them to the Codex continuation guidance. The guidance asks the agent to resolve handled requests with `resolvedByRunId` and, when available, `resultCommentId`.

## Attachments

Attachment upload uses multipart form data with `entity_type`, `entity_id`, and `file`. Supported files are PNG, JPEG, GIF, and WebP images up to 5 MiB.

File bytes are stored below `$TQ_HOME/system/data/attachments`; SQLite stores metadata and relative paths. Issue descriptions and comments refer to images with Markdown such as `![screenshot](attachment://att_...)`.

## CLI access

Use typed `tq` commands for routine issue, artifact, comment, project, and workflow operations. Use the allowlisted `tq api` command only when no typed command exposes the required issue-tracker operation.

The raw command has its own fail-closed method-and-route allowlist. It does not automatically expose new OpenAPI routes, and it temporarily excludes `POST /api/v1/attachments` because multipart request construction is not supported. See the [tq command reference](../references/tq.md#raw-api-requests) for syntax, validation, output, and exit-status behavior.

## Orchestrator inspection API

The orchestrator can expose a separate loopback HTTP API for runtime inspection when enabled with `--port` or `server.port`. It is not part of the issue-tracker API. Issue runtime details include historical run summaries, and a run may include `thread_id` after its Codex app-server thread has been persisted.
