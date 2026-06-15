---
id: cli-reference
title: CLI リファレンス
sidebar_position: 2
---

# CLI リファレンス

`tq` は、Tasq の issue 管理、project setup、local services、logs、migrations、Web UI を扱う command-line interface です。

## グローバルフラグ

```sh
tq [--api-url URL] [--output text|json] <resource> <action> [flags]
```

| フラグ | 説明 |
| --- | --- |
| `--api-url URL` | issue-tracker API URL。environment と state discovery より優先されます。 |
| `--output text\|json` | output format。default は `text` です。 |

`--api-url` を省略した場合、`tq` は `TQ_API_URL`、`$TQ_HOME/system/state.json`、`http://localhost:37651` の順に確認します。

## issue

issue-tracker API 経由で issue を管理します。

| コマンド | フラグ | 例 |
| --- | --- | --- |
| `tq issue list` | `--project <project-key>` optional | `tq issue list --project tasq` |
| `tq issue get <id>` | none | `tq issue get 1` |
| `tq issue create` | `--project`, `--title`, `--description`, `--status`, `--priority`, `--assignee`, `--attach` | `tq issue create --project tasq --title "Add docs"` |
| `tq issue update <id>` | `--title`, `--description`, `--status`, `--priority`, `--assignee`, `--attach` | `tq issue update 1 --status in_progress` |
| `tq issue close <id>` | none | `tq issue close 1` |
| `tq issue cancel <id>` | none | `tq issue cancel 1` |
| `tq issue ready <id>` | none | `tq issue ready 1` |
| `tq issue draft <id>` | none | `tq issue draft 1` |
| `tq issue rename <id> <title>` | none | `tq issue rename 1 "Write docs"` |
| `tq issue edit <id> <description>` | none | `tq issue edit 1 "Updated description"` |

issue status は `backlog`、`ready`、`in_progress`、`review`、`done`、`blocked`、`failed` です。

priority は `low`、`normal`、`high`、`urgent` です。

## comment

issue comment を管理します。

| コマンド | フラグ | 例 |
| --- | --- | --- |
| `tq comment add <issue-id>` | `--author`, `--type`, `--body`, `--attach` | `tq comment add 1 --type progress --body "Started work."` |
| `tq comment list <issue-id>` | none | `tq comment list 1` |

comment type は `progress`、`blocker`、`handoff`、`general` です。

`--author` を省略した場合、`tq` は `TQ_AUTHOR`、configured author、`USER` environment variable の順に使います。

## project

repository の登録と workflow setup の検証を行います。

| コマンド | フラグ | 例 |
| --- | --- | --- |
| `tq project add [path]` | `--key` optional | `tq project add --key tasq .` |
| `tq project remove <project-key>` | none | `tq project remove tasq` |
| `tq project check [project-key]` | none | `tq project check tasq` |
| `tq project list` | none | `tq project list` |

`project add` は `WORKFLOW.md` がない場合に default file を作成し、`.gitignore` に `.worktrees` が含まれるようにします。

## workflow

project workflow override の管理と resolved workflow の確認を行います。

| コマンド | フラグ | 例 |
| --- | --- | --- |
| `tq workflow add` | `--project`, `--file` または `--body` のどちらか 1 つ | `tq workflow add --project tasq --file WORKFLOW.md` |
| `tq workflow remove` | `--project` | `tq workflow remove --project tasq` |
| `tq workflow show` | `--project`, `--json` optional | `tq workflow show --project tasq` |

`workflow show` は project-local file、database override、global workflow file の順に resolved workflow を探します。

## service

local Tasq services を起動・停止します。

| コマンド | フラグ | 例 |
| --- | --- | --- |
| `tq service start` | none | `tq service start` |
| `tq service stop` | none | `tq service stop` |
| `tq service status` | none | `tq service status` |

`service start` は issue-tracker、orchestrator、web を fixed local ports で起動します。

## web

running Web UI を default browser で開きます。

```sh
tq web
```

Web UI は `tq service start` によって起動済みである必要があります。

## logs

`$TQ_HOME/system/log/` の service log を読みます。

| コマンド | フラグ | 例 |
| --- | --- | --- |
| `tq logs <service>` | `-n <lines>`, `-f` | `tq logs issue-tracker -n 200` |

service は `tracker`、`issue-tracker`、`orchestrator`、`web` です。

## migrate

local SQLite database migrations の apply、rollback、status inspection を行います。

| コマンド | フラグ | 例 |
| --- | --- | --- |
| `tq migrate` | none | `tq migrate` |
| `tq migrate down` | none | `tq migrate down` |
| `tq migrate status` | none | `tq migrate status` |

database が新規の場合や pending migration がある場合は、service 起動前に `tq migrate` を実行します。

## version

version information を表示します。

```sh
tq version
```
