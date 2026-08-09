# Tasq API

この文書では、Issue Tracker API の動作と所有範囲を説明します。パス、HTTP メソッド、パラメーター、スキーマの正式な仕様は [Issue Tracker の OpenAPI 文書](../openapi/issue-tracker.yml)です。コンポーネントの所有範囲については [architecture.ja.md](architecture.ja.md)、ローカルでの運用と検証については [operations.ja.md](operations.ja.md)を参照してください。

## 契約

Issue Tracker は、Tasq が利用者向けに公開する API です。JSON の成功レスポンスは `{ "data": ..., "meta": {} }`、エラーレスポンスは `{ "error": { "code": "...", "message": "..." }, "meta": {} }` の形式です。

API を変更するときは、[development.ja.md](../development.ja.md)の手順に従って OpenAPI 文書と生成済みクライアントを更新します。この文書では、個々のスキーマ定義から切り離して説明した方が理解しやすい、エンドポイント横断の動作を扱います。

## プロジェクト

すべての課題は、必ず1つのプロジェクトに所属します。プロジェクトのレスポンスには数値 ID とキーの両方が含まれますが、コマンドでは通常、プロジェクトキーを指定します。

`DELETE /api/v1/projects/{id}` は、プロジェクトと、そのプロジェクトが所有する次の Issue Tracker データを削除します。

- 課題と、その課題を参照する依存関係
- コメントと変更依頼
- 添付ファイルのレコードと `$TQ_HOME/system/data/attachments` 以下のファイル
- 保存済みのプロジェクト別ワークフロー上書き

この操作は、対象課題が所有するオーケストレーターの実行時データも削除します。対象は、実行、ランナーイベント、ワークスペースのメタデータ、ワークスペース準備の失敗記録です。途中で失敗しても課題 ID が残っている状態から再試行できるように、オーケストレーターのデータを Issue Tracker のレコードより先に削除します。

対象課題に `running` 状態のオーケストレーター実行がある場合、削除は `projects.delete.running_runs` を伴う `409 Conflict` を返し、何も変更しません。`project.location` に記録されたディレクトリや、そのワークツリーを削除または変更することはありません。

## 課題と依存関係

`POST /api/v1/issues` では `projectId` が必須です。任意の `dependency_ids` を指定すると、作成と同じ操作で初期の依存関係を設定できます。省略するか空配列を渡すと、依存関係のない課題を作成します。

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

`PATCH /api/v1/issues/{id}` で `dependency_ids` を指定した場合は、依存関係全体を置き換えます。このフィールドを省略すると既存の依存関係を維持し、空配列を渡すとすべて削除します。作成と更新のどちらでも、存在しない課題への依存、自分自身への依存、ID の重複、依存関係の循環を拒否します。

課題のレスポンスには `projectId`、`projectKey`、`dependency_ids`、`artifacts` が含まれます。依存関係と Artifact の一覧は必須配列であり、空の場合は `[]` です。Artifact は `type` 昇順で返します。

### Artifact

`PUT /api/v1/issues/{issue_id}/artifacts/{type}` は、課題と type に対応する Artifact を 1 件作成または更新します。リクエスト本文には `data_value` のみを指定し、`data_type` はサーバーが決定します。作成・更新ともに、Artifact の公開フィールドである `type`、`data_type`、`data_value` を含む `200` を返します。

`DELETE /api/v1/issues/{issue_id}/artifacts/{type}` は Artifact を削除し、本文のない `204` を返します。課題または Artifact が存在しない場合は `404`、未対応の type、不正な本文、不正な URL は `400` です。

初期対応の type は `pull_request` で、`data_type` は `url` です。URL は前後の空白を除去してから検証します。host を持つ絶対 `http` / `https` URL であること、userinfo を含まないこと、UTF-8 で 4,096 bytes 以下であることが必要です。空白除去以外に API が URL を正規化することはありません。

### 一覧と検索

`GET /api/v1/issues` では、`states`、`project_id`、`project_ids`、`priorities`、`assignee`、`search`、`limit`、`offset`、`sort_by`、`sort_direction` を使用できます。

- プロジェクトの絞り込みを省略すると、すべてのプロジェクトの課題を取得します。
- `project_id` は1つのプロジェクトを選択し、`project_ids` はカンマ区切りで複数指定できます。
- `priorities` は優先度をカンマ区切りで指定します。
- `search` は課題 ID との完全一致、またはタイトルとの大文字・小文字を区別しない部分一致です。数字だけの検索文字列は、どちらにも一致する可能性があります。
- 空または空白だけの `search` は無視します。
- 検索と絞り込みを適用してから、並べ替えとページ分割を行います。
- `sort_by` には `id`、`priority`、`created_at`、`updated_at`、`sort_direction` には `asc` または `desc` を指定できます。

