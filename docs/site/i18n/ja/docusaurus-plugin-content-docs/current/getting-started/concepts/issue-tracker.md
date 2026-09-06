---
id: issue-tracker
title: Issue Tracker
sidebar_position: 3
---

# Issue Tracker

issue-tracker は Tasq work の user-facing source of truth です。

プロジェクト、課題、課題 Artifact、コメント、添付ファイルのメタデータ、サマリーデータを SQLite に保存します。
また、`tq`、Web UI、agents、その他の clients が使う API を提供します。

## 責務

- project、issue、comment、attachment データを保存・validate し、Tasq work の唯一の source of truth となる。
- `tq`、Web UI、agents が使う project、issue、comment、attachment、workflow、summary API を提供する。
- プルリクエスト URL など、課題に関連付ける外部 Artifact を保存する。

責務の全リストと所有境界は [Architecture: issue-tracker](https://github.com/version-1/tasq/blob/main/docs/design/architecture.ja.md#issue-tracker) を参照してください。

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

各課題はプロジェクト、タイトル、説明、ステータス、優先度、任意の担当者、時刻、コメント、任意の画像添付、`artifacts` 配列を持ちます。初期の Artifact は、各課題に 1 件だけ関連付けられる `pull_request` URL です。Markdown の説明とコメントでは、`attachment://<id>` URL を使ってアップロード済み画像を参照できます。

issue-tracker API は JSON success responses を `{ "data": ..., "meta": {} }` として、
error responses を `{ "error": { "code": "...", "message": "..." }, "meta": {} }` として
返します。
