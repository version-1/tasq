---
id: api
title: API
sidebar_position: 2
---

# API

The issue-tracker is the user-facing Tasq API. It owns project, issue, artifact, comment, change-request, attachment, workflow, and summary data.

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
| `PUT` | `/api/v1/issues/{issueId}/artifacts/{type}` | Create or update an artifact. |
| `DELETE` | `/api/v1/issues/{issueId}/artifacts/{type}` | Delete an artifact. |
| `GET` | `/api/v1/issues/{issueId}/comments` | List comments. |
| `POST` | `/api/v1/issues/{issueId}/comments` | Add a comment. |
| `PATCH` | `/api/v1/comments/{id}` | Update a comment. |
| `GET` | `/api/v1/issues/{issueId}/change-requests` | List change requests for an issue. |
| `POST` | `/api/v1/issues/{issueId}/change-requests` | Create an open change request. |
| `GET` | `/api/v1/change-requests/{id}` | Read a change request. |
| `PATCH` | `/api/v1/change-requests/{id}` | Edit its body or perform an allowed status transition. |
| `POST` | `/api/v1/change-requests/{id}/cancel` | Cancel an open or in-progress change request. |
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

Issue responses always include an `artifacts` array, sorted by `type`, including `[]` when no artifacts exist. The initial `pull_request` artifact accepts only `data_value` on `PUT`; the server returns `type`, `data_type`, and `data_value`. `DELETE` returns an empty `204`. Invalid types or URLs return `400`; a missing issue or artifact returns `404`.

Change requests capture additional work for a later agent run. Creation sets the status to `open`. Allowed transitions are `open` to `in_progress` or `canceled`, and `in_progress` to `resolved` or `canceled`. Resolved and canceled requests are immutable. Cancellation is a state transition; there is no physical delete endpoint.

## Contract Sources

The [issue-tracker OpenAPI document](https://github.com/version-1/tasq/blob/main/docs/openapi/issue-tracker.yml) and [orchestrator OpenAPI document](https://github.com/version-1/tasq/blob/main/docs/openapi/orchestrator.yml) define request parameters, bodies, responses, and error status codes. This page is the concise endpoint and behavior reference.

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
