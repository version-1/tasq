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

Relative `workspace.root` values resolve relative to the selected workflow file. Environment
variable indirection and `~` expansion are supported for path fields.

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

## Prompt Template

Everything after the closing `---` of the front matter is the prompt template. The orchestrator
renders it once per agent attempt and sends the result as the initial message to the coding agent.

### Available Variables

| Variable                 | Type   | Description                                      |
|--------------------------|--------|--------------------------------------------------|
| `{{ issue.id }}`         | int    | Numeric issue ID from the issue-tracker           |
| `{{ issue.title }}`      | string | Issue title                                       |
| `{{ issue.description }}`| string | Issue description body                            |
| `{{ attempt }}`          | int    | Attempt number (0 for first run, >=1 for retries) |

Variables are replaced by simple string substitution. Unrecognized `{{ ... }}` tokens are left
as-is.

### Authoring Guidelines

A prompt template should tell the agent **what to do** and **what tools are available** for
interacting with the issue-tracker.

#### Issue Status Updates

The agent updates issue status through the `tq` CLI. The runtime environment must have `tq` on
`PATH` and `TQ_API_URL` set to the issue-tracker endpoint.

Example instruction in a prompt template:

```
When work is complete, update the issue status:

  tq issue update {{ issue.id }} -status review
```

#### Deliverables

Define the expected deliverables so the agent knows when the task is done. Common patterns:

- Commit changes and open a pull request.
- Update issue status to a handoff state (e.g., `review`).
- Leave implementation notes in the issue description or a comment.

Be explicit about which status the agent should transition the issue to on success and on failure.
If the agent writes a blocker comment, it should also transition the issue to `blocked` in the same
handoff. This applies even when the implementation is locally complete but push, pull request
creation, verification, or another handoff step failed.

#### What Not to Include

- Runtime configuration (poll intervals, concurrency, timeouts) belongs in front matter, not in
  the prompt.
- Secrets and credentials should come from the environment, not from the template text.
