---
id: web
title: Web
sidebar_position: 4
---

# Web

Web UI は、browser で Tasq issues を表示・更新するための Go-served Vite and React
application です。

hosted multi-tenant product ではなく、local operations surface として設計されています。
server は service startup 中に local loopback Web port で動作し、API calls を local
backends に proxy します。

## 責務

- issue-tracker API を通じて、ブラウザ上で Tasq issue summaries を render・更新する。
- 課題に該当 Artifact がある場合、課題カードと詳細サイドバーにプルリクエストへのリンクを表示する。
- blocked になったセッションを resume するときに Codex thread ID などの run context を
  確認できるよう、issue activity を表示する。

SPA routing と tracker/orchestrator proxy path を含む責務の全リストは [Architecture: web-ui](https://github.com/version-1/tasq/blob/main/docs/design/architecture.ja.md#web-ui) を参照してください。

## リクエスト経路

```mermaid
sequenceDiagram
  participant Browser
  participant Web as Web server
  participant Tracker as Issue Tracker
  participant Orchestrator

  Browser->>Web: Open local Web UI
  Browser->>Web: Request /tracker/api/v1/summary
  Web->>Tracker: Proxy summary request
  Tracker-->>Web: Summary JSON
  Web-->>Browser: Board data
  Browser->>Web: Request /orchestrator/...
  Web->>Orchestrator: Proxy runtime inspection
```

## 概念上の境界

Web UI は persistence を所有しません。issue-tracker API を通じて issue state を提示・
変更し、該当 views が利用可能な場合は Web server proxy を通じて orchestrator
runtime state を inspect できます。Activity と run links は navigation aids であり、
issue-tracker と orchestrator が引き続き authoritative stores です。

Artifact リンクは表示専用です。安全な新しいタブで開き、プルリクエスト Artifact がない課題ではリンク自体を表示しません。
