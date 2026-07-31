---
name: tasq-cli
description: Index for invoking the tq CLI, including typed issue-tracker resources, allowlisted raw API requests, local services, migrations, and release updates. Use for tq command syntax, flags, output, and operational workflows.
---

# Purpose

Drive the `tq` CLI (issue tracker, comments, projects, workflows, allowlisted raw API requests, migrations, local services, release updates). This file is the index: load only the reference(s) that match the task so the agent's working context stays small.

Pair with [[tq-orchestrator]] when polling and dispatching ready issues — that skill covers *when* to call `tq`; this one covers *how*.

# Invocation shape

```
tq [--api-url URL] [--output text|json] <resource> <action> [flags] [positional args]
```

- Global flags must appear *before* the resource. `--flag value` and `--flag=value` both work.
- `tq <resource> --help` / `tq <resource> <action> --help` prints scoped help.
- `tq api` is a command rather than an action resource: `tq api <method> <path> [flags]`.
- Prefer the installed `tq` binary so the CLI version matches the running services. `tq service start` will fall back to `go run ./cmd/<service>` when a sibling service binary is missing, but `tq` itself should be the installed binary.

# Pick the reference

Load the smallest set of references that matches the task. Each file is self-contained.

## By resource (`references/resources/`)

| Resource | When to load |
| --- | --- |
| [issue.md](references/resources/issue.md) | create, get, list, update, status transitions, rename, edit, watch |
| [comment.md](references/resources/comment.md) | add or list issue comments, set author / type |
| [project.md](references/resources/project.md) | register, remove, check, list projects |
| [workflow.md](references/resources/workflow.md) | add, remove, show project workflow overrides |
| [api.md](references/resources/api.md) | call an allowlisted raw issue-tracker endpoint when no typed command exposes the operation |
| [migrate.md](references/resources/migrate.md) | apply, roll back, inspect local DB migrations |
| [service-and-logs.md](references/resources/service-and-logs.md) | start / stop / status of services, tail logs, open web UI, version, update |
| [enums.md](references/resources/enums.md) | valid status / priority / comment-type values |
| [globals.md](references/resources/globals.md) | global flags, environment variables, API URL resolution, output / exit codes |

## By use case (`references/usecases/`)

| Goal | Load |
| --- | --- |
| Register a repo with tasq end-to-end | [bootstrap-project.md](references/usecases/bootstrap-project.md) |
| Move an issue through backlog → done | [issue-lifecycle.md](references/usecases/issue-lifecycle.md) |
| Record progress, blockers, or handoffs as comments | [comment-flows.md](references/usecases/comment-flows.md) |
| Upload an image to an issue or comment | [attachments.md](references/usecases/attachments.md) |
| Stream ready issues for the orchestrator | [watch-and-dispatch.md](references/usecases/watch-and-dispatch.md) |
| Boot, stop, inspect, or update local services | [operate-services.md](references/usecases/operate-services.md) |
| Apply or roll back migrations safely | [run-migrations.md](references/usecases/run-migrations.md) |
| Parse `tq` output from a script | [script-with-json.md](references/usecases/script-with-json.md) |

# Conventions

- IDs are positive integers. `tq` rejects `0` and negative values.
- `--project` takes the *project key* (kebab-case), not the numeric id.
- Long-running commands (`tq issue watch`, `tq logs -f`, `tq service start`) belong under the Monitor tool, not a blocking Bash call.
- When an agent posts comments programmatically, always pass `--author` explicitly (e.g. `--author claude-code`) so the source is attributable.
