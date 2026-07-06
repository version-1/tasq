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
| ProjectKey  | `string`   | response           | —                  | —           | 参照先 project からコピーされること                                |
| Title       | `string`   | yes                | optional (`*string`) | —          | min 1, max 500 chars                                              |
| Description | `string`   | no                 | optional (`*string`) | `""`       | max 10,000 chars                                                  |
| Status      | `Status`   | no                 | optional (`*Status`) | `backlog`  | enum: `backlog`, `ready`, `in_progress`, `review`, `done`, `blocked`, `failed`, `cancelled`, `duplicate` |
| Priority    | `Priority` | no                 | optional (`*Priority`) | `normal` | enum: `low`, `normal`, `high`, `urgent`                           |
| Assignee    | `string`   | no                 | optional (`*string`) | `""`       | max 200 chars、自由入力                                           |
| DependencyIDs | `[]int64` | no               | optional (`*[]int64`) | `[]`      | 指定時は全置換。各 ID は issue を参照すること。重複、自己依存、cycle は不可 |
| CreatedAt   | `time.Time`| auto               | —                  | `now()`     | —                                                                  |
| UpdatedAt   | `time.Time`| auto               | auto               | `now()`     | —                                                                  |

課題の説明には Markdown を含められます。画像添付ファイルは `![alt](attachment://<attachment-id>)` として参照されます。
すべての課題は必ず 1 つのプロジェクトに属します。課題が属するプロジェクトは作成時に設定され、update API では変更できません。
`dependency_ids` は単一 issue response と list response の両方で公開されます。依存がない issue は `null` ではなく `[]` を返します。

### CreateIssueInput

`CreateIssueInput` は `POST /api/v1/issues` の request body です。

| Field         | Go Type    | Required | Default   | Constraints |
|---------------|------------|----------|-----------|-------------|
| ProjectID     | `int64`    | yes      | —         | `> 0`、参照先 project が存在すること |
| Title         | `string`   | yes      | —         | min 1, max 500 chars |
| Description   | `string`   | no       | `""`      | max 10,000 chars |
| Status        | `Status`   | no       | `backlog` | enum: `backlog`, `ready`, `in_progress`, `review`, `done`, `blocked`, `failed`, `cancelled`, `duplicate` |
| Priority      | `Priority` | no       | `normal`  | enum: `low`, `normal`, `high`, `urgent` |
| Assignee      | `string`   | no       | `""`      | max 200 chars、自由入力 |
| DependencyIDs | `[]int64`  | no       | `[]`      | 初期 dependency set。各 ID は issue を参照すること。重複、自己依存、cycle は不可 |

`dependency_ids` は issue creation と同じ transaction で適用されます。dependency validation が失敗した場合、issue は作成されません。

### IssueDependency

`issue_dependencies` table は、issue 間の directed dependency edge を保存します。`parent_issue_id` が `dependency_issue_id` に依存します。

| Field             | Go Type  | Required | Default | Constraints |
|-------------------|----------|----------|---------|-------------|
| ParentIssueID      | `int64`  | yes      | —       | `issues.id` を参照する。`ON DELETE CASCADE` |
| DependencyIssueID  | `int64`  | yes      | —       | `issues.id` を参照する。`ON DELETE CASCADE` |
| CreatedAt          | `string` | auto     | `now()` | RFC3339Nano timestamp text |

primary key は `(parent_issue_id, dependency_issue_id)` で、重複 edge は拒否されます。check constraint は直接の自己依存を拒否します。foreign key cascade により、parent issue または dependency issue が削除されたとき dependency row も削除されます。

dependency update は edge set を置き換える前に store layer で検証されます。graph は DAG でなければなりません。検証は直接の自己依存、存在しない dependency issue、重複した dependency ID、`A -> B -> C -> A` のような multi-hop cycle を拒否します。cycle error には cycle に含まれる issue ID path が含まれます。

