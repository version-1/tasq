---
id: api
title: API
sidebar_position: 2
---

# API

issue-tracker は user-facing Tasq API です。project、issue、comment、attachment、workflow、summary data を所有します。

## Response Envelope

success responses は次を使います。

```json
{ "data": {}, "meta": {} }
```

error responses は次を使います。

```json
{ "error": { "code": "invalid_request", "message": "..." }, "meta": {} }
```

## Issue-Tracker Endpoints

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
| `GET` | `/api/v1/issues` | issues を list します。 |
| `POST` | `/api/v1/issues` | issue を作成します。 |
| `POST` | `/api/v1/issues/states` | issue states を bulk で読みます。 |
| `GET` | `/api/v1/queue` | エージェントが実行できる issue を、依存関係から算出した queue status と一緒に list します。 |
| `GET` | `/api/v1/issues/{id}` | issue を読みます。 |
| `PATCH` | `/api/v1/issues/{id}` | issue を更新します。 |
| `PUT` | `/api/v1/issues/{issueId}/artifacts/{type}` | Artifact を作成または更新します。 |
| `DELETE` | `/api/v1/issues/{issueId}/artifacts/{type}` | Artifact を削除します。 |
| `GET` | `/api/v1/issues/{issueId}/comments` | comments を list します。 |
| `POST` | `/api/v1/issues/{issueId}/comments` | comment を追加します。 |
| `PATCH` | `/api/v1/comments/{id}` | comment を更新します。 |
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

課題レスポンスには、Artifact がない場合も `[]` となる `artifacts` 配列が常に含まれ、`type` 昇順で返ります。初期の `pull_request` Artifact では、`PUT` に `data_value` だけを指定します。サーバーは `type`、`data_type`、`data_value` を返します。`DELETE` は本文のない `204` を返します。不正な type または URL は `400`、課題または Artifact が存在しない場合は `404` です。

## Attachments

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
