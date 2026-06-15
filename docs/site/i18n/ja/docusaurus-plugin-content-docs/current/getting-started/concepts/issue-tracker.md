---
id: issue-tracker
title: Issue Tracker
sidebar_position: 3
---

# Issue Tracker

issue-tracker は Tasq work の user-facing source of truth です。

projects、issues、comments、attachment metadata、summary data を SQLite に保存します。また、`tq`、Web UI、その他の clients が使う API を提供します。

## 責務

- issue data を保存して validate する。
- 各 issue が 1 つの project に属することを要求する。
- comments と image attachment metadata を保存する。
- attachment bytes を `$TQ_HOME/system/data/attachments` 配下に保存する。
- project、issue、comment、attachment、workflow、summary endpoints を提供する。
- linked issues が存在する間は project deletion を防ぐ。

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

## データ構造

各 issue は project、title、description、status、priority、optional assignee、timestamps、comments、optional image attachments を持ちます。Markdown descriptions と comments は `attachment://<id>` URLs で uploaded images を参照できます。

issue-tracker API は JSON success responses を `{ "data": ..., "meta": {} }` として、error responses を `{ "error": { "code": "...", "message": "..." }, "meta": {} }` として返します。