queue state はこの table と issue status から導出されます。`queued` は課題が `ready` で、すべての依存先が満たされた status（`done`, `cancelled`, `duplicate`）にあることを意味します。`pending` は課題が `ready` だが、少なくとも 1 件の依存先がそれ以外の status（`backlog`, `ready`, `in_progress`, `review`, `blocked`, `failed` など）に残っていることを意味します。

### Comment

| Field       | Go Type       | Required (Create) | Required (Update) | Default     | Constraints                                                        |
|-------------|---------------|--------------------|--------------------|-------------|--------------------------------------------------------------------|
| ID          | `int64`       | auto               | path param         | autoincrement | `> 0`                                                            |
| IssueID     | `int64`       | yes                | —                  | —           | `> 0`、参照先 issue が存在すること                                 |
| Author      | `string`      | yes                | —                  | —           | min 1                                                              |
| Type        | `CommentType` | no                 | —                  | `general`   | enum: `progress`, `blocker`, `handoff`, `general`                  |
| Body        | `string`      | yes                | optional (`*string`) | —         | min 1, max 10,000 chars、Markdown attachment refs を含められること |
| CreatedAt   | `time.Time`   | auto               | —                  | `now()`     | —                                                                  |

### Attachment

| Field       | Go Type    | Required (Create) | Required (Update) | Default | Constraints                                                        |
|-------------|------------|--------------------|--------------------|---------|--------------------------------------------------------------------|
| ID          | `string`   | generated          | path param         | —       | min 1, max 80 chars                                                |
| EntityType  | `string`   | yes                | —                  | —       | enum: `issue`, `comment`                                           |
| EntityID    | `string`   | yes                | —                  | —       | min 1, max 80 chars、upload 時点で参照先 entity が存在すること     |
| Filename    | `string`   | yes                | —                  | —       | basename only, min 1, max 255 chars                                |
| Path        | `string`   | generated          | —                  | —       | `$TQ_HOME` 配下の相対 path、max 1,000 chars                        |
| ContentType | `string`   | yes                | —                  | —       | `image/png`, `image/jpeg`, `image/gif`, or `image/webp`            |
| Size        | `int64`    | yes                | —                  | —       | `> 0`, max 5 MiB                                                   |
| CreatedAt   | `time.Time`| auto               | —                  | `now()` | —                                                                  |

添付ファイルレコードは SQLite に保存され、ファイルのバイト列は `$TQ_HOME/system/data/attachments/{entity_type}/{entity_id}/{attachment_id}.{ext}` 配下に保存されます。API は相対パスだけを保存するため、`$TQ_HOME` を移動しても row の書き換えは不要です。

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

project を削除すると、issue-tracker が所有する子孫データも cascade 削除されます。対象は issues、その issues を参照する issue dependency edges、comments、attachment records と `$TQ_HOME/system/data/attachments` 配下の attachment files、保存済み project workflow overrides です。所有する issues に紐づく orchestrator runtime の子孫データである runs、runner events、workspace metadata、workspace setup failures も削除します。所有 issue に `running` の orchestrator run が 1 件でもある場合、project 削除は拒否されます。project 削除はディスク上の `Location` や worktrees を削除・変更しません。

### ProjectWorkflow

| Field       | Go Type           | Required | Default | Constraints                         |
|-------------|-------------------|----------|---------|-------------------------------------|
| ProjectID   | `int64`           | yes      | —       | `> 0`、1 project につき 1 workflow  |
| Frontmatter | `map[string]any`  | yes      | `{}`    | SQLite に保存される JSON object     |
| Body        | `string`          | yes      | —       | 生の workflow Markdown body         |
| Checksum    | `string`          | yes      | —       | workflow content checksum           |
| CreatedAt   | `time.Time`       | auto     | `now()` | —                                   |
| UpdatedAt   | `time.Time`       | auto     | `now()` | —                                   |

`GET /api/v1/projects/{id}/workflow` は、stored workflow を持つ project に対して `ProjectWorkflow` を返し、workflow row がない場合は 404 を返します。
project workflow row を削除すると、その project は project `WORKFLOW.md` や global fallback workflow などの file-based workflow resolution に戻ります。

## migrations

### SchemaMigration

