# Tasq API

このドキュメントでは、ユーザー向けの issue-tracker API を扱います。所有境界とコンポーネントの責務は [architecture.ja.md](architecture.ja.md) を参照してください。ローカル開発と検証は [operations.ja.md](operations.ja.md) を参照してください。

## API Surface

issue-tracker はユーザー向け API です。

現在の issue-tracker エンドポイント:

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
- `GET /api/v1/issues/{issueId}/change-requests`
- `POST /api/v1/issues/{issueId}/change-requests`
- `GET /api/v1/change-requests/{id}`
- `PATCH /api/v1/change-requests/{id}`
- `POST /api/v1/change-requests/{id}/cancel`
- `GET /api/v1/attachments`
- `POST /api/v1/attachments`
- `GET /api/v1/attachments/{id}/content`
- `DELETE /api/v1/attachments/{id}`

`DELETE /api/v1/projects/{id}` は project と、その project が所有する issue-tracker 上の子孫データを削除します。対象は issues、その issues を参照する issue dependency edges、comments、change requests、attachment records と `$TQ_HOME/system/data/attachments` 配下の attachment files、保存済み project workflow overrides です。所有する issues に紐づく orchestrator runtime の子孫データである runs、runner events、workspace metadata、workspace setup failures も削除します。所有 issue に `running` の orchestrator run が 1 件でもある場合、endpoint は `409 Conflict` と `projects.delete.running_runs` を返し、issue-tracker と orchestrator のどちらのレコードも削除しません。途中失敗後も issue IDs を使って再実行できるように、orchestrator runtime records を issue-tracker records より先に削除します。`project.location` に記録されたユーザーの project directory や worktrees は削除・変更しません。

添付ファイルのアップロードは、`entity_type`、`entity_id`、`file` を持つ multipart form data を受け取ります。最初の実装では PNG、JPEG、GIF、WebP の画像ファイルを 5 MiB までサポートします。添付ファイルのバイト列は `$TQ_HOME/system/data/attachments` 配下に保存し、SQLite にはメタデータと相対パスを保存します。課題とコメントの本文は、`![screenshot](attachment://att_...)` のような Markdown image link で添付ファイルを参照します。

課題は必ず 1 つのプロジェクトに属します。`POST /api/v1/issues` は `projectId` を必須とし、初期の依存関係として `dependency_ids` を受け取ります。課題のレスポンスは `projectId` と `projectKey` の両方を返します。課題レスポンスには `dependency_ids` が含まれます。依存がない場合は空配列を返します。`GET /api/v1/issues` は任意の query parameter として `states`、`project_id`、`project_ids`、`priorities`、`assignee`、`search`、`limit`、`offset`、`sort_by`、`sort_direction` を受け取ります。project filter を省略した場合は、すべてのプロジェクトの課題を一覧表示します。`project_id` は 1 つのプロジェクトに絞り込み、`project_ids` はテーブルフィルター用にカンマ区切りの複数プロジェクトを受け取ります。`priorities` はカンマ区切りの優先度を受け取ります。`search` は課題 ID の完全一致と、課題タイトルの大文字小文字を区別しない部分一致で検索します。数値の search text は、完全一致する ID またはその文字列を含むタイトルに一致します。空文字または空白のみの `search` は無視します。検索は他のフィルターと組み合わせ、sorting と pagination より前に適用します。`sort_by` は `id`、`priority`、`created_at`、`updated_at` のみ、`sort_direction` は `asc`、`desc` のみ受け付けます。

`GET /api/v1/summary` は課題ボードのカラムを返します。各課題サマリーには、システム全体のキューから見た課題の状態である `queueStatus` が含まれます。`queueStatus=backlog` は課題の status が `backlog` の状態です。`queueStatus=pending` は課題が `ready` だが、ブロック中の依存先が 1 件以上残っている状態です。`queueStatus=queued` は課題が `ready` で、ブロック中の依存先がない状態です。`queueStatus=processing` は課題の status が `in_progress` の状態です。`queueStatus=completed` は課題の status が `done` の状態です。`queueStatus=inactive` はキュー処理の対象外で、`review`、`blocked`、`failed`、`cancelled`、`duplicate` を含みます。status の定義と想定される遷移は [status.ja.md](status.ja.md) を参照してください。

