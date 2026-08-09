---
id: issue-tracker
title: Issue Tracker
sidebar_position: 3
---

# Issue Tracker

The issue-tracker is the user-facing source of truth for Tasq work.

It stores projects, issues, issue artifacts, comments, attachment metadata, and summary data in
SQLite. It also serves the API used by `tq`, the Web UI, agents, and other
clients.

## Responsibilities

- Store and validate issue data.
- Require each issue to belong to one project.
- Store comments and image attachment metadata.
- Associate issues with external artifacts such as pull-request URLs.
- Store attachment bytes under `$TQ_HOME/system/data/attachments`.
- Serve project, issue, comment, attachment, workflow, and summary endpoints.
- Prevent project deletion while linked issues exist.

## Issue Workflow

```mermaid
stateDiagram-v2
  [*] --> backlog
  backlog --> ready
  ready --> in_progress
  in_progress --> review
  review --> done
  in_progress --> blocked
  blocked --> in_progress
  in_progress --> failed
  failed --> backlog
```

## Data Shape

Each issue has a project, title, description, status, priority, optional assignee, timestamps, comments, optional image attachments, and an `artifacts` array. The initial artifact is one `pull_request` URL per issue. Markdown descriptions and comments can reference uploaded images with `attachment://<id>` URLs.

The issue-tracker API returns JSON success responses as `{ "data": ..., "meta": {} }` and error responses as `{ "error": { "code": "...", "message": "..." }, "meta": {} }`.
