---
id: cli-reference
title: CLI Reference
sidebar_position: 1
---

# CLI Reference

`tq` は issue management、project setup、workflow configuration、local services、logs、migrations、Web UI のための command-line interface です。

## 全体の形式

```sh
tq [--api-url URL] [--output text|json] <resource> <action> [flags]
```

| Flag | Description |
| --- | --- |
| `--api-url URL` | Issue-tracker API URL。environment と state discovery を override します。 |
| `--output text\|json` | Output format。default は `text` です。 |

API URL resolution order は `--api-url`、`TQ_API_URL`、`$TQ_HOME/system/state.json`、最後に `http://localhost:37651` です。

## Issue Commands

| Command | Purpose |
| --- | --- |
| `tq issue list [--project <key>]` | issues を list します。任意で 1 project に scope できます。 |
| `tq issue get <id>` | 1 つの issue を表示します。 |
| `tq issue create --project <key> --title <title>` | issue を作成します。 |
| `tq issue update <id> [flags]` | issue fields を更新します。 |
| `tq issue close <id>` | issue を `done` に移動します。 |
| `tq issue cancel <id>` | issue を `failed` に移動します。 |
| `tq issue ready <id>` | issue を `ready` に移動します。 |
| `tq issue draft <id>` | issue を `backlog` に移動します。 |
| `tq issue rename <id> <title>` | title を更新します。 |
| `tq issue edit <id> <description>` | description を更新します。 |

create と update は、該当する場合に `--title`、`--description`、`--status`、`--priority`、`--assignee`、`--attach` を受け付けます。update では、依存関係を置き換える `--dependency <comma-separated-ids>` と、依存関係を削除する `--clear-dependencies` も指定できます。

## Comment Commands

| Command | Purpose |
| --- | --- |
| `tq comment add <issue-id> --body <body>` | comment を追加します。 |
| `tq comment list <issue-id>` | issue の comments を list します。 |

許可される comment types は `progress`、`blocker`、`handoff`、`general` です。

## Project and Workflow Commands

| Command | Purpose |
| --- | --- |
| `tq project add [path] [--key <key>]` | repository を登録します。 |
| `tq project remove <key>` | project を削除します。 |
| `tq project check [key]` | project setup を validate します。 |
| `tq project list` | registered projects を list します。 |
| `tq workflow add --project <key> (--file <path> \| --body <text>)` | workflow override を保存します。 |
| `tq workflow remove --project <key>` | stored override を削除します。 |
| `tq workflow show --project <key> [--json]` | resolved workflow を表示します。 |

## Runtime Commands

| Command | Purpose |
| --- | --- |
| `tq service start` | issue-tracker、orchestrator、Web UI を起動します。 |
| `tq service stop` | local services を停止します。 |
| `tq service status` | service status を表示します。 |
| `tq logs <service> [-n <lines>] [-f]` | service logs を読みます。 |
| `tq migrate` | migrations を適用します。 |
| `tq migrate down` | migrations を rollback します。 |
| `tq migrate status` | migration status を表示します。 |
| `tq web` | 実行中の Web UI を開きます。 |
| `tq version` | version information を出力します。 |