各 SQLite database はそれぞれ `schema_migrations` table を所有します。migration engine は migration が正常に適用された後に row を 1 つ書き込み、その migration が rollback されたときに row を削除します。

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
| RunID          | `string`     | auto               | —                  | `uuid.NewString()` | store が生成すること                                        |
| IssueID        | `int64`      | yes                | —                  | —           | `> 0`                                                              |
| Status         | `run.Status` | auto               | yes                | `queued`    | enum: `queued`, `running`, `succeeded`, `failed`, `cancelled`      |
| Workspace      | `string`     | no                 | —                  | `""`        | max 1,000 chars                                                    |
| ThreadID       | `string`     | no                 | optional           | `NULL`      | max 200 chars、resume 用の Codex app-server thread identifier を保存すること |
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
| PayloadJSON | `string`    | no       | `""`      | max 50,000 chars、non-empty の場合は `json.Valid` に合格すること   |
| OccurredAt  | `time.Time` | yes      | —         | zero でないこと（`time.IsZero()` check）                           |

### WorkspaceMetadata

| Field        | Go Type | Required (Upsert) | Default | Constraints                                                        |
|--------------|---------|--------------------|---------|--------------------------------------------------------------------|
| WorkspaceKey | `string`| yes                | —       | min 1, max 200 chars（upsert key）                                 |
| IssueID      | `int64` | yes                | —       | `> 0`                                                              |
| Path         | `string`| yes                | —       | absolute path (`/` prefix), max 1,000 chars                        |
| CreatedNow   | `bool`  | yes                | —       | —                                                                  |
| SourcePath   | `string`| no                 | `""`    | non-empty の場合: absolute path (`/` prefix), max 1,000 chars      |
| CreatedAt    | `time.Time` | auto           | `now()` | —                                                                  |
| UpdatedAt    | `time.Time` | auto           | `now()` | —                                                                  |

> Additional DB-only columns (`populated_at`, `cleanup_status`, `cleanup_at`, `last_error`) are managed internally by store methods and not exposed via input structs.

### WorkspaceSetupFailure

`RecordWorkspaceSetupFailure(ctx, issueID, workspaceKey, path, errText)` で記録されます。入力 struct はありません。

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

すべての path field は絶対 path（`/` で始まること）を要求します。project path は client host から見た path として記録され、API server runtime から見えない場合があるため、API の検証では directory の存在を確認しません。

対象の filesystem にアクセスできる client は、record 作成前に directory の存在を確認します。たとえば、`tq project add <path>` は path を host の絶対 path に解決し、issue-tracker API に送信する前にローカルで存在確認を行います。

API は次の項目について directory の存在を確認しません。

- Project.Location
- Workspace.Path
- WorkspaceMetadata.Path（upsert 時点では存在しない場合があります）
- WorkspaceMetadata.SourcePath
- WorkspaceSetupFailure.Path（失敗後に記録されます）

### Enum fields

| Field            | Valid Values                                                          |
|------------------|-----------------------------------------------------------------------|
| Issue.Status     | `backlog`, `ready`, `in_progress`, `review`, `done`, `blocked`, `failed`, `cancelled`, `duplicate` |
| Issue.Priority   | `low`, `normal`, `high`, `urgent`                                     |
| Comment.Type     | `progress`, `blocker`, `handoff`, `general`                           |
| Attachment.EntityType | `issue`, `comment`                                                |
| Attachment.ContentType | `image/png`, `image/jpeg`, `image/gif`, `image/webp`             |
| Workspace.Status | `active`, `inactive`, `archived`                                      |
| Run.Status       | `queued`, `running`, `succeeded`, `failed`, `cancelled`               |

### Format constraints

| Field       | Pattern                          | Description                              |
|-------------|----------------------------------|------------------------------------------|
| Project.Key | `^([A-Z][A-Z0-9_]{0,19}\|[a-z][a-z0-9-]{0,63})$` | 1-20 chars uppercase legacy key (`A-Z`, `0-9`, `_`) or 1-64 chars lowercase kebab-case key |