## キューとサマリー

`GET /api/v1/queue` は、実行準備ができた課題を `queued` と `pending` の配列に分けます。

- `queued` は、処理を妨げる依存関係がない課題です。
- `pending` は、処理を妨げる依存関係が1つ以上ある課題で、`blocked_dependency_ids` を含みます。

依存先の状態が `done`、`cancelled`、`duplicate` の場合は解決済みとして扱い、それ以外は処理を妨げます。どちらの配列も優先度順（`urgent`、`high`、`normal`、`low`）、同じ優先度では課題 ID の昇順に並びます。課題一覧と同じ `project_id` による絞り込みを使用できます。

`GET /api/v1/summary` は、ボード表示用の情報を返します。各課題のサマリーには、キューでの位置づけを示す `queueStatus` が含まれます。

| `queueStatus` | 意味 |
|---|---|
| `backlog` | 課題の状態が `backlog`。 |
| `pending` | 課題は `ready` だが、処理を妨げる依存関係がある。 |
| `queued` | 課題は `ready` で、処理を妨げる依存関係がない。 |
| `processing` | 課題の状態が `in_progress`。 |
| `completed` | 課題の状態が `done`。 |
| `inactive` | `review`、`blocked`、`failed`、`cancelled`、`duplicate` など、キューの処理対象外。 |

状態の定義と遷移については [status.ja.md](status.ja.md)を参照してください。

## コメントと変更依頼

コメントは議論を記録します。変更依頼は、利用者やレビュアーが追加で求める作業を表し、ワークフロー上の状態を持ちます。

コメント一覧は、既存の forward 契約を既定値として維持します。`cursor` より大きい ID のコメントを時系列順で返します。`direction=backward` を指定すると、`cursor` より小さい ID のコメントを新しい順で返します。backward 方向で cursor にゼロを指定した場合は最新ページを取得します。時系列順かつ最新コメントを末尾に表示するクライアントは、backward ページ内の順序を反転して先頭へ追加します。

課題に変更依頼を作成すると、状態は `open` になります。`open` の変更依頼は、内容を編集するか `in_progress` に移行できます。`in_progress` の変更依頼は、解決済みまたは取り消しにできます。許可する遷移は次のとおりです。

- `open -> in_progress`
- `open -> canceled`
- `in_progress -> resolved`
- `in_progress -> canceled`

`resolved` と `canceled` の変更依頼は変更できません。取り消しには `POST /api/v1/change-requests/{id}/cancel` を使用し、物理削除のエンドポイントは提供しません。

オーケストレーターが過去の実行を持つ課題の作業を継続するときは、`open` の変更依頼を古い順に最大20件取得し、対象を `in_progress` に移して Codex の継続指示へ加えます。継続指示では、対応した変更依頼に `resolvedByRunId` を設定し、結果コメントがある場合は `resultCommentId` も設定して `resolved` にするようエージェントへ求めます。

## 添付ファイル

添付ファイルのアップロードには、`entity_type`、`entity_id`、`file` を含む multipart form data を使用します。PNG、JPEG、GIF、WebP 形式の画像を5 MiBまでアップロードできます。

ファイル本体は `$TQ_HOME/system/data/attachments` 以下に保存し、SQLite にはメタデータと相対パスを保存します。課題の説明とコメントでは、`![screenshot](attachment://att_...)` のような Markdown で画像を参照します。

## CLI からの利用

通常の課題、Artifact、コメント、プロジェクト、ワークフロー操作には、型付きの `tq` コマンドを使用します。必要な Issue Tracker 操作が型付きコマンドにない場合に限り、許可リストで制限された `tq api` を使用します。

生の API コマンドは、HTTP メソッドとルートの独自の許可リストを持ち、閉じた状態を既定とします。OpenAPI にルートを追加しても自動では公開しません。また、multipart リクエストの組み立てに対応していないため、`POST /api/v1/attachments` を一時的に除外しています。構文、入力検証、出力、終了ステータスについては、[tq コマンドリファレンス](../references/tq.ja.md#生の-api-リクエスト)を参照してください。

## オーケストレーター調査 API

オーケストレーターは、`--port` または `server.port` で有効にした場合に、実行時の調査に使う別のループバック HTTP API を公開できます。この API は Issue Tracker API には含まれません。課題の実行時詳細には過去の実行サマリーが含まれ、Codex app-server のスレッドを永続化した後は、各実行に `thread_id` が含まれる場合があります。
