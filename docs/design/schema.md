# Schema Field Specifications

This document defines the data format, constraints, and validation rules for all entities in the issue-tracker and orchestrator.

Validation is enforced at the **store layer** (Go code) on every create and update operation.

---

## issue-tracker

### Issue

| Field       | Go Type    | Required (Create) | Required (Update) | Default     | Constraints                                                        |
|-------------|------------|--------------------|--------------------|-------------|--------------------------------------------------------------------|
| ID          | `int64`    | auto               | path param         | autoincrement | `> 0`                                                            |
| ProjectID   | `int64`    | yes                | —                  | —           | `> 0`, referenced project must exist                               |
| ProjectKey  | `string`   | response           | —                  | —           | copied from referenced project                                     |
| Title       | `string`   | yes                | optional (`*string`) | —          | min 1, max 500 chars                                              |
| Description | `string`   | no                 | optional (`*string`) | `""`       | max 10,000 chars                                                  |
| Status      | `Status`   | no                 | optional (`*Status`) | `backlog`  | enum: `backlog`, `ready`, `in_progress`, `review`, `done`, `blocked`, `failed`, `cancelled`, `duplicate` |
| Priority    | `Priority` | no                 | optional (`*Priority`) | `normal` | enum: `low`, `normal`, `high`, `urgent`                           |
| Assignee    | `string`   | no                 | optional (`*string`) | `""`       | max 200 chars, free text                                          |
| DependencyIDs | `[]int64` | no               | optional (`*[]int64`) | `[]`      | full replacement when set; each ID must reference an issue; no duplicates; no self-dependency; no cycles |
| Artifacts | `[]Artifact` | response | — | `[]` | sorted by `type`; every issue response includes an array, never `null` |
| CreatedAt   | `time.Time`| auto               | —                  | `now()`     | —                                                                  |
| UpdatedAt   | `time.Time`| auto               | auto               | `now()`     | —                                                                  |

Issue descriptions may contain Markdown. Image attachments are referenced as `![alt](attachment://<attachment-id>)`.
Every issue belongs to exactly one project. Issue project ownership is set at create time and cannot be changed through update APIs.
`dependency_ids` and `artifacts` are exposed on single-issue and list responses. Issues without dependencies or artifacts return `[]`, not `null`.

### Artifact

Artifacts associate an issue with one external result. An issue can have at most one artifact for each `type`. The initial supported type is `pull_request`; its server-selected `data_type` is `url`.

| Public field | Go Type | Constraints |
|--------------|---------|-------------|
| Type | `ArtifactType` | `pull_request` |
| DataType | `ArtifactDataType` | `url` for `pull_request` |
| DataValue | `string` | trimmed absolute `http` or `https` URL; host required; userinfo forbidden; at most 4,096 UTF-8 bytes |

Artifact rows retain their creation and update timestamps internally, but those timestamps are not public fields. `(issue_id, type)` is unique. Deleting an issue cascades to its artifacts. Setting an existing type preserves `created_at` and changes only the value and `updated_at`.

Validation trims surrounding whitespace before parsing, rejects invalid UTF-8, and otherwise preserves the submitted URL unchanged. Issue responses sort artifacts by `type`.

### CreateIssueInput

`CreateIssueInput` is the request body for `POST /api/v1/issues`.

| Field         | Go Type    | Required | Default   | Constraints |
|---------------|------------|----------|-----------|-------------|
| ProjectID     | `int64`    | yes      | —         | `> 0`, referenced project must exist |
| Title         | `string`   | yes      | —         | min 1, max 500 chars |
| Description   | `string`   | no       | `""`      | max 10,000 chars |
| Status        | `Status`   | no       | `backlog` | enum: `backlog`, `ready`, `in_progress`, `review`, `done`, `blocked`, `failed`, `cancelled`, `duplicate` |
| Priority      | `Priority` | no       | `normal`  | enum: `low`, `normal`, `high`, `urgent` |
| Assignee      | `string`   | no       | `""`      | max 200 chars, free text |
| DependencyIDs | `[]int64`  | no       | `[]`      | initial dependency set; each ID must reference an issue; no duplicates; no self-dependency; no cycles |

`dependency_ids` is applied in the same transaction as issue creation. If dependency validation fails, the issue is not created.

### IssueDependency

The `issue_dependencies` table stores directed issue dependency edges. `parent_issue_id` depends on `dependency_issue_id`.

| Field             | Go Type  | Required | Default | Constraints |
|-------------------|----------|----------|---------|-------------|
| ParentIssueID      | `int64`  | yes      | —       | references `issues.id`, `ON DELETE CASCADE` |
| DependencyIssueID  | `int64`  | yes      | —       | references `issues.id`, `ON DELETE CASCADE` |
| CreatedAt          | `string` | auto     | `now()` | RFC3339Nano timestamp text |

