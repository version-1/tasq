# `tq workflow`

Manage per-project workflow overrides. The workflow content drives the orchestration rules — it must include YAML front matter at the top of the file or body.

## Actions

| Action | Usage |
| --- | --- |
| `add` | `tq workflow add --project KEY (--file PATH \| --body TEXT)` |
| `remove` | `tq workflow remove --project KEY` |
| `show` | `tq workflow show --project KEY [--json]` |

## `add`

- `--project KEY` is required.
- Exactly one of `--file PATH` or `--body TEXT` must be supplied. Passing both, or neither, is a usage error.
- The content must begin with YAML front matter (delimited by `---`); the front matter must parse as a YAML object and supply the workflow's required fields. Otherwise the call fails with `workflow front matter is required` / `workflow front matter must be a YAML object` / `missing fields …`.

## `remove`

- `--project KEY` is required. Clears the project's stored override; resolution falls back to the default workflow.

## `show`

- `--project KEY` is required.
- Resolution order (first hit wins):
  1. The project's local `WORKFLOW.md` on disk (`<project location>/WORKFLOW.md`).
  2. The DB-stored override added via `tq workflow add`.
  3. The global workflow file under `$TQ_HOME`.
  - Fails with `workflow not found` if none of the three exist.
- The result includes a `source.type` field (`file` / `db` / `global`) so a caller can tell which layer answered.
- `--json` emits structured JSON for the show command. Setting the global `--output json` produces the same JSON, so either flag works.

## See also

- [project.md](project.md) — `tq project add` seeds `WORKFLOW.md` from the default template; `tq workflow add` is for overriding what the API uses.
- [usecases/bootstrap-project.md](../usecases/bootstrap-project.md) — typical bootstrap also pushes the override.
