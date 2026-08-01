---
name: tasq-cli
description: "Index for invoking the tq CLI: typed issue-tracker resources, allowlisted raw API requests, and local operations. Use for command syntax, flags, output, and operational workflows."
---

# `tq` CLI index

Use this skill to choose the smallest reference set for a `tq` task. Resource references define commands; use cases define ordered procedures. Pair with [[tq-orchestrator]] for dispatch policy: this skill covers CLI invocation only.

## Shared rules

```text
tq [--api-url URL] [--output text|json] <resource-or-command> [action-or-args] [flags]
```

- Put global flags before the resource or command. Use scoped `--help` for the installed CLI version.
- Use the installed `tq` binary; it matches the local services. `tq service start` may fall back to `go run` only for missing sibling service binaries.
- Numeric issue-tracker resource IDs are positive integers; attachment IDs are non-empty strings. `--project` always takes a kebab-case project key, not a numeric ID.
- Treat `tq api` writes and state-changing operational commands as mutations requiring explicit user authorization. For agent-created comments, pass `--author` explicitly.
- Run `tq issue watch`, `tq logs -f`, and `tq service start` under the Monitor tool rather than a blocking Bash call.

## References

Load a resource reference for command syntax and semantics. Load a use case only when its ordered procedure is needed.

| Task | Reference |
| --- | --- |
| Shared flags, API resolution, output, exit codes | [globals.md](references/resources/globals.md) |
| Status, priority, comment-type values and shortcuts | [enums.md](references/resources/enums.md) |
| Issues, including `watch` | [issue.md](references/resources/issue.md) |
| Comments | [comment.md](references/resources/comment.md) |
| Projects | [project.md](references/resources/project.md) |
| Workflow overrides | [workflow.md](references/resources/workflow.md) |
| Allowlisted raw API request | [api.md](references/resources/api.md) |
| Local migrations | [migrate.md](references/resources/migrate.md) |
| Services, logs, Web UI, version, update, configuration | [service-and-logs.md](references/resources/service-and-logs.md) |

| Procedure | Reference |
| --- | --- |
| Register and validate a repository | [bootstrap-project.md](references/usecases/bootstrap-project.md) |
| Create, progress, and finish an issue | [issue-lifecycle.md](references/usecases/issue-lifecycle.md) |
| Record progress, blockers, and handoffs | [comment-flows.md](references/usecases/comment-flows.md) |
| Attach an image | [attachments.md](references/usecases/attachments.md) |
| Stream queued issues for dispatch | [watch-and-dispatch.md](references/usecases/watch-and-dispatch.md) |
| Operate local services | [operate-services.md](references/usecases/operate-services.md) |
| Apply or roll back migrations | [run-migrations.md](references/usecases/run-migrations.md) |
| Consume JSON from a script | [script-with-json.md](references/usecases/script-with-json.md) |
