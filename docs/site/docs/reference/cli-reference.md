---
id: cli-reference
title: CLI Reference
sidebar_position: 1
---

# CLI Reference

`tq` is the command-line interface for issue management, project setup, workflow configuration, local services, logs, migrations, and the Web UI.

## Global Form

```sh
tq [--api-url URL] [--output text|json] <resource> <action> [flags]
```

| Flag | Description |
| --- | --- |
| `--api-url URL` | Issue-tracker API URL. Overrides environment and state discovery. |
| `--output text\|json` | Output format. Defaults to `text`. |

API URL resolution order is `--api-url`, `TQ_API_URL`, `$TQ_HOME/system/state.json`, then `http://localhost:37651`.

## Issue Commands

| Command | Purpose |
| --- | --- |
| `tq issue list [--project <key>]` | List issues, optionally scoped to one project. |
| `tq issue get <id>` | Show one issue. |
| `tq issue create --project <key> --title <title>` | Create an issue. |
| `tq issue update <id> [flags]` | Update issue fields. |
| `tq issue watch [--interval <duration>] [--seen-ttl <duration>] [--verbose]` | Poll ready issues and emit JSON event envelopes. |
| `tq issue close <id>` | Move an issue to `done`. |
| `tq issue cancel <id>` | Move an issue to `failed`. |
| `tq issue ready <id>` | Move an issue to `ready`. |
| `tq issue draft <id>` | Move an issue to `backlog`. |
| `tq issue rename <id> <title>` | Update the title. |
| `tq issue edit <id> <description>` | Update the description. |

Create and update accept `--title`, `--description`, `--status`, `--priority`,
`--assignee`, and `--attach` where applicable. Update also accepts
`--dependency <comma-separated-ids>` to replace dependencies and
`--clear-dependencies` to remove them.

`tq issue watch` is intended for agent loops. It reads the ready queue,
deduplicates emitted issues for the configured seen TTL, emits `issue-ready`
events, and continues polling after transient API errors.

## Comment Commands

| Command | Purpose |
| --- | --- |
| `tq comment add <issue-id> --body <body>` | Add a comment. |
| `tq comment list <issue-id>` | List comments for an issue. |

Allowed comment types are `progress`, `blocker`, `handoff`, and `general`.

## Project and Workflow Commands

| Command | Purpose |
| --- | --- |
| `tq project add [path] [--key <key>]` | Register a repository. |
| `tq project remove [-y] <key>` | Remove a project after key confirmation, or skip the prompt with `-y`. |
| `tq project check [key]` | Validate project setup. |
| `tq project list` | List registered projects. |
| `tq workflow add --project <key> (--file <path> \| --body <text>)` | Store a workflow override. |
| `tq workflow remove --project <key>` | Remove the stored override. |
| `tq workflow show --project <key> [--json]` | Show the resolved workflow. |

## Runtime Commands

| Command | Purpose |
| --- | --- |
| `tq service start` | Start issue-tracker, orchestrator, and Web UI. |
| `tq service stop` | Stop local services. |
| `tq service status` | Show service status. |
| `tq logs <service> [-n <lines>] [-f]` | Read service logs. |
| `tq migrate` | Apply migrations. |
| `tq migrate down` | Roll back migrations. |
| `tq migrate status` | Show migration status. |
| `tq web` | Open the running Web UI. |
| `tq version` | Print version information. |
| `tq update [-y] [--tag <tag>]` | Install a release, migrate databases, and restart services. |

Log services are `tracker` or `issue-tracker`, `orchestrator`, and `web`.

`tq update` prints the current and target versions, confirms that local services will stop and restart, installs the latest formal release by default, verifies the newly installed `tq version`, runs migrations, and starts services. `-y` skips the confirmation prompt. `--tag` installs a specific release or prerelease tag.
