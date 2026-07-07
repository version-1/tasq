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
| `DELETE` | `/api/v1/projects/{id}` | Delete a project, issue-tracker descendants, and orchestrator runtime descendants unless a target run is running. |
| `GET` | `/api/v1/projects/{id}/workflow` | Read stored workflow override. |
| `PUT` | `/api/v1/projects/{id}/workflow` | Store workflow override. |
| `DELETE` | `/api/v1/projects/{id}/workflow` | Remove workflow override. |
| `POST` | `/api/v1/projects/{id}/check` | Validate project setup. |
| `GET` | `/api/v1/issues` | List issues. |
| `POST` | `/api/v1/issues` | Create an issue. |
| `POST` | `/api/v1/issues/states` | Read issue states in bulk. |
| `GET` | `/api/v1/queue` | List issues that are eligible for agent execution, including dependency-derived queue status. |
| `GET` | `/api/v1/issues/{id}` | Read an issue. |
| `PATCH` | `/api/v1/issues/{id}` | Update an issue. |
| `GET` | `/api/v1/issues/{issueId}/comments` | List comments. |
| `POST` | `/api/v1/issues/{issueId}/comments` | Add a comment. |
| `PATCH` | `/api/v1/comments/{id}` | Update a comment. |
| `GET` | `/api/v1/attachments` | List attachments. |
| `POST` | `/api/v1/attachments` | Upload an attachment. |
| `GET` | `/api/v1/attachments/{id}/content` | Download attachment bytes. |
| `DELETE` | `/api/v1/attachments/{id}` | Delete an attachment. |

Issue listing supports filters for `states`, `project_id`, `project_ids`,
`priorities`, `assignee`, and `search`. Sorting supports `sort_by` values
`id`, `priority`, `created_at`, and `updated_at` with `sort_direction` `asc` or
`desc`. Pagination uses `limit` and `offset`; `limit` is capped at `50`.

Comment listing supports `cursor` and `limit`. Attachment listing supports
`entity_type` and `entity_id`.

## Attachments

Attachment uploads use multipart form data with `entity_type`, `entity_id`, and `file`. Supported image types are PNG, JPEG, GIF, and WebP up to 5 MiB.

Attachment bytes live under `$TQ_HOME/system/data/attachments`. SQLite stores metadata and relative paths so `TQ_HOME` can move without rewriting rows.

## Orchestrator API

The orchestrator exposes loopback HTTP APIs for runtime inspection. It is not
the user-facing issue mutation API.

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/state` | Inspect running and retrying runs plus aggregated runtime metadata. |
| `POST` | `/api/v1/refresh` | Request an orchestrator refresh when a refresher is configured. |
| `GET` | `/api/v1/{issue_identifier}` | Inspect runtime state, runs, workspace path, and recent events for one issue. |
| `GET` | `/api/v1/{issue_identifier}/runs/{run_id}/conversations` | Read conversation events for a run. |

`issue_identifier` accepts the orchestrator issue identifier form such as
`issue-12`.
