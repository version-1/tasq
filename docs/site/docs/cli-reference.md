---
id: cli-reference
title: CLI Reference
sidebar_position: 2
---

# CLI Reference

`tq` is the command-line interface for Tasq issue management, project setup, local services, logs, migrations, and the Web UI.

## Global Flags

```sh
tq [--api-url URL] [--output text|json] <resource> <action> [flags]
```

| Flag | Description |
| --- | --- |
| `--api-url URL` | Issue-tracker API URL. Overrides environment and state discovery. |
| `--output text\|json` | Output format. Defaults to `text`. |

When `--api-url` is omitted, `tq` checks `TQ_API_URL`, `$TQ_HOME/system/state.json`, and then `http://localhost:37651`.

## issue

Manage issues through the issue-tracker API.

| Command | Flags | Example |
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

Allowed issue statuses are `backlog`, `ready`, `in_progress`, `review`, `done`, `blocked`, and `failed`.

Allowed priorities are `low`, `normal`, `high`, and `urgent`.

## comment

Manage comments for an issue.

| Command | Flags | Example |
| --- | --- | --- |
| `tq comment add <issue-id>` | `--author`, `--type`, `--body`, `--attach` | `tq comment add 1 --type progress --body "Started work."` |
| `tq comment list <issue-id>` | none | `tq comment list 1` |

Allowed comment types are `progress`, `blocker`, `handoff`, and `general`.

When `--author` is omitted, `tq` uses `TQ_AUTHOR`, the configured author, or the `USER` environment variable.

## project

Register repositories and validate workflow setup.

| Command | Flags | Example |
| --- | --- | --- |
| `tq project add [path]` | `--key` optional | `tq project add --key tasq .` |
| `tq project remove <project-key>` | none | `tq project remove tasq` |
| `tq project check [project-key]` | none | `tq project check tasq` |
| `tq project list` | none | `tq project list` |

`project add` creates a default `WORKFLOW.md` when one is missing and ensures `.worktrees` is present in `.gitignore`.

## workflow

Manage project workflow overrides and inspect the resolved workflow.

| Command | Flags | Example |
| --- | --- | --- |
| `tq workflow add` | `--project`, exactly one of `--file` or `--body` | `tq workflow add --project tasq --file WORKFLOW.md` |
| `tq workflow remove` | `--project` | `tq workflow remove --project tasq` |
| `tq workflow show` | `--project`, `--json` optional | `tq workflow show --project tasq` |

`workflow show` resolves project-local files, database overrides, and the global workflow file in priority order.

## service

Run local Tasq services.

| Command | Flags | Example |
| --- | --- | --- |
| `tq service start` | none | `tq service start` |
| `tq service stop` | none | `tq service stop` |
| `tq service status` | none | `tq service status` |

`service start` starts issue-tracker, orchestrator, and web on fixed local ports.

## web

Open the running Web UI in the default browser.

```sh
tq web
```

The Web UI must already be running through `tq service start`.

## logs

Read service logs from `$TQ_HOME/system/log/`.

| Command | Flags | Example |
| --- | --- | --- |
| `tq logs <service>` | `-n <lines>`, `-f` | `tq logs issue-tracker -n 200` |

Services are `tracker`, `issue-tracker`, `orchestrator`, and `web`.

## migrate

Apply, roll back, or inspect local SQLite database migrations.

| Command | Flags | Example |
| --- | --- | --- |
| `tq migrate` | none | `tq migrate` |
| `tq migrate down` | none | `tq migrate down` |
| `tq migrate status` | none | `tq migrate status` |

Run `tq migrate` before starting services when databases are new or migrations are pending.

## version

Print version information.

```sh
tq version
```
