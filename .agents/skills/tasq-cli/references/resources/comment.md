# `tq comment`

Add and list issue comments. Comment types live in [enums.md](enums.md).

## Actions

| Action | Usage |
| --- | --- |
| `add` | `tq comment add <issue-id> --body TEXT [--author NAME] [--type general\|progress\|blocker\|handoff] [--attach PATH]` |
| `list` | `tq comment list <issue-id>` |

## `add`

- `<issue-id>` is a positive integer positional.
- `--body` is required. Empty body is rejected.
- `--author` defaults to `TQ_AUTHOR` → config `author` → `$USER`. Agents should always pass `--author` explicitly so the source is attributable (e.g. `--author claude-code`).
- `--type` defaults to `general`. Use `progress`, `blocker`, or `handoff` to drive the orchestration UX.
- `--attach PATH` first creates the comment, then uploads the file, then patches the comment body with a Markdown image reference. If the body patch fails, the uploaded attachment is deleted, but the already-created comment row remains.

## `list`

- `<issue-id>` is a positive integer positional. No flags.
- Honors the global `--output text|json`.

## See also

- [usecases/comment-flows.md](../usecases/comment-flows.md) — common progress / blocker / handoff patterns.
- [usecases/attachments.md](../usecases/attachments.md) — full attachment flow.
