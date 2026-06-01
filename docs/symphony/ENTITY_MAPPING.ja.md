# Symphony SPEC に対する Tasq エンティティマッピング

この文書は、Symphony SPEC のドメインモデル概念と Tasq エンティティの対応関係を示します。Tasq に取り組む開発者が、どの Tasq 型がどの SPEC 概念に対応するか、またそれらがどう関係するかを確認するための参照です。

SPEC との差異と理由については [DEVIATIONS.md](DEVIATIONS.md) を参照してください。

## Project

### SPEC との対応

Symphony SPEC は Project エンティティを定義していません。最も近い概念は `tracker.project_slug`（SPEC §5.3.1）で、外部トラッカー（Linear）に対する issue クエリを単一 project にスコープします。

### Tasq 型

`internal/issue/domain/entity/entity.go` の `entity.Project`

### 所有関係

- Project は 0 個以上の Workspace（`entity.Workspace`）を所有します。
- 現在のスキーマでは、Issue は Project から独立して存在します。

### 用途

- issue-tracker service によって管理されます。
- `Project.Key`（例: `TASQ`）は CLI と API のコンテキストで project を識別します。
- orchestrator からは参照されません。orchestrator は関連する issue filtering が適用された issue-tracker の issue API を使うため、project scope を必要としません。

### 関係

- `entity.Workspace.ProjectID` → `entity.Project.ID`

## entity.Workspace（Issue Tracker）

### SPEC との対応

直接対応する SPEC 概念はありません。SPEC の Workspace（§4.1.4）は orchestrator 側の runtime concept です。`entity.Workspace` は issue-tracker service が所有する Tasq 固有の管理レコードです。

### Tasq 型

`internal/issue/domain/entity/entity.go` の `entity.Workspace`

### 所有関係

- 1 つの Project（`ProjectID`）に属します。

### 用途

- CRUD 操作用に issue-tracker service によって管理されます。
- 名前付き workspace の metadata として、名前、filesystem path、lifecycle status（`active`, `inactive`, `archived`）を保持します。
- orchestrator が issue ごとの workspace を作成または管理するときには使われません。

### 関係

- `entity.Workspace.ProjectID` → `entity.Project.ID`

## workspace.Workspace（Orchestrator）

### SPEC との対応

SPEC §4.1.4 Workspace に直接対応します。

| SPEC Field      | Tasq Field     |
|-----------------|----------------|
| `path`          | `Path`         |
| `workspace_key` | `WorkspaceKey` |
| `created_now`   | `CreatedNow`   |

### Tasq 型

`internal/orchestrator/workspace/workspace.go` の `workspace.Workspace`

### 所有関係

- `workspace.Manager` によって作成および管理されます。
- `entity.Project` と `entity.Workspace` からは独立しています。

### 用途

- orchestrator が agent work を dispatch するときに、issue ごとに作成されます。
- workspace key は `issue-<ID>` から導出されます（[DEVIATIONS.md](DEVIATIONS.md#workspace-key) を参照）。
- workspace directory は `<workspace.root>/<sanitized_key>` です。
- lifecycle hook（`after_create`, `before_run`, `after_run`, `before_remove`）は workspace directory で実行されます。hook configuration については [WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md) を参照してください。
- 同じ issue の run 間で再利用されます。issue が terminal state に到達したときに cleanup されます。

### 関係

- `workspace.Manager` は `workflow.Config` 由来の workspace root path と hook configuration を保持します。
- `entity.Project` または `entity.Workspace` への foreign key はありません。