The primary key is `(parent_issue_id, dependency_issue_id)`, so duplicate edges are rejected. A check constraint rejects direct self-dependency. Foreign key cascades remove dependency rows when either the parent issue or dependency issue is deleted.

Dependency updates are validated in the store layer before replacing the edge set. The graph must remain a DAG. Validation rejects direct self-dependency, missing dependency issues, duplicate dependency IDs, and multi-hop cycles such as `A -> B -> C -> A`. Cycle errors include the issue ID path involved in the cycle.

Queue state is derived from this table and issue status. `queued` means the issue is `ready` and all dependencies are in satisfied statuses: `done`, `cancelled`, or `duplicate`. `pending` means the issue is `ready` and at least one dependency remains in any other status, including `backlog`, `ready`, `in_progress`, `review`, `blocked`, or `failed`.

### Comment

| Field       | Go Type       | Required (Create) | Required (Update) | Default     | Constraints                                                        |
|-------------|---------------|--------------------|--------------------|-------------|--------------------------------------------------------------------|
| ID          | `int64`       | auto               | path param         | autoincrement | `> 0`                                                            |
| IssueID     | `int64`       | yes                | —                  | —           | `> 0`, referenced issue must exist                                 |
| Author      | `string`      | yes                | —                  | —           | min 1                                                              |
| Type        | `CommentType` | no                 | —                  | `general`   | enum: `progress`, `blocker`, `handoff`, `general`                  |
| Body        | `string`      | yes                | optional (`*string`) | —         | min 1, max 10,000 chars; may contain Markdown attachment refs      |
| CreatedAt   | `time.Time`   | auto               | —                  | `now()`     | —                                                                  |

### ChangeRequest

Change requests store additional user or reviewer requests that should be handled by a later agent run. They are separate from comments because they have workflow state.

| Field           | Go Type               | Required (Create) | Required (Update) | Default       | Constraints |
|-----------------|-----------------------|-------------------|-------------------|---------------|-------------|
| ID              | `int64`               | auto              | path param        | autoincrement | `> 0` |
| IssueID         | `int64`               | yes               | —                 | —             | `> 0`, referenced issue must exist |
| Author          | `string`              | yes               | —                 | —             | min 1 |
| Body            | `string`              | yes               | optional (`*string`) | —          | min 1, max 10,000 chars; editable only while `open` |
| Status          | `ChangeRequestStatus` | auto              | optional          | `open`        | enum: `open`, `in_progress`, `resolved`, `canceled` |
| CreatedAt       | `time.Time`           | auto              | —                 | `now()`       | — |
| UpdatedAt       | `time.Time`           | auto              | auto              | `now()`       | — |
| ResolvedAt      | `*time.Time`          | —                 | auto on resolve   | `NULL`        | set when status becomes `resolved` |
| ResolvedByRunID | `*string`             | —                 | optional on resolve | `NULL`      | max 200 chars; orchestrator run ID |
| ResultCommentID | `*int64`              | —                 | optional on resolve | `NULL`      | must reference a comment on the same issue |

Allowed status transitions are `open -> in_progress`, `open -> canceled`, `in_progress -> resolved`, and `in_progress -> canceled`. `resolved` and `canceled` requests are immutable. The orchestrator includes up to 20 open change requests in chronological order when continuing an issue with a previous run, then marks the included requests `in_progress`. Failed or interrupted sessions do not automatically move `in_progress` requests back to `open`.

### Attachment

| Field       | Go Type    | Required (Create) | Required (Update) | Default | Constraints                                                        |
|-------------|------------|--------------------|--------------------|---------|--------------------------------------------------------------------|
| ID          | `string`   | generated          | path param         | —       | min 1, max 80 chars                                                |
| EntityType  | `string`   | yes                | —                  | —       | enum: `issue`, `comment`                                           |
| EntityID    | `string`   | yes                | —                  | —       | min 1, max 80 chars; referenced entity must exist on upload        |
| Filename    | `string`   | yes                | —                  | —       | basename only, min 1, max 255 chars                                |
| Path        | `string`   | generated          | —                  | —       | relative path under `$TQ_HOME`, max 1,000 chars                    |
| ContentType | `string`   | yes                | —                  | —       | `image/png`, `image/jpeg`, `image/gif`, or `image/webp`            |
| Size        | `int64`    | yes                | —                  | —       | `> 0`, max 5 MiB                                                   |
| CreatedAt   | `time.Time`| auto               | —                  | `now()` | —                                                                  |

Attachment records live in SQLite, while file bytes are stored under `$TQ_HOME/system/data/attachments/{entity_type}/{entity_id}/{attachment_id}.{ext}`. The API stores only the relative path so `$TQ_HOME` can move without rewriting rows.

### Project

