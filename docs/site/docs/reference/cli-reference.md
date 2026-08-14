---
id: cli-reference
title: CLI Reference
sidebar_position: 1
---

# CLI Reference

`tq` is the command-line interface for issue management, project setup, workflow configuration, raw API access, local services, logs, migrations, and the Web and terminal UIs.

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
| `tq issue cancel <id>` | Move an issue to `cancelled`. |
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

## Artifact Commands

| Command | Purpose |
| --- | --- |
| `tq artifact set <issue-id> --type pull_request <url>` | Create or replace an issue pull-request URL. |
| `tq artifact delete <issue-id> --type pull_request` | Delete an issue pull-request URL. |

Both commands require a positive issue ID and `--type`, and support the global text and JSON output modes.

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
| `tq tui` | Open the experimental read-only terminal UI. Aliases: `tq console`, `tq c`. |
| `tq config` | Show build, home-directory, and resolved configuration information. |
| `tq version` | Print version information. |
| `tq update [-y] [--tag <tag>]` | Install a release, migrate databases, and restart services. |

Log services are `tracker` or `issue-tracker`, `orchestrator`, and `web`.

`tq tui [--orchestrator-url URL]` requires a terminal and supports only text output. It reads issues, comments, artifacts, and run state without sending mutation requests. Use `--orchestrator-url` to override orchestrator discovery from `state.json`.

`tq config` prints the version, build profile, `TQ_HOME` override, resolved home directory, configuration file path, and resolved values. It does not print the raw YAML. Use global `--output json` for scripts.

`tq update` prints the current and target versions, confirms that local services will stop and restart, installs the latest formal release by default, verifies the newly installed `tq version`, runs migrations, and starts services. `-y` skips the confirmation prompt. `--tag` installs a specific release or prerelease tag.

`tq update` is unavailable when the binary has a non-empty build profile such as `dev`, because generic release artifacts do not retain that profile.

For step-by-step examples and service interruption guidance, see
[Update Tasq](pathname:///guides/update-tasq).

## Raw API Command

Use `tq api` for an issue-tracker operation that has no typed command:

```sh
tq api GET /api/v1/issues --query states=ready
tq api POST /api/v1/issues --header 'X-Request-ID: local-123' --data @request.json
```

```text
tq api <method> <path> [--query key=value] [--header 'Name: value'] [--data value|@file|-]
```

The path must be an allowlisted, unencoded absolute `/api/v1/...` path. Complete URLs, fragments, dot segments, empty segments, and trailing slashes are rejected. `--query` and `--header` are repeatable. `--data` accepts a literal value, `@file`, or `-` for standard input and is available only for `POST`, `PUT`, and `PATCH`.

The command does not prompt before writes or deletes, does not follow redirects, and times out after 10 seconds. Response bytes are written unchanged, so global `--output` does not transform them. Exit status is `0` for HTTP `2xx`, `1` for HTTP or transport failures, and `2` for invalid usage, input, or a request outside the allowlist.
