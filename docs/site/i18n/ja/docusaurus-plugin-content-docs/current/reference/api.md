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
| `DELETE` | `/api/v1/projects/{id}` | linked issues がない場合に project を削除します。 |
| `GET` | `/api/v1/projects/{id}/workflow` | stored workflow override を読みます。 |
| `PUT` | `/api/v1/projects/{id}/workflow` | workflow override を保存します。 |
| `DELETE` | `/api/v1/projects/{id}/workflow` | workflow override を削除します。 |
| `POST` | `/api/v1/projects/{id}/check` | project setup を validate します。 |
| `GET` | `/api/v1/issues` | issues を list します。 |
| `POST` | `/api/v1/issues` | issue を作成します。 |
| `POST` | `/api/v1/issues/states` | issue states を bulk で読みます。 |
| `GET` | `/api/v1/issues/{id}` | issue を読みます。 |
| `PATCH` | `/api/v1/issues/{id}` | issue を更新します。 |
| `GET` | `/api/v1/issues/{issueId}/comments` | comments を list します。 |
| `POST` | `/api/v1/issues/{issueId}/comments` | comment を追加します。 |
| `PATCH` | `/api/v1/comments/{id}` | comment を更新します。 |
| `GET` | `/api/v1/attachments` | attachments を list します。 |
| `POST` | `/api/v1/attachments` | attachment を upload します。 |
| `GET` | `/api/v1/attachments/{id}/content` | attachment bytes を download します。 |
| `DELETE` | `/api/v1/attachments/{id}` | attachment を削除します。 |

## Attachments

Attachment uploads は `entity_type`、`entity_id`、`file` を含む multipart form data を使います。supported image types は PNG、JPEG、GIF、WebP で、上限は 5 MiB です。

Attachment bytes は `$TQ_HOME/system/data/attachments` 配下に置かれます。SQLite は metadata と relative paths を保存するため、rows を書き換えずに `TQ_HOME` を移動できます。

## Orchestrator API

orchestrator は、port configuration で enabled になっている場合に runtime inspection 向けの optional loopback HTTP APIs を公開します。user-facing issue mutation API ではありません。
