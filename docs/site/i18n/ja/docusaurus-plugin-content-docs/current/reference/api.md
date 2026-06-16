---
id: api
title: API
sidebar_position: 2
---

# API

issue-tracker は Tasq の user-facing API です。project、issue、comment、attachment、workflow、summary data を所有します。

## Response envelope

Success response は次の形式です。

```json
{ "data": {}, "meta": {} }
```

Error response は次の形式です。

```json
{ "error": { "code": "invalid_request", "message": "..." }, "meta": {} }
```

## Issue-tracker endpoint

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/health` | health check。 |
| `GET` | `/api/v1/summary` | board と project summary。 |
| `GET` | `/api/v1/projects` | project を一覧表示します。 |
| `POST` | `/api/v1/projects` | project を作成します。 |
| `GET` | `/api/v1/projects/{id}` | project を読み取ります。 |
| `PATCH` | `/api/v1/projects/{id}` | project を更新します。 |
| `DELETE` | `/api/v1/projects/{id}` | linked issue がない場合に project を削除します。 |
| `GET` | `/api/v1/projects/{id}/workflow` | 保存された workflow override を読み取ります。 |
| `PUT` | `/api/v1/projects/{id}/workflow` | workflow override を保存します。 |
| `DELETE` | `/api/v1/projects/{id}/workflow` | workflow override を削除します。 |
| `POST` | `/api/v1/projects/{id}/check` | project setup を validate します。 |
| `GET` | `/api/v1/issues` | issue を一覧表示します。 |
| `POST` | `/api/v1/issues` | issue を作成します。 |
| `POST` | `/api/v1/issues/states` | issue state を bulk read します。 |
| `GET` | `/api/v1/issues/{id}` | issue を読み取ります。 |
| `PATCH` | `/api/v1/issues/{id}` | issue を更新します。 |
| `GET` | `/api/v1/issues/{issueId}/comments` | comment を一覧表示します。 |
| `POST` | `/api/v1/issues/{issueId}/comments` | comment を追加します。 |
| `PATCH` | `/api/v1/comments/{id}` | comment を更新します。 |
| `GET` | `/api/v1/attachments` | attachment を一覧表示します。 |
| `POST` | `/api/v1/attachments` | attachment を upload します。 |
| `GET` | `/api/v1/attachments/{id}/content` | attachment bytes を download します。 |
| `DELETE` | `/api/v1/attachments/{id}` | attachment を削除します。 |

## Attachment

Attachment upload は `entity_type`、`entity_id`、`file` を含む multipart form data を使います。対応する image type は PNG、JPEG、GIF、WebP で、上限は 5 MiB です。

Attachment bytes は `$TQ_HOME/system/data/attachments` 配下に置かれます。SQLite は metadata と relative path を保存するため、`TQ_HOME` が移動しても row を書き換える必要がありません。

## Orchestrator API

orchestrator は port configuration で有効化された場合に、runtime inspection のための optional loopback HTTP API を公開します。user-facing issue mutation API ではありません。