change request は、issue に対するユーザーまたは reviewer からの追加依頼です。workflow state を持つため comments とは分離します。`POST /api/v1/issues/{issueId}/change-requests` は `open` request を作成します。`GET /api/v1/issues/{issueId}/change-requests` は issue の requests を一覧し、任意の `status` と `limit` query parameters を受け取ります。`PATCH /api/v1/change-requests/{id}` は open request の本文編集または status 更新を行います。本文編集は request が `open` の間だけ許可します。`POST /api/v1/change-requests/{id}/cancel` は request を `canceled` に移します。物理 delete endpoint は公開しません。

許可される change request の遷移は `open -> in_progress`、`open -> canceled`、`in_progress -> resolved`、`in_progress -> canceled` です。`resolved` と `canceled` は immutable です。orchestrator は、前回 run を持つ issue の継続作業を開始するとき、`open` change requests を時系列で最大 20 件取得し、含めた request を `in_progress` に移して Codex continuation guidance に含めます。guidance は、対応済み request を `resolved` にし、`resolvedByRunId` に orchestrator run ID を設定し、結果 comment がある場合は `resultCommentId` を設定するよう agent に指示します。

### `POST /api/v1/issues`

プロジェクト内に課題を作成します。`dependency_ids` は任意で、同じ create 操作の中で初期 dependency issue IDs を設定します。`dependency_ids` を省略するか空配列を渡すと、依存関係のない課題を作成します。API は存在しない dependency issue、自己依存、重複した dependency ID、dependency cycle を拒否します。

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

`PATCH /api/v1/issues/{id}` は、任意の full replacement field として `dependency_ids` を受け取ります。省略した場合、既存の依存関係は維持されます。空配列を渡すと、すべての依存関係を削除します。API は存在しない dependency issue、自己依存、重複した dependency ID、dependency cycle を作る更新を拒否します。

`GET /api/v1/queue` は `ready` の課題を `queued` と `pending` の配列に分けて返します。`queued` は `ready` かつブロック中の依存先がない課題です。`pending` は `ready` だが、ブロック中の依存先が 1 件以上残っている課題です。満たされた依存先の status は `done`、`cancelled`、`duplicate` です。それ以外の依存先 status は課題を `pending` に残します。各配列は priority desc（`urgent`, `high`, `normal`, `low`）と ID asc で並びます。この endpoint は課題一覧と同じ `project_id` filter semantics を受け取ります。`pending` の項目には、pending の原因になっている依存先の `blocked_dependency_ids` が含まれます。

JSON の成功レスポンスは `{ "data": ..., "meta": {} }` を使います。JSON のエラーレスポンスは `{ "error": { "code": "...", "message": "..." }, "meta": {} }` を使います。

`tq` CLI は課題 CRUD エンドポイントを次のコマンドでラップします。

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

`tq` は既定では人が読みやすい出力を使い、`--output json` が指定された場合は JSON 出力を使います。

`tq api <method> <path>` は、同じ issue-tracker ベース URL 解決を使う、制約付きの生 API 呼び出しです。method と route template の許可リストは CLI 内で管理し、API に route が増えても fail-closed になります。現在の許可リストは上記 endpoint のうち一時的に除外する `POST /api/v1/attachments` 以外を対象にします。attachment の `PATCH` は公開しません。エンコードされていない厳格な `/api/v1/...` path だけを受け付け、method は大文字に正規化します。path 内の生 query と、指定順に追加する繰り返し指定可能な `--query key=value` を使えます。`--header 'Name: value'` も繰り返し指定でき、同名は最後の値を使用します。transport が管理する header は拒否します。`--data value|@file|-` は `POST`、`PUT`、`PATCH` に限定し、content type を省略した場合は JSON を使います。

このコマンドは redirect を追跡せず、破壊的な操作でも確認を求めません。timeout は 10 秒で、envelope の解析や出力変換を行わずにレスポンスのバイト列をコピーします。HTTP `2xx` は終了ステータス `0`、受信した `3xx`-`5xx` レスポンスは本文コピー後に `1`、transport 失敗は `1`、入力・許可リストのエラーは `2` です。

orchestrator は、`--port` または `server.port` で有効化したときに、実行時調査用の任意の loopback HTTP API を公開します。課題の実行時詳細レスポンスには過去の実行サマリーが含まれます。各実行は、Codex app-server thread が永続化された後に `thread_id` を含む場合があります。
