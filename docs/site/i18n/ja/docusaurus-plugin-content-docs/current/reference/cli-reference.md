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
| `tq issue watch [--interval <duration>] [--seen-ttl <duration>] [--verbose]` | ready な issue を polling し、JSON event envelopes を出力します。 |
| `tq issue close <id>` | issue を `done` に移動します。 |
| `tq issue cancel <id>` | issue を `failed` に移動します。 |
| `tq issue ready <id>` | issue を `ready` に移動します。 |
| `tq issue draft <id>` | issue を `backlog` に移動します。 |
| `tq issue rename <id> <title>` | title を更新します。 |
| `tq issue edit <id> <description>` | description を更新します。 |

create と update は、該当する場合に `--title`、`--description`、`--status`、
`--priority`、`--assignee`、`--attach` を受け付けます。update では、依存関係を
置き換える `--dependency <comma-separated-ids>` と、依存関係を削除する
`--clear-dependencies` も指定できます。

`tq issue watch` はエージェントの loop で使うことを想定しています。ready queue を
読み、設定された seen TTL の間は同じ issue の出力を重複排除し、`issue-ready` event
を出力します。一時的な API error が起きても polling を継続します。

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
| `tq project remove [-y] <key>` | key 入力確認後に project を削除します。`-y` で prompt を skip できます。 |
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
| `tq update [-y] [--tag <tag>]` | release を install し、databases を migrate して services を再起動します。 |

log service には `tracker` または `issue-tracker`、`orchestrator`、`web` を指定できます。

`tq update` は現在のバージョンと更新先のバージョンを表示し、ローカルサービスの停止と再起動が入ることを確認してから、既定では最新の正式リリースをインストールします。その後、新しくインストールされた `tq version` を確認し、マイグレーションを実行してサービスを起動します。`-y` は確認プロンプトを省略します。`--tag` は特定のリリースまたはプレリリースのタグをインストールします。

具体的な手順とサービス停止時の注意事項は、[Tasq を更新する](pathname:///guides/update-tasq)を参照してください。