| Field       | Go Type    | Required (Create) | Required (Update) | Default     | Constraints                                                        |
|-------------|------------|--------------------|--------------------|-------------|--------------------------------------------------------------------|
| ID          | `int64`    | auto               | path param         | autoincrement | `> 0`                                                            |
| Key         | `string`   | yes                | optional (`*string`) | —          | regex: `^[A-Z][A-Z0-9_]{0,19}$` (1-20 chars, uppercase start)    |
| Name        | `string`   | yes                | optional (`*string`) | —          | min 1, max 200 chars                                              |
| Description | `string`   | no                 | optional (`*string`) | `""`       | max 10,000 chars                                                  |
| Location    | `string`   | yes                | optional (`*string`) | —          | absolute path (`/` prefix), max 1,000 chars, `os.Stat` directory existence check on set |
| CreatedAt   | `time.Time`| auto               | —                  | `now()`     | —                                                                  |
| UpdatedAt   | `time.Time`| auto               | auto               | `now()`     | —                                                                  |

Deleting a project cascades through issue-tracker-owned descendants: issues, issue dependency edges that reference those issues, comments, change requests, attachment records and attachment files under `$TQ_HOME/system/data/attachments`, and stored project workflow overrides. It also deletes orchestrator runtime descendants for the owned issues: runs, runner events, workspace metadata, and workspace setup failures. Project deletion is rejected when any owned issue has a `running` orchestrator run. Project deletion does not delete or modify `Location` on disk or any worktrees.

### ProjectWorkflow

| Field       | Go Type           | Required | Default | Constraints                         |
|-------------|-------------------|----------|---------|-------------------------------------|
| ProjectID   | `int64`           | yes      | —       | `> 0`, one workflow per project     |
| Frontmatter | `map[string]any`  | yes      | `{}`    | JSON object stored in SQLite        |
| Body        | `string`          | yes      | —       | Raw workflow Markdown body          |
| Checksum    | `string`          | yes      | —       | Workflow content checksum           |
| CreatedAt   | `time.Time`       | auto     | `now()` | —                                   |
| UpdatedAt   | `time.Time`       | auto     | `now()` | —                                   |

`GET /api/v1/projects/{id}/workflow` returns a `ProjectWorkflow` for projects with a stored workflow and returns 404 when no workflow row exists.
Deleting a project workflow row returns the project to file-based workflow resolution, such as the project `WORKFLOW.md` or a global fallback workflow.

## migrations

### SchemaMigration

Each SQLite database owns a `schema_migrations` table. The migration engine writes one row after a migration successfully applies and removes the row when that migration is rolled back.

| Field     | Go Type  | Required | Default             | Constraints                     |
|-----------|----------|----------|---------------------|---------------------------------|
| Version   | `string` | yes      | —                   | primary key, timestamp version  |
| Name      | `string` | yes      | —                   | migration name from filename    |
| AppliedAt | `string` | auto     | `CURRENT_TIMESTAMP` | SQLite timestamp text           |

## orchestrator

### Run

| Field          | Go Type      | Required (Create) | Required (Update) | Default     | Constraints                                                        |
|----------------|--------------|--------------------|--------------------|-------------|--------------------------------------------------------------------|
| ID             | `int64`      | auto               | —                  | autoincrement | `> 0`                                                            |
| RunID          | `string`     | auto               | —                  | `uuid.NewString()` | generated by store                                          |
| IssueID        | `int64`      | yes                | —                  | —           | `> 0`                                                              |
| Status         | `run.Status` | auto               | yes                | `queued`    | enum: `queued`, `running`, `succeeded`, `failed`, `cancelled`      |
| Workspace      | `string`     | no                 | —                  | `""`        | max 1,000 chars                                                    |
| ThreadID       | `string`     | no                 | optional           | `NULL`      | max 200 chars; stores the Codex app-server thread identifier for resume |
| Attempt        | `int`        | yes                | —                  | —           | `>= 0`                                                             |
| Error          | `string`     | no                 | —                  | `""`        | max 10,000 chars                                                   |
| OrchestratorID | `string`     | yes                | —                  | —           | min 1, max 200 chars                                               |
| CreatedAt      | `time.Time`  | auto               | —                  | `now()`     | —                                                                  |
| UpdatedAt      | `time.Time`  | auto               | auto               | `now()`     | —                                                                  |

### RunnerEvent

| Field       | Go Type     | Required | Default   | Constraints                                                        |
|-------------|-------------|----------|-----------|--------------------------------------------------------------------|
| ID          | `int64`     | auto     | autoincrement | `> 0`                                                          |
| RunID       | `string`    | yes      | —         | min 1, max 200 chars                                               |
| EventType   | `string`    | no       | `"event"` | max 200 chars (falls back to `"event"` if empty)                   |
| Message     | `string`    | no       | `""`      | max 10,000 chars                                                   |
| PayloadJSON | `string`    | no       | `""`      | max 50,000 chars; if non-empty, must pass `json.Valid`             |
| OccurredAt  | `time.Time` | yes      | —         | must not be zero (`time.IsZero()` check)                           |

