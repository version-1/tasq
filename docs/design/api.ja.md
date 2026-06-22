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
- `GET /api/v1/issues/{id}`
- `PATCH /api/v1/issues/{id}`
- `GET /api/v1/issues/{issueId}/comments`
- `POST /api/v1/issues/{issueId}/comments`
- `PATCH /api/v1/comments/{id}`
- `GET /api/v1/attachments`
- `POST /api/v1/attachments`
- `GET /api/v1/attachments/{id}/content`
- `DELETE /api/v1/attachments/{id}`

添付ファイルのアップロードは、`entity_type`、`entity_id`、`file` を持つ multipart form data を受け取ります。最初の実装では PNG、JPEG、GIF、WebP の画像ファイルを 5 MiB までサポートします。添付ファイルのバイト列は `$TQ_HOME/system/data/attachments` 配下に保存し、SQLite にはメタデータと相対パスを保存します。課題とコメントの本文は、`![screenshot](attachment://att_...)` のような Markdown image link で添付ファイルを参照します。

課題は必ず 1 つのプロジェクトに属します。`POST /api/v1/issues` は `projectId` を必須とし、課題のレスポンスは `projectId` と `projectKey` の両方を返します。`GET /api/v1/issues` は任意の query parameter として `states` と `project_id` を受け取ります。`project_id` を省略した場合は、すべてのプロジェクトの課題を一覧表示します。

JSON の成功レスポンスは `{ "data": ..., "meta": {} }` を使います。JSON のエラーレスポンスは `{ "error": { "code": "...", "message": "..." }, "meta": {} }` を使います。

`tq` CLI は課題 CRUD エンドポイントを次のコマンドでラップします。

- `tq issue list [--project <project-key>]`
- `tq issue get <id>`
- `tq issue create --project <project-key> --title <title> [--description ...] [--status ...] [--priority ...] [--assignee ...]`
- `tq issue update <id> [--title ...] [--description ...] [--status ...] [--priority ...] [--assignee ...]`
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

orchestrator は、`--port` または `server.port` で有効化したときに、実行時調査用の任意の loopback HTTP API を公開します。課題の実行時詳細レスポンスには過去の実行サマリーが含まれます。各実行は、Codex app-server thread が永続化された後に `thread_id` を含む場合があります。
