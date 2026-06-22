# スキーマフィールド仕様

この文書は、issue-tracker と orchestrator の全エンティティについて、データ形式、制約、検証ルールを定義します。

検証は、すべての create / update 操作で **store layer**（Go code）によって強制されます。

開発判断の一次情報は [schema.md](schema.md) です。この日本語版は同じ構成と内容を同期して保持します。

---

## issue-tracker

### Issue

| Field       | Go Type    | Required (Create) | Required (Update) | Default     | Constraints                                                        |
|-------------|------------|--------------------|--------------------|-------------|--------------------------------------------------------------------|
| ID          | `int64`    | auto               | path param         | autoincrement | `> 0`                                                            |
| ProjectID   | `int64`    | yes                | —                  | —           | `> 0`、参照先 project が存在すること                               |
| ProjectKey  | `string`   | response           | —                  | —           | 参照先 project からコピーされます                                  |
| Title       | `string`   | yes                | optional (`*string`) | —          | min 1, max 500 chars                                              |
| Description | `string`   | no                 | optional (`*string`) | `""`       | max 10,000 chars                                                  |
| Status      | `Status`   | no                 | optional (`*Status`) | `backlog`  | enum: `backlog`, `ready`, `in_progress`, `review`, `done`, `blocked`, `failed` |
| Priority    | `Priority` | no                 | optional (`*Priority`) | `normal` | enum: `low`, `normal`, `high`, `urgent`                           |
| Assignee    | `string`   | no                 | optional (`*string`) | `""`       | max 200 chars、free text                                          |
| CreatedAt   | `time.Time`| auto               | —                  | `now()`     | —                                                                  |
| UpdatedAt   | `time.Time`| auto               | auto               | `now()`     | —                                                                  |

Issue description は Markdown を含められます。画像 attachment は `![alt](attachment://<attachment-id>)` として参照されます。
すべての issue は必ず 1 つの project に属します。Issue の project ownership は create 時に設定され、update API では変更できません。

### Comment

| Field       | Go Type       | Required (Create) | Required (Update) | Default     | Constraints                                                        |
|-------------|---------------|--------------------|--------------------|-------------|--------------------------------------------------------------------|
| ID          | `int64`       | auto               | path param         | autoincrement | `> 0`                                                            |
| IssueID     | `int64`       | yes                | —                  | —           | `> 0`、参照先 issue が存在すること                                 |
| Author      | `string`      | yes                | —                  | —           | min 1                                                              |
| Type        | `CommentType` | no                 | —                  | `general`   | enum: `progress`, `blocker`, `handoff`, `general`                  |
| Body        | `string`      | yes                | optional (`*string`) | —         | min 1, max 10,000 chars、Markdown attachment refs を含められます   |
| CreatedAt   | `time.Time`   | auto               | —                  | `now()`     | —                                                                  |

### Attachment

| Field       | Go Type    | Required (Create) | Required (Update) | Default | Constraints                                                        |
|-------------|------------|--------------------|--------------------|---------|--------------------------------------------------------------------|
| ID          | `string`   | generated          | path param         | —       | min 1, max 80 chars                                                |
| EntityType  | `string`   | yes                | —                  | —       | enum: `issue`, `comment`                                           |
| EntityID    | `string`   | yes                | —                  | —       | min 1, max 80 chars、参照先 entity が upload 時点で存在すること    |
| Filename    | `string`   | yes                | —                  | —       | basename only, min 1, max 255 chars                                |
| Path        | `string`   | generated          | —                  | —       | `$TQ_HOME` 配下の相対 path、max 1,000 chars                        |
| ContentType | `string`   | yes                | —                  | —       | `image/png`, `image/jpeg`, `image/gif`, or `image/webp`            |
| Size        | `int64`    | yes                | —                  | —       | `> 0`, max 5 MiB                                                   |
| CreatedAt   | `time.Time`| auto               | —                  | `now()` | —                                                                  |

Attachment records は SQLite に保存され、file bytes は `$TQ_HOME/system/data/attachments/{entity_type}/{entity_id}/{attachment_id}.{ext}` に保存されます。API は相対 path だけを保存するため、`$TQ_HOME` を移動しても row の書き換えは不要です。

### Project

| Field       | Go Type    | Required (Create) | Required (Update) | Default     | Constraints                                                        |
|-------------|------------|--------------------|--------------------|-------------|--------------------------------------------------------------------|
| ID          | `int64`    | auto               | path param         | autoincrement | `> 0`                                                            |
| Key         | `string`   | yes                | optional (`*string`) | —          | regex: `^[A-Z][A-Z0-9_]{0,19}$` (1-20 chars, uppercase start)    |
| Name        | `string`   | yes                | optional (`*string`) | —          | min 1, max 200 chars                                              |
| Description | `string`   | no                 | optional (`*string`) | `""`       | max 10,000 chars                                                  |
| Location    | `string`   | yes                | optional (`*string`) | —          | absolute path (`/` prefix), max 1,000 chars、set 時に `os.Stat` で directory existence check |
| CreatedAt   | `time.Time`| auto               | —                  | `now()`     | —                                                                  |
| UpdatedAt   | `time.Time`| auto               | auto               | `now()`     | —                                                                  |

Projects は linked issues が存在する間は削除できません。

### ProjectWorkflow

