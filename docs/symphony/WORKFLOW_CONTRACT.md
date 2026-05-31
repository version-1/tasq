# Tasq Symphony Workflow Contract

Tasq loads `WORKFLOW.md` once at orchestrator process startup. Runtime reload is intentionally not
enabled; operators should restart the orchestrator after changing workflow runtime settings.

The front matter parser supports the Tasq-specific scalar fields below. It is not a general-purpose
YAML parser.

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
```

Relative `workspace.root` and `workspace.source` values resolve relative to the selected workflow
file. Environment variable indirection and `~` expansion are supported for path fields.

Continuation turns are disabled by default even when `agent.max_turns` is greater than one. Enable
`agent.continuation_turns_enabled` when the workflow is prepared for multiple turns on the same live
Codex thread.

Runner progress, workspace metadata, workspace setup failures, and cleanup state are stored in the
orchestrator SQLite database. Large transcript artifacts are not written to separate filesystem
artifacts by the current implementation.
