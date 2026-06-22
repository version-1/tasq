# スキーマフィールド仕様

この文書は、issue-tracker と orchestrator のエンティティのデータ形式、制約、検証ルールを定義します。

開発判断の一次情報は [schema.md](schema.md) です。この日本語版は同期対象の要点を保持します。

## orchestrator

### Run

| Field          | Go Type      | Required (Create) | Required (Update) | Default     | Constraints                                                        |
|----------------|--------------|--------------------|--------------------|-------------|--------------------------------------------------------------------|
| ID             | `int64`      | auto               | —                  | autoincrement | `> 0`                                                            |
| RunID          | `string`     | auto               | —                  | `uuid.NewString()` | store が生成します                                           |
| IssueID        | `int64`      | yes                | —                  | —           | `> 0`                                                              |
| Status         | `run.Status` | auto               | yes                | `queued`    | enum: `queued`, `running`, `succeeded`, `failed`, `cancelled`      |
| Workspace      | `string`     | no                 | —                  | `""`        | 最大 1,000 文字                                                    |
| ThreadID       | `string`     | no                 | optional           | `NULL`      | 最大 200 文字。resume 用の Codex app-server thread 識別子を保持します |
| Attempt        | `int`        | yes                | —                  | —           | `>= 0`                                                             |
| Error          | `string`     | no                 | —                  | `""`        | 最大 10,000 文字                                                   |
| OrchestratorID | `string`     | yes                | —                  | —           | 1-200 文字                                                         |
| CreatedAt      | `time.Time`  | auto               | —                  | `now()`     | —                                                                  |
| UpdatedAt      | `time.Time`  | auto               | auto               | `now()`     | —                                                                  |

### 文字列長の要点

| Limit       | Fields                                                                 |
|-------------|------------------------------------------------------------------------|
| max 200     | Issue.Assignee, Project.Key, Project.Name, Workspace.Name, Run.ThreadID, Run.OrchestratorID, RunnerEvent.RunID, RunnerEvent.EventType, WorkspaceMetadata.WorkspaceKey, WorkspaceSetupFailure.WorkspaceKey |

