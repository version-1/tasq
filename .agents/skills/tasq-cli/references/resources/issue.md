# `tq issue`

Manage issues. Enum values live in [enums.md](enums.md). Global flags / env in [globals.md](globals.md).

## Actions

| Action | Usage |
| --- | --- |
| `create` | `tq issue create --project KEY --title TITLE [--description TEXT] [--status STATUS] [--priority PRIORITY] [--assignee NAME] [--dependency IDS] [--attach PATH]` |
| `get` | `tq issue get <id>` |
| `list` | `tq issue list [--project KEY]` |
| `update` | `tq issue update <id> [--title …] [--description …] [--status …] [--priority …] [--assignee …] [--dependency IDS] [--clear-dependencies] [--attach PATH]` |
| `close` | `tq issue close <id>` — sets status to `done` |
| `cancel` | `tq issue cancel <id>` — sets status to `cancelled` |
| `ready` | `tq issue ready <id>` — sets status to `ready` |
| `draft` | `tq issue draft <id>` — sets status to `backlog` |
| `rename` | `tq issue rename <id> <title>` |
| `edit` | `tq issue edit <id> <description>` — overwrites description |
| `watch` | `tq issue watch [--interval SECONDS] [--seen-ttl SECONDS] [--verbose]` |

## Required-field rules

| Action | Required |
| --- | --- |
| `create` | `--title` and `--project` (the project *key*, not the id). |
| `update` | At least one of `--title` / `--description` / `--status` / `--priority` / `--assignee` / `--dependency` / `--clear-dependencies` / `--attach`. |
| `get` / `close` / `cancel` / `ready` / `draft` | A single positive integer `<id>` positional. |
| `rename` | `<id> <title>` — title is a single positional after the id. |
| `edit` | `<id> <description>` — description is a single positional after the id. |

## Defaults applied on create

| Field | Default when omitted |
| --- | --- |
| `--status` | `backlog` |
| `--priority` | `normal` |
| `--description`, `--assignee` | empty string |

## Dependencies

`--dependency IDS` sets or replaces the full dependency set with a comma-separated list of issue IDs. Empty values are rejected.

Use `--clear-dependencies` with `tq issue update <id>` to remove all dependencies. It cannot be combined with `--dependency`.

## `--attach`

`--attach PATH` uploads the file, appends a Markdown image reference (`![alt](attachment://<id>)`) to the description, and rolls the attachment back if the issue write fails. The path is resolved from the CLI's current working directory.

See [usecases/attachments.md](../usecases/attachments.md) for the full flow.

## `tq issue watch` (experimental, NDJSON)

Streams one JSON envelope per line:

| Shape | When emitted |
| --- | --- |
| `{"type":"event","eventType":"issue-ready","body":<issue>}` | A queued issue from `GET /api/v1/queue` is detected. Always emitted. |
| `{"type":"error","body":"<message>"}` | Transient polling failure. Loop keeps running. Always emitted. |
| `{"type":"info","body":"<message>"}` | Startup config and per-cycle summary. Emitted only with `--verbose`. |

Flags:

| Flag | Default | Notes |
| --- | --- | --- |
| `--interval` | `30` | Polling interval in seconds. Must be positive. |
| `--seen-ttl` | `900` | Suppress re-emitting the same issue for this many seconds. Must be strictly greater than `--interval`. |
| `--verbose` | off | Also emit info envelopes. |

`tq issue watch` is dependency-aware: it emits only the `queued` array returned by `GET /api/v1/queue`, not every issue with `status=ready`. Pending issues with active dependencies are intentionally suppressed until the issue-tracker classifies them as queued.

`tq issue watch` ignores `--output`. Run it under the Monitor tool, not a blocking Bash call. See [usecases/watch-and-dispatch.md](../usecases/watch-and-dispatch.md).
