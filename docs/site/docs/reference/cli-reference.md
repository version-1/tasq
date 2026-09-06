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
| `tq issue watch [--interval <seconds>] [--seen-ttl <seconds>] [--verbose]` | Poll ready issues and emit JSON event envelopes. |
| `tq issue close <id>` | Move an issue to `done`. |
| `tq issue cancel <id>` | Move an issue to `cancelled`. |
| `tq issue ready <id>` | Move an issue to `ready`. |
| `tq issue draft <id>` | Move an issue to `backlog`. |
| `tq issue rename <id> <title>` | Update the title. |
| `tq issue edit <id> <description>` | Update the description. |

`issue create` requires `--project` and `--title`. It accepts `--description`,
`--status`, `--priority`, `--assignee`, `--dependency <comma-separated-ids>`,
and `--attach`; omitted status and priority default to `backlog` and `normal`.
`issue update` requires at least one update flag. It accepts the same mutable
fields, uses `--dependency` to replace dependencies, and uses
`--clear-dependencies` to remove them; those two dependency flags cannot be
combined. Empty dependency values are rejected.

`--attach` accepts PNG, JPEG, GIF, or WebP images and appends an
`attachment://` Markdown reference. If updating that reference fails after an
upload, the CLI removes the uploaded attachment.

`tq issue watch` is intended for agent loops. It emits NDJSON event envelopes,
reads the ready queue, deduplicates emitted issues for the configured seen TTL,
and continues polling after transient API errors. `--interval` defaults to 30
seconds and must be positive; `--seen-ttl` defaults to 900 seconds and must be
greater than `--interval`. It ignores global `--output`.

## Artifact Commands

| Command | Purpose |
| --- | --- |
| `tq artifact set <issue-id> --type pull_request <url>` | Create or replace an issue pull-request URL. |
| `tq artifact delete <issue-id> --type pull_request` | Delete an issue pull-request URL. |

Both commands require a positive issue ID and `--type`, and support the global text and JSON output modes.

Only `pull_request` is currently supported. The URL must be an absolute
`http` or `https` URL with a host and no userinfo, up to 4,096 UTF-8 bytes.
Repeating `artifact set` for an issue and type replaces its URL.

## Comment Commands

| Command | Purpose |
| --- | --- |
| `tq comment add <issue-id> --body <body>` | Add a comment. |
| `tq comment list <issue-id>` | List comments for an issue. |

`comment add` accepts `--type` (`general` by default; allowed values are
`progress`, `blocker`, `handoff`, and `general`), `--author` (resolved from
`TQ_AUTHOR`, the configuration's `author`, then `USER`), and `--attach` for a
PNG, JPEG, GIF, or WebP image.

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

`project add` resolves its path to a host-local absolute path and verifies it
exists before registering it. `project remove` warns about deleting the
project and descendant issues, comments, attachments, workflow overrides, and
run data; it requires the exact project key unless `-y` is supplied. It fails
when the project has running runs.

`workflow show` resolves sources in this order: the registered project's
`WORKFLOW.md`, the stored project override, then `$TQ_HOME/WORKFLOW.md`.

## Runtime Commands

| Command | Purpose |
| --- | --- |
| `tq service start` | Start issue-tracker, orchestrator, and Web UI. |
| `tq service stop` | Stop local services. |
| `tq service status` | Show service status. |
| `tq orchestrator start` | Start only the local orchestrator; requires a running local issue-tracker. |
| `tq orchestrator stop` | Gracefully stop only the local orchestrator. |
| `tq orchestrator status` | Show local orchestrator status. |
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

`service start` checks for pending issue-tracker and orchestrator migrations
before launching processes and directs you to `tq migrate` if needed. It uses
ports 37651, 37652, and 37653 by default; if any is occupied, it proposes
loopback replacements and requires confirmation unless `-y` is provided.
`service stop` stops Web, orchestrator, then issue-tracker. `service status`
reports service state, PID, port, and uptime and supports JSON output.

`logs` reads files under `$TQ_HOME/system/log/`, supports `-n` and `-f`, and
does not support JSON output. `web` opens the URL from service state and fails
if the Web UI is not running.

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
