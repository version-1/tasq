# Enum reference

Used by `tq issue create`, `tq issue update`, `tq comment add`, and the status shortcuts.

## Issue status

| Value | Meaning |
| --- | --- |
| `backlog` | Default on create when status is omitted. Set via `tq issue draft <id>`. |
| `ready` | Eligible to be picked up. Set via `tq issue ready <id>`. |
| `in_progress` | Work has started. |
| `review` | Work is awaiting review. |
| `done` | Completed. Set via `tq issue close <id>`. |
| `blocked` | Cannot proceed; record the reason as a `blocker` comment. |
| `failed` | Work attempted but did not succeed. |
| `cancelled` | Will not be worked on. Set via `tq issue cancel <id>`. |
| `duplicate` | Same as another issue. |

## Issue priority

| Value | Notes |
| --- | --- |
| `low` | |
| `normal` | Default when `--priority` is omitted on create. |
| `high` | |
| `urgent` | |

## Comment type

| Value | Notes |
| --- | --- |
| `general` | Default. |
| `progress` | Status update during in-progress work. |
| `blocker` | Records why work cannot proceed; pair with status `blocked`. |
| `handoff` | Records context being handed to a subsequent agent or human. |

## Status shortcuts

These wrap `tq issue update --status X` so an agent does not have to spell the enum:

| Shortcut | Equivalent status |
| --- | --- |
| `tq issue close <id>` | `done` |
| `tq issue cancel <id>` | `cancelled` |
| `tq issue ready <id>` | `ready` |
| `tq issue draft <id>` | `backlog` |

For every other status (`in_progress`, `review`, `blocked`, `failed`, `duplicate`), use `tq issue update <id> --status <value>` directly.
