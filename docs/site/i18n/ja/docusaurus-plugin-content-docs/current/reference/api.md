---
id: api
title: API
sidebar_position: 2
---

# API リファレンス

Issue Tracker は利用者向けの Tasq API です。プロジェクト、課題、Artifact、コメント、change request、添付ファイル、ワークフロー、サマリーのデータを所有します。

## レスポンス形式

成功時のレスポンスは次の形式です。

```json
{ "data": {}, "meta": {} }
```

エラー時のレスポンスは次の形式です。

```json
{ "error": { "code": "invalid_request", "message": "..." }, "meta": {} }
```

## Issue Tracker エンドポイント

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/health` | Health check。 |
| `GET` | `/api/v1/summary` | Board と project summary。 |
| `GET` | `/api/v1/projects` | projects を list します。 |
| `POST` | `/api/v1/projects` | project を作成します。 |
| `GET` | `/api/v1/projects/{id}` | project を読みます。 |
| `PATCH` | `/api/v1/projects/{id}` | project を更新します。 |
| `DELETE` | `/api/v1/projects/{id}` | target run が running でない場合に、project、issue-tracker descendants、orchestrator runtime descendants を削除します。 |
| `GET` | `/api/v1/projects/{id}/workflow` | stored workflow override を読みます。 |
| `PUT` | `/api/v1/projects/{id}/workflow` | workflow override を保存します。 |
| `DELETE` | `/api/v1/projects/{id}/workflow` | workflow override を削除します。 |
| `POST` | `/api/v1/projects/{id}/check` | project setup を validate します。 |
| `GET` | `/api/v1/projects/{id}/artifacts/documents` | プロジェクト所有の文書 artifact メタデータを一覧表示します。 |
| `POST` | `/api/v1/projects/{id}/artifacts/documents` | プロジェクト所有の文書 artifact を作成します。 |
| `GET` | `/api/v1/projects/{id}/artifacts/documents/{documentId}` | プロジェクト所有の文書 artifact を 1 件取得します。 |
| `PATCH` | `/api/v1/projects/{id}/artifacts/documents/{documentId}` | プロジェクト所有の文書 artifact を更新します。 |
| `DELETE` | `/api/v1/projects/{id}/artifacts/documents/{documentId}` | 関連のないプロジェクト所有の文書 artifact を削除します。 |
| `POST` | `/api/v1/projects/{id}/artifacts/documents/{documentId}/links/issues/{issueId}` | 文書 artifact を課題に関連付けます。 |
| `DELETE` | `/api/v1/projects/{id}/artifacts/documents/{documentId}/links/issues/{issueId}` | 課題との関連を解除します。 |
| `GET` | `/api/v1/issues` | issues を list します。 |
| `POST` | `/api/v1/issues` | issue を作成します。 |
| `POST` | `/api/v1/issues/states` | issue states を bulk で読みます。 |
| `GET` | `/api/v1/queue` | エージェントが実行できる issue を、依存関係から算出した queue status と一緒に list します。 |
| `GET` | `/api/v1/issues/{id}` | issue を読みます。 |
| `PATCH` | `/api/v1/issues/{id}` | issue を更新します。 |
| `PUT` | `/api/v1/issues/{issueId}/artifacts/{type}` | Artifact を作成または更新します。 |
| `DELETE` | `/api/v1/issues/{issueId}/artifacts/{type}` | Artifact を削除します。 |
| `GET` | `/api/v1/issues/{issueId}/artifacts/documents` | Document Artifact のメタデータを一覧表示します。 |
| `POST` | `/api/v1/issues/{issueId}/artifacts/documents` | Document Artifact を作成します。 |
| `GET` | `/api/v1/issues/{issueId}/artifacts/documents/{documentId}` | Document Artifact を 1 件取得します。 |
| `PATCH` | `/api/v1/issues/{issueId}/artifacts/documents/{documentId}` | Document Artifact を更新します。 |
| `GET` | `/api/v1/issues/{issueId}/comments` | comments を list します。 |
| `POST` | `/api/v1/issues/{issueId}/comments` | comment を追加します。 |
| `PATCH` | `/api/v1/comments/{id}` | comment を更新します。 |
| `GET` | `/api/v1/issues/{issueId}/change-requests` | 課題の change request を一覧表示します。 |
| `POST` | `/api/v1/issues/{issueId}/change-requests` | `open` 状態の change request を作成します。 |
| `GET` | `/api/v1/change-requests/{id}` | change request を取得します。 |
| `PATCH` | `/api/v1/change-requests/{id}` | `open` の間に本文を編集するか、許可された状態遷移を行います。 |
| `POST` | `/api/v1/change-requests/{id}/cancel` | `open` または `in_progress` の change request を取り消します。 |
| `GET` | `/api/v1/attachments` | attachments を list します。 |
| `POST` | `/api/v1/attachments` | attachment を upload します。 |
| `GET` | `/api/v1/attachments/{id}/content` | attachment bytes を download します。 |
| `DELETE` | `/api/v1/attachments/{id}` | attachment を削除します。 |

issue listing では `states`、`project_id`、`project_ids`、`priorities`、
`assignee`、`search` で絞り込めます。sorting では `sort_by` に `id`、
`priority`、`created_at`、`updated_at` を指定でき、`sort_direction` は `asc` または
`desc` です。pagination は `limit` と `offset` を使い、`limit` の上限は `50` です。

comment listing は `cursor` と `limit` を受け付けます。attachment listing は
`entity_type` と `entity_id` を受け付けます。

change request の一覧取得では、任意の `status` フィルターと `1`〜`100` の `limit` を指定できます。`limit` の既定値は `50` です。

課題レスポンスには、Artifact がない場合も `[]` となる `artifacts` 配列が常に含まれ、`type` 昇順で返ります。初期の `pull_request` Artifact では、`PUT` に `data_value` だけを指定します。サーバーは `type`、`data_type`、`data_value` を返します。`DELETE` は本文のない `204` を返します。不正な type または URL は `400`、課題または Artifact が存在しない場合は `404` です。

文書 artifact は URL artifact とは別に Markdown を保存し、プロジェクトが所有します。文書にはサーバー生成の ID、`type`（`reference_document` または `work_report`）、タイトル、本文があります。課題との関連はなくてもよく、別プロジェクトの課題を含む複数の課題に関連できます。すべての関連は同じ本文を共有します。作成、取得、更新、削除、関連、解除にはプロジェクト route を正規 API として使用します。一覧には本文を含めず、個別取得は保存済みの Markdown をそのまま返します。`PATCH` は省略したフィールドを維持します。関連がある文書は `409` となり削除できません。課題の削除では、その関連だけが削除されます。課題の文書 route は互換性のために残ります。作成時には初期関連も作成され、取得にはその関連が必要です。所有者が別プロジェクトである文書を課題 route から更新すると拒否されます。

現在のサービスには認証・認可 middleware がありません。プロジェクト所有は将来のポリシー層のデータ境界であり、別プロジェクトの課題との関連はアクセス権を与えません。

change request は、後続のエージェント実行で対応する追加作業を記録します。作成時の状態は `open` です。許可される状態遷移は `open` から `in_progress` または `canceled`、`in_progress` から `resolved` または `canceled` です。`resolved` と `canceled` は変更できません。取り消しは状態遷移として扱われ、物理削除のエンドポイントはありません。

## コントラクトの参照先

[Issue Tracker の OpenAPI 文書](https://github.com/version-1/tasq/blob/main/docs/openapi/issue-tracker.yml)と [orchestrator の OpenAPI 文書](https://github.com/version-1/tasq/blob/main/docs/openapi/orchestrator.yml)が、リクエストパラメーター、本文、レスポンス、エラーステータスを定義します。このページはエンドポイントと動作の要約です。

## 添付ファイル

Attachment uploads は `entity_type`、`entity_id`、`file` を含む multipart form data を使います。supported image types は PNG、JPEG、GIF、WebP で、上限は 5 MiB です。

Attachment bytes は `$TQ_HOME/system/data/attachments` 配下に置かれます。SQLite は metadata と relative paths を保存するため、rows を書き換えずに `TQ_HOME` を移動できます。

## Orchestrator API

orchestrator は runtime inspection 向けの loopback HTTP APIs を公開します。user-facing
issue mutation API ではありません。

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/state` | running / retrying runs と集約された runtime metadata を確認します。 |
| `POST` | `/api/v1/refresh` | refresher が設定されている場合に orchestrator refresh を要求します。 |
| `GET` | `/api/v1/{issue_identifier}` | 1 つの issue の runtime state、runs、workspace path、recent events を確認します。 |
| `GET` | `/api/v1/{issue_identifier}/runs/{run_id}/conversations` | run の conversation events を読みます。 |

`issue_identifier` には `issue-12` のような orchestrator issue identifier form を指定します。
