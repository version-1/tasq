---
id: schema
title: Schema
sidebar_position: 4
---

# Schema

Tasq validates entity data at the store layer on create and update operations. This reference summarizes the public field constraints that clients should respect.

## Issue Tracker Entities

| Entity | Required fields | Key constraints |
| --- | --- | --- |
| Issue | `projectId`, `title` | title 1-500 chars, description max 10,000 chars, assignee max 200 chars, immutable project ownership |
| Artifact | `type`, `dataType`, `dataValue` | one per issue and type; initial type `pull_request` has data type `url`; URL is trimmed, absolute HTTP(S), host-required, userinfo-free, max 4,096 UTF-8 bytes |
| Comment | `issueId`, `author`, `body` | body 1-10,000 chars, type defaults to `general` |
| ChangeRequest | `issueId`, `author`, `body` | body 1-10,000 chars; status defaults to `open`; terminal requests are immutable |
| Attachment | `entityType`, `entityId`, `file` | image PNG/JPEG/GIF/WebP, max 5 MiB |
| Project | `key`, `name`, `location` | key format, name 1-200 chars, description max 10,000 chars, absolute location |
| ProjectWorkflow | `projectId`, `frontmatter`, `body`, `checksum` | one workflow override per project, checksum is SHA256 hex |

## Orchestrator Entities

| Entity | Required fields | Key constraints |
| --- | --- | --- |
| Run | `issueId`, `attempt`, `orchestratorId` | run ID generated, status defaults to `queued` |
| RunnerEvent | `runId`, `occurredAt` | payload JSON must be valid when present |
| WorkspaceMetadata | `workspaceKey`, `issueId`, `path` | workspace key max 200 chars, paths must be absolute |
| WorkspaceSetupFailure | `issueId`, `error` | records setup failure context |

## Enums

| Field | Values |
| --- | --- |
| Issue status | `backlog`, `ready`, `in_progress`, `review`, `done`, `blocked`, `failed`, `cancelled`, `duplicate` |
| Queue status | `backlog`, `pending`, `queued`, `processing`, `completed`, `inactive` |
| Issue priority | `low`, `normal`, `high`, `urgent` |
| Comment type | `progress`, `blocker`, `handoff`, `general` |
| Change-request status | `open`, `in_progress`, `resolved`, `canceled` |
| Attachment entity type | `issue`, `comment` |
| Artifact type | `pull_request` |
| Artifact data type | `url` |
| Run status | `queued`, `running`, `succeeded`, `failed`, `cancelled` |

## String and Path Limits

```mermaid
flowchart TD
  Short[200 chars] --> Assignee[assignee, project name, event type]
  Medium[500 chars] --> Title[issue title]
  Path[1,000 chars] --> Paths[project, workspace, attachment paths]
  Long[10,000 chars] --> Bodies[descriptions, comments, errors]
  Payload[50,000 chars] --> JSON[runner event payload JSON]
```

Absolute path fields must start with `/`. Clients such as `tq project add` check local directory existence when they can access the target filesystem.

## State Transitions

Change requests allow `open` to `in_progress` or `canceled`, and `in_progress` to `resolved` or `canceled`. Resolving can record the orchestrator run and a result comment on the same issue. `resolved` and `canceled` are terminal states.
