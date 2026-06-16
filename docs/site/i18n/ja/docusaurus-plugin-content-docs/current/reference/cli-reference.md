---
id: cli-reference
title: CLI リファレンス
sidebar_position: 1
---

# CLI リファレンス

`tq` は issue management、project setup、workflow configuration、local service、log、migration、Web UI のための command-line interface です。

## 全体の形式

```sh
tq [--api-url URL] [--output text|json] <resource> <action> [flags]
```

| Flag | Description |
| --- | --- |
| `--api-url URL` | Issue-tracker API URL。environment と state discovery を上書きします。 |
| `--output text\|json` | 出力形式。default は `text` です。 |

API URL の解決順序は `--api-url`、`TQ_API_URL`、`$TQ_HOME/system/state.json`、`http://localhost:37651` です。

## Issue command

| Command | Purpose |
| --- | --- |
| `tq issue list [--project <key>]` | issue を一覧表示します。任意で 1 つの project に scope できます。 |
| `tq issue get <id>` | 1 件の issue を表示します。 |
| `tq issue create --project <key> --title <title>` | issue を作成します。 |
| `tq issue update <id> [flags]` | issue field を更新します。 |
| `tq issue close <id>` | issue を `done` に移動します。 |
| `tq issue cancel <id>` | issue を `failed` に移動します。 |
| `tq issue ready <id>` | issue を `ready` に移動します。 |
| `tq issue draft <id>` | issue を `backlog` に移動します。 |
| `tq issue rename <id> <title>` | title を更新します。 |
| `tq issue edit <id> <description>` | description を更新します。 |

create と update は、適用できる場所で `--title`、`--description`、`--status`、`--priority`、`--assignee`、`--attach` を受け付けます。

## Comment command

| Command | Purpose |
| --- | --- |
| `tq comment add <issue-id> --body <body>` | comment を追加します。 |
| `tq comment list <issue-id>` | issue の comment を一覧表示します。 |

許可される comment type は `progress`、`blocker`、`handoff`、`general` です。

## Project and workflow command

| Command | Purpose |
| --- | --- |
| `tq project add [path] [--key <key>]` | repository を登録します。 |
| `tq project remove <key>` | project を削除します。 |
| `tq project check [key]` | project setup を validate します。 |
| `tq project list` | 登録済み project を一覧表示します。 |
| `tq workflow add --project <key> (--file <path> \| --body <text>)` | workflow override を保存します。 |
| `tq workflow remove --project <key>` | 保存された override を削除します。 |
| `tq workflow show --project <key> [--json]` | 解決された workflow を表示します。 |

## Runtime command

| Command | Purpose |
| --- | --- |
| `tq service start` | issue-tracker、orchestrator、Web UI を起動します。 |
| `tq service stop` | local service を停止します。 |
| `tq service status` | service status を表示します。 |
| `tq logs <service> [-n <lines>] [-f]` | service log を読みます。 |
| `tq migrate` | migration を適用します。 |
| `tq migrate down` | migration を rollback します。 |
| `tq migrate status` | migration status を表示します。 |
| `tq web` | 実行中の Web UI を開きます。 |
| `tq version` | version information を表示します。 |
