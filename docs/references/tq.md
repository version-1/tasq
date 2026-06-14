# tq Command Reference

`tq` is the command-line client for the issue-tracker API. It is intended for agents, workflow tools, and local development commands that need to create, inspect, update, and annotate issues without dealing with raw HTTP requests.

## Invocation

Use the Makefile target during local dev-container development:

```sh
make run-tq ARGS="issue list"
```

The target runs the installed `tq` binary inside the already-running dev container without starting, stopping, or restarting service processes. In the default workflow, `tq` resolves the issue-tracker API from `$TQ_HOME/system/state.json`.

You can also run `tq` directly in host-only workflows:

```sh
TQ_HOME=./.tasq go run ./cmd/tq --api-url http://localhost:37651 issue list
```

## Global Options

```text
tq [--api-url URL] [--output text|json] <resource|command> <action> [flags]
```

| Option | Default | Description |
|---|---|---|
| `--api-url URL` | `TQ_API_URL`, then `$TQ_HOME/system/state.json`, then `http://localhost:37651` | Issue-tracker API base URL. |
| `--output text\|json` | `text` | Output format. JSON output is intended for scripts and agents. |

## Resources

| Resource | Actions |
|---|---|
| `issue` | `create`, `get`, `list`, `update` |
| `comment` | `add`, `list` |
| `project` | `add`, `remove`, `check`, `list` |
| `service` | `start`, `stop`, `status` |
| `version` | show version information |

## Version

Print the `tq` version and the build commit.

```sh
tq version
```

Release builds installed from a versioned module or GitHub Release print the tag version when Go build metadata includes it. Local builds fall back to `dev`.

## Issues

### `issue list`

List issues. Pass `--project KEY` to limit the result to one project.

```sh
make run-tq ARGS="issue list"
```

```sh
make run-tq ARGS="issue list --project tasq"
```

JSON output:

```sh
make run-tq ARGS="--output json issue list"
```

### `issue get`

Get one issue by numeric ID.

```sh
make run-tq ARGS="issue get 1"
```

### `issue create`

Create an issue. `--project` is required and must reference an existing project key.

```sh
make run-tq ARGS='issue create --project tasq --title "Write tq reference"'
```

Flags:

| Flag | Required | Description |
|---|---:|---|
| `--project KEY` | yes | Project key that owns the issue. |
| `--title TITLE` | yes | Issue title. |
| `--description TEXT` | no | Issue description. |
| `--status STATUS` | no | Issue status. Defaults to `backlog` when omitted. |
| `--priority PRIORITY` | no | Issue priority. Defaults to `normal` when omitted. |
| `--assignee NAME` | no | Assignee name. |
| `--attach PATH` | no | Upload a PNG, JPEG, GIF, or WebP image and append a Markdown image reference to the description. |

Example:

```sh
make run-tq ARGS='issue create --project tasq --title "Improve project list" --description "Render project list as a readable table." --status ready --priority high --assignee codex'
```

### `issue update`

Update one or more fields on an issue.

```sh
make run-tq ARGS='issue update 1 --status in_progress'
```

At least one update flag is required.

Flags:

| Flag | Description |
|---|---|
| `--title TITLE` | Replace the issue title. |
| `--description TEXT` | Replace the issue description. |
| `--status STATUS` | Replace the issue status. |
| `--priority PRIORITY` | Replace the issue priority. |
| `--assignee NAME` | Replace the issue assignee. |
| `--attach PATH` | Upload a PNG, JPEG, GIF, or WebP image and append a Markdown image reference to the description. |

Attachment references use `![filename](attachment://<id>)`. The issue-tracker serves those images through the attachment content API, and the Web UI renders them from Markdown.

## Comments

### `comment add`

Add a comment to an issue.

```sh
make run-tq ARGS='comment add 1 --body "Started implementation."'
```

Flags:

| Flag | Required | Description |
|---|---:|---|
| `--body TEXT` | yes | Comment body. |
| `--author NAME` | no | Comment author. Defaults to `TQ_AUTHOR`, then `USER`. |
| `--type TYPE` | no | Comment type. Defaults to `general`. |
| `--attach PATH` | no | Upload a PNG, JPEG, GIF, or WebP image and append a Markdown image reference to the comment body. |

### `comment list`

List comments for an issue.

```sh
make run-tq ARGS="comment list 1"
```

## Services

### `service start`

Start issue-tracker and orchestrator as host-local background processes. The command starts issue-tracker first, waits for its health endpoint, and then starts orchestrator. Logs are appended under `$TQ_HOME/system/log/`.

```sh
TQ_HOME=./.tasq go run ./cmd/tq service start
```

Default service ports:

| Service | Port | Log |
|---|---:|---|
| issue-tracker | `37651` | `$TQ_HOME/system/log/issue-tracker.log` |
| orchestrator | `37652` | `$TQ_HOME/system/log/orchestrator.log` |

### `service status`

Show service state, PID, port, and uptime. JSON output is available for scripts.

```sh
TQ_HOME=./.tasq go run ./cmd/tq service status
```

```sh
TQ_HOME=./.tasq go run ./cmd/tq --output json service status
```

### `service stop`

Stop orchestrator first and issue-tracker second. Each process receives `SIGTERM`; if it does not exit within the grace period, it is killed.

```sh
TQ_HOME=./.tasq go run ./cmd/tq service stop
```

## Projects

### `project add`

Register a local repository as a project.

```sh
make run-tq ARGS="project add ."
```

By default, `project add` uses the current directory. It resolves the path to a host-local absolute path and checks that it exists locally before sending it to the issue-tracker API.

Flags:

| Flag | Description |
|---|---|
| `--key KEY` | Project key. Defaults to a kebab-case key derived from the project directory name. |

Examples:

```sh
make run-tq ARGS='project add --key tasq .'
make run-tq ARGS='project add ../another-project'
```

### `project list`

List registered projects.

```sh
make run-tq ARGS="project list"
```

Use JSON for machine-readable output:

```sh
make run-tq ARGS="--output json project list"
```

### `project check`

Check local project workflow files.

```sh
make run-tq ARGS="project check"
make run-tq ARGS="project check tasq"
```

When no project key is provided, `project check` tries to find the project registered for the current directory.

### `project remove`

Remove a project by key.

```sh
make run-tq ARGS="project remove tasq"
```

## Workflow Commands

### `workflow remove`

Remove the database workflow override for a project. After removal, workflow resolution falls back to the project `WORKFLOW.md` file or the global workflow fallback.

```sh
make run-tq ARGS="workflow remove --project tasq"
```

## Valid Values

### Issue Status

```text
backlog
ready
in_progress
review
blocked
failed
done
```

### Issue Priority

```text
low
normal
high
urgent
```

### Comment Type

```text
progress
blocker
handoff
general
```

## Path Handling

Project paths are stored as host-local absolute paths. This means `project add .` records the path as seen by the user on the host machine, not as a container-only runtime path such as `/workspace`.

The issue-tracker API validates that project paths are absolute, but it does not check whether they exist on the API server filesystem. The `tq project add` client performs the local existence check before creating project records.
