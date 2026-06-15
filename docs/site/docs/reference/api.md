---
id: api
title: API
sidebar_position: 2
---

# API

The issue-tracker is the user-facing Tasq API. It owns project, issue, comment, attachment, workflow, and summary data.

## Response Envelope

Success responses use:

```json
{ "data": {}, "meta": {} }
```

Error responses use:

```json
{ "error": { "code": "invalid_request", "message": "..." }, "meta": {} }
```

## Issue-Tracker Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/health` | Health check. |
| `GET` | `/api/v1/summary` | Board and project summary. |
| `GET` | `/api/v1/projects` | List projects. |
| `POST` | `/api/v1/projects` | Create a project. |
| `GET` | `/api/v1/projects/{id}` | Read a project. |
| `PATCH` | `/api/v1/projects/{id}` | Update a project. |
| `DELETE` | `/api/v1/projects/{id}` | Delete a project when no issues link to it. |
| `GET` | `/api/v1/projects/{id}/workflow` | Read stored workflow override. |
| `PUT` | `/api/v1/projects/{id}/workflow` | Store workflow override. |
| `DELETE` | `/api/v1/projects/{id}/workflow` | Remove workflow override. |
| `POST` | `/api/v1/projects/{id}/check` | Validate project setup. |
| `GET` | `/api/v1/issues` | List issues. |
| `POST` | `/api/v1/issues` | Create an issue. |
| `POST` | `/api/v1/issues/states` | Read issue states in bulk. |
| `GET` | `/api/v1/issues/{id}` | Read an issue. |
| `PATCH` | `/api/v1/issues/{id}` | Update an issue. |
| `GET` | `/api/v1/issues/{issueId}/comments` | List comments. |
| `POST` | `/api/v1/issues/{issueId}/comments` | Add a comment. |
| `PATCH` | `/api/v1/comments/{id}` | Update a comment. |
| `GET` | `/api/v1/attachments` | List attachments. |
| `POST` | `/api/v1/attachments` | Upload an attachment. |
| `GET` | `/api/v1/attachments/{id}/content` | Download attachment bytes. |
| `DELETE` | `/api/v1/attachments/{id}` | Delete an attachment. |

## Attachments

Attachment uploads use multipart form data with `entity_type`, `entity_id`, and `file`. Supported image types are PNG, JPEG, GIF, and WebP up to 5 MiB.

Attachment bytes live under `$TQ_HOME/system/data/attachments`. SQLite stores metadata and relative paths so `TQ_HOME` can move without rewriting rows.

## Orchestrator API

The orchestrator exposes optional loopback HTTP APIs for runtime inspection when enabled with a port configuration. It is not the user-facing issue mutation API.