| Field       | Go Type           | Required | Default | Constraints                         |
|-------------|-------------------|----------|---------|-------------------------------------|
| ProjectID   | `int64`           | yes      | —       | `> 0`、one workflow per project     |
| Frontmatter | `map[string]any`  | yes      | `{}`    | JSON object stored in SQLite        |
| Body        | `string`          | yes      | —       | Raw workflow Markdown body          |
| Checksum    | `string`          | yes      | —       | Workflow content checksum           |
| CreatedAt   | `time.Time`       | auto     | `now()` | —                                   |
| UpdatedAt   | `time.Time`       | auto     | `now()` | —                                   |

`GET /api/v1/projects/{id}/workflow` は stored workflow を持つ project に対して `ProjectWorkflow` を返し、workflow row がない場合は 404 を返します。
Project workflow row を削除すると、その project は project-local `WORKFLOW.md` や global fallback workflow などの file-based resolution に戻ります。

## migrations

### SchemaMigration

各 SQLite database はそれぞれ `schema_migrations` table を所有します。Migration engine は migration が正常に apply された後に row を 1 つ書き込み、その migration が rollback されたときに row を削除します。

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
| RunID          | `string`     | auto               | —                  | `uuid.NewString()` | store が生成します                                           |
| IssueID        | `int64`      | yes                | —                  | —           | `> 0`                                                              |
| Status         | `run.Status` | auto               | yes                | `queued`    | enum: `queued`, `running`, `succeeded`, `failed`, `cancelled`      |
| Workspace      | `string`     | no                 | —                  | `""`        | max 1,000 chars                                                    |
| ThreadID       | `string`     | no                 | optional           | `NULL`      | max 200 chars、resume 用の Codex app-server thread identifier を保存します |
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
| EventType   | `string`    | no       | `"event"` | max 200 chars、empty の場合は `"event"` に fallback                |
| Message     | `string`    | no       | `""`      | max 10,000 chars                                                   |
| PayloadJSON | `string`    | no       | `""`      | max 100,000 chars、must be valid JSON object/array/string/number/bool/null when non-empty |
| OccurredAt  | `time.Time` | auto     | `now()`   | must be non-zero                                                   |

### WorkspaceMetadata

| Field         | Go Type     | Required | Default | Constraints                                                        |
|---------------|-------------|----------|---------|--------------------------------------------------------------------|
| WorkspaceKey  | `string`    | yes      | —       | min 1, max 200 chars, primary key                                  |
| IssueID       | `int64`     | yes      | —       | `> 0`                                                              |
| Path          | `string`    | yes      | —       | min 1, max 1,000 chars                                             |
| CreatedNow    | `bool`      | yes      | `false` | stored as `0` / `1`                                                |
| SourcePath    | `string`    | no       | `""`    | max 1,000 chars                                                    |
| PopulatedAt   | `time.Time` | auto     | `now()` | stored as RFC3339Nano string                                       |
| CleanupStatus | `string`    | no       | `""`    | max 200 chars                                                      |
| CleanupAt     | `string`    | no       | `""`    | timestamp string or empty                                          |
| LastError     | `string`    | no       | `""`    | max 10,000 chars                                                   |
| CreatedAt     | `time.Time` | auto     | `now()` | —                                                                  |
| UpdatedAt     | `time.Time` | auto     | `now()` | —                                                                  |

Workspace metadata は `WorkspaceKey` で upsert されます。新しい workspace population が成功すると、最新の usable workspace state が見えるように `cleanup_status`、`cleanup_at`、`last_error` は clear されます。

### WorkspaceSetupFailure

| Field        | Go Type     | Required | Default | Constraints                         |
|--------------|-------------|----------|---------|-------------------------------------|
| ID           | `int64`     | auto     | autoincrement | `> 0`                         |
| IssueID      | `int64`     | yes      | —       | `> 0`                               |
| WorkspaceKey | `string`    | no       | `""`    | max 200 chars                       |
| Path         | `string`    | no       | `""`    | max 1,000 chars                     |
| Error        | `string`    | yes      | —       | min 1, max 10,000 chars             |
| OccurredAt   | `time.Time` | auto     | `now()` | stored as RFC3339Nano string        |

## 文字列長の要点

| Limit       | Fields                                                                 |
|-------------|------------------------------------------------------------------------|
| max 80      | Attachment.ID, Attachment.EntityID                                     |
| max 200     | Issue.Assignee, Project.Key, Project.Name, Workspace.Name, Run.ThreadID, Run.OrchestratorID, RunnerEvent.RunID, RunnerEvent.EventType, WorkspaceMetadata.WorkspaceKey, WorkspaceSetupFailure.WorkspaceKey |
| max 255     | Attachment.Filename                                                    |
| max 500     | Issue.Title                                                            |
| max 1,000   | Project.Location, Attachment.Path, Run.Workspace, WorkspaceMetadata.Path, WorkspaceMetadata.SourcePath, WorkspaceSetupFailure.Path |
| max 10,000  | Issue.Description, Comment.Body, Project.Description, Run.Error, RunnerEvent.Message, WorkspaceMetadata.LastError, WorkspaceSetupFailure.Error |
| max 100,000 | RunnerEvent.PayloadJSON                                                |

## enum の要点

| Enum        | Values                                                                 |
|-------------|------------------------------------------------------------------------|
| Status      | `backlog`, `ready`, `in_progress`, `review`, `done`, `blocked`, `failed` |
| Priority    | `low`, `normal`, `high`, `urgent`                                      |
| CommentType | `progress`, `blocker`, `handoff`, `general`                            |
| RunStatus   | `queued`, `running`, `succeeded`, `failed`, `cancelled`                 |