### WorkspaceMetadata

| Field        | Go Type | Required (Upsert) | Default | Constraints                                                        |
|--------------|---------|--------------------|---------|--------------------------------------------------------------------|
| WorkspaceKey | `string`| yes                | —       | min 1, max 200 chars (upsert key)                                  |
| IssueID      | `int64` | yes                | —       | `> 0`                                                              |
| Path         | `string`| yes                | —       | absolute path (`/` prefix), max 1,000 chars                        |
| CreatedNow   | `bool`  | yes                | —       | —                                                                  |
| SourcePath   | `string`| no                 | `""`    | if non-empty: absolute path (`/` prefix), max 1,000 chars          |
| CreatedAt    | `time.Time` | auto           | `now()` | —                                                                  |
| UpdatedAt    | `time.Time` | auto           | `now()` | —                                                                  |

> Additional DB-only columns (`populated_at`, `cleanup_status`, `cleanup_at`, `last_error`) are managed internally by store methods and not exposed via input structs.

### WorkspaceSetupFailure

Recorded via `RecordWorkspaceSetupFailure(ctx, issueID, workspaceKey, path, errText)` — no input struct.

| Field        | Go Type | Required | Default | Constraints                                                        |
|--------------|---------|----------|---------|--------------------------------------------------------------------|
| ID           | `int64` | auto     | autoincrement | `> 0`                                                        |
| IssueID      | `int64` | yes      | —       | `> 0`                                                              |
| WorkspaceKey | `string`| no       | `""`    | max 200 chars                                                      |
| Path         | `string`| no       | `""`    | max 1,000 chars                                                    |
| Error        | `string`| yes      | —       | min 1, max 10,000 chars                                            |
| OccurredAt   | `time.Time` | auto | `now()` | —                                                                  |

---

## Validation Rules Summary

### String length

| Limit       | Fields                                                                 |
|-------------|------------------------------------------------------------------------|
| max 200     | Issue.Assignee, Project.Key, Project.Name, Workspace.Name, Run.ThreadID, Run.OrchestratorID, RunnerEvent.RunID, RunnerEvent.EventType, WorkspaceMetadata.WorkspaceKey, WorkspaceSetupFailure.WorkspaceKey |
| max 500     | Issue.Title                                                            |
| max 1,000   | Attachment.Path, Project.Location, Workspace.Path, Run.Workspace, WorkspaceMetadata.Path, WorkspaceMetadata.SourcePath, WorkspaceSetupFailure.Path |
| max 10,000  | Issue.Description, Comment.Body, Project.Description, Run.Error, RunnerEvent.Message, WorkspaceSetupFailure.Error |
| max 50,000  | RunnerEvent.PayloadJSON                                                |

### Path fields

All path fields require an absolute path (must start with `/`). API validation does not check directory existence because project paths are recorded from the client host perspective and may not be visible from the API server runtime.

Directory existence is checked by clients before creating records when they can access the target filesystem. For example, `tq project add <path>` resolves the path to a host absolute path and checks that it exists locally before sending it to the issue-tracker API.

Directory existence is not checked by the API for:
- Project.Location
- Workspace.Path
- WorkspaceMetadata.Path (may not exist at upsert time)
- WorkspaceMetadata.SourcePath
- WorkspaceSetupFailure.Path (recorded after failure)

### Enum fields

| Field            | Valid Values                                                          |
|------------------|-----------------------------------------------------------------------|
| Issue.Status     | `backlog`, `ready`, `in_progress`, `review`, `done`, `blocked`, `failed`, `cancelled`, `duplicate` |
| Issue.Priority   | `low`, `normal`, `high`, `urgent`                                     |
| Comment.Type     | `progress`, `blocker`, `handoff`, `general`                           |
| ChangeRequest.Status | `open`, `in_progress`, `resolved`, `canceled`                    |
| Attachment.EntityType | `issue`, `comment`                                                |
| Attachment.ContentType | `image/png`, `image/jpeg`, `image/gif`, `image/webp`             |
| Workspace.Status | `active`, `inactive`, `archived`                                      |
| Run.Status       | `queued`, `running`, `succeeded`, `failed`, `cancelled`               |

### Format constraints

| Field       | Pattern                          | Description                              |
|-------------|----------------------------------|------------------------------------------|
| Project.Key | `^([A-Z][A-Z0-9_]{0,19}\|[a-z][a-z0-9-]{0,63})$` | 1-20 chars uppercase legacy key (`A-Z`, `0-9`, `_`) or 1-64 chars lowercase kebab-case key |
