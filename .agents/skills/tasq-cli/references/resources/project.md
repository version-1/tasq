# `tq project`

Manage project registration. A project ties a repository directory to a key used by every other `tq` command's `--project` flag.

## Actions

| Action | Usage |
| --- | --- |
| `add` | `tq project add [--key KEY] [path]` |
| `remove` | `tq project remove <project-key>` |
| `check` | `tq project check [project-key]` |
| `list` | `tq project list` |

## `add`

- `path` (positional, optional) — project root. Defaults to `.` (cwd). Must be an existing directory.
- `--key` — project key. When omitted, derived as the kebab-case of the directory name.
- Side effects on success:
  - Creates `WORKFLOW.md` with the default template if it does not already exist.
  - Appends `.worktrees` to `.gitignore` (creating it if absent).
  - If any step fails the project is rolled back from the API and local file changes are reverted.

## `remove`

- `<project-key>` is required. Removes the project record from the API.

## `check`

- Without an argument, infers the project from the current directory.
- With `[project-key]`, checks the named project.
- Verifies `WORKFLOW.md` exists, has the required front matter, and that the API accepts the workflow. Also verifies `AGENTS.md` exists.
- Exits with code `1` if any check fails (`project check failed`).

## `list`

- Lists every registered project (key, name, location). No flags. Honors `--output text|json`.

## See also

- [usecases/bootstrap-project.md](../usecases/bootstrap-project.md) — full add → workflow add → check flow.
- [workflow.md](workflow.md) — managing the workflow override after `project add`.
