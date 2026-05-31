# Tasq Symphony Workflow Contract

Tasq loads `WORKFLOW.md` once at orchestrator process startup. Runtime reload is intentionally not
enabled; operators should restart the orchestrator after changing workflow runtime settings.

The front matter parser uses YAML and supports the Tasq-specific fields below. Unknown fields are
ignored for forward compatibility.

## Supported Fields

```yaml
polling:
  interval_ms: 30000
workspace:
  root: .workspaces
  source: .
agent:
  max_concurrent_agents: 10
  max_turns: 20
  continuation_turns_enabled: false
  max_retry_attempts: 3
  max_retry_backoff_ms: 300000
codex:
  command: codex app-server
  read_timeout_ms: 5000
  turn_timeout_ms: 3600000
  stall_timeout_ms: 300000
server:
  port: 8080
hooks:
  after_create: |
    echo "created workspace"
  before_run: |
    echo "before agent run"
  after_run: |
    echo "after agent run"
  before_remove: |
    echo "before workspace cleanup"
  timeout_ms: 60000
```

Relative `workspace.root` and `workspace.source` values resolve relative to the selected workflow
file. Environment variable indirection and `~` expansion are supported for path fields.

Continuation turns are disabled by default even when `agent.max_turns` is greater than one. Enable
`agent.continuation_turns_enabled` when the workflow is prepared for multiple turns on the same live
Codex thread.

Runner progress, workspace metadata, workspace setup failures, and cleanup state are stored in the
orchestrator SQLite database. Large transcript artifacts are not written to separate filesystem
artifacts by the current implementation.

Workspace hooks run through `bash -lc` with the issue workspace directory as `cwd`.
`hooks.timeout_ms` applies to all hooks and defaults to `60000` milliseconds. Non-positive timeout
values fail workflow configuration validation.

Hook failure behavior:

- `after_create` runs only for newly created workspaces. Failure aborts workspace creation and
  removes the partial workspace directory.
- `before_run` runs before each agent attempt. Failure aborts the current attempt.
- `after_run` runs after each agent attempt. Failure is logged and ignored.
- `before_remove` runs before workspace deletion when the directory exists. Failure is logged and
  cleanup proceeds.

Because front matter is YAML, multiline hook scripts can use literal block scalars (`|`).
