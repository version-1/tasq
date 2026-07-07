---
id: schema
title: Schema
sidebar_position: 4
---

# Schema

Tasq は create と update operations において store layer で entity data を validate します。この reference は clients が守るべき public field constraints を要約します。

## Issue Tracker Entities

| Entity | Required fields | Key constraints |
| --- | --- | --- |
| Issue | `projectId`, `title` | title 1-500 chars、description max 10,000 chars、assignee max 200 chars、immutable project ownership |
| Comment | `issueId`, `author`, `body` | body 1-10,000 chars、type defaults to `general` |
| Attachment | `entityType`, `entityId`, `file` | image PNG/JPEG/GIF/WebP、max 5 MiB |
| Project | `key`, `name`, `location` | key format、name 1-200 chars、description max 10,000 chars、absolute location |
| ProjectWorkflow | `projectId`, `frontmatter`, `body`, `checksum` | one workflow override per project、checksum は SHA256 hex |

## Orchestrator Entities

| Entity | Required fields | Key constraints |
| --- | --- | --- |
| Run | `issueId`, `attempt`, `orchestratorId` | run ID generated、status defaults to `queued` |
| RunnerEvent | `runId`, `occurredAt` | payload JSON must be valid when present |
| WorkspaceMetadata | `workspaceKey`, `issueId`, `path` | workspace key max 200 chars、paths must be absolute |
| WorkspaceSetupFailure | `issueId`, `error` | records setup failure context |

## Enums

| Field | Values |
| --- | --- |
| Issue status | `backlog`, `ready`, `in_progress`, `review`, `done`, `blocked`, `failed`, `cancelled`, `duplicate` |
| Queue status | `backlog`, `pending`, `queued`, `processing`, `completed`, `inactive` |
| Issue priority | `low`, `normal`, `high`, `urgent` |
| Comment type | `progress`, `blocker`, `handoff`, `general` |
| Attachment entity type | `issue`, `comment` |
| Run status | `queued`, `running`, `succeeded`, `failed`, `cancelled` |

## 文字列とパスの上限

```mermaid
flowchart TD
  Short[200 chars] --> Assignee[assignee, project name, event type]
  Medium[500 chars] --> Title[issue title]
  Path[1,000 chars] --> Paths[project, workspace, attachment paths]
  Long[10,000 chars] --> Bodies[descriptions, comments, errors]
  Payload[50,000 chars] --> JSON[runner event payload JSON]
```

absolute path fields は `/` で始まる必要があります。`tq project add` のような clients は、target filesystem に access できる場合に local directory existence を確認します。
