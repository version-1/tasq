# `tq project`

Register and validate repository projects. A project key identifies the repository for every `--project` flag.

## Actions

| Action | Usage |
| --- | --- |
| `add` | `tq project add [--key KEY] [path]` |
| `remove` | `tq project remove [-y] <project-key>` |
| `check` | `tq project check [project-key]` |
| `list` | `tq project list` |

## Behavior

- `add` defaults `path` to the current directory, requires an existing directory, and derives `KEY` as the directory name in kebab case when omitted. It creates an API record, creates a default `WORKFLOW.md` if absent, and appends `.worktrees` to `.gitignore`. If a later local step fails, it removes the API record and reverts the local changes it made.
- `remove` is destructive: it deletes the project and descendant server data such as issues, comments, attachments, workflow overrides, and run data. It requires explicit user authorization and asks the user to type the project key as confirmation; `-y` skips that prompt. It does not change local `WORKFLOW.md` or `.gitignore` files. The API rejects removal while the project has running runs.
- `check` uses the specified key or the project registered for the current directory. It checks `WORKFLOW.md`, validates its front matter, and asks the API to validate the workflow. It exits `1` when any check fails.
- `list` has no flags and honors the global `--output text|json` flag.

## See also

- [bootstrap-project.md](../usecases/bootstrap-project.md) for the registration procedure.
- [workflow.md](workflow.md) for stored workflow overrides.
