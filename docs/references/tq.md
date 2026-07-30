# tq Command Reference

`tq` is the command-line client for Tasq. It provides typed commands for common issue-tracker operations, local service and migration commands, and an allowlisted raw API command.

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
tq [--api-url URL] [--output text|json] <command> [args] [flags]
```

| Option | Default | Description |
|---|---|---|
| `--api-url URL` | `TQ_API_URL`, then `$TQ_HOME/system/state.json`, then `http://localhost:37651` | Issue-tracker API base URL. |
| `--output text\|json` | `text` | Output format. JSON output is intended for scripts and agents. |

## Commands

| Command | Actions or purpose |
|---|---|
| `issue` | `create`, `get`, `list`, `watch`, `update`, `close`, `cancel`, `ready`, `draft`, `rename`, `edit` |
| `comment` | `add`, `list` |
| `project` | `add`, `remove`, `check`, `list` |
| `workflow` | `add`, `remove`, `show` |
| `migrate` | apply pending migrations, roll back with `down`, or inspect with `status` |
| `service` | `start`, `stop`, `status` |
| `logs` | show or follow service logs |
| `web` | open the running Web UI |
| `config` | show build, home, and resolved configuration information |
| `update` | install a release and restart services |
| `version` | show version information |
| `api` | send an allowlisted raw issue-tracker API request |

## Raw API requests

Use `tq api` when the issue-tracker operation is not exposed by a typed command. It sends a raw request to the resolved issue-tracker base URL; it is not a general-purpose HTTP client. See [Tasq API](../design/api.md) for API semantics and the [OpenAPI document](../openapi/issue-tracker.yml) for the normative endpoint contract.

```sh
tq api GET /api/v1/issues --query states=ready

tq api POST /api/v1/issues --header 'X-Request-ID: local-123' --data @request.json
```

The command syntax is:

```text
tq api <method> <path> [--query key=value] [--header 'Name: value'] [--data value|@file|-]
```

Methods are case-insensitive and normalized to uppercase. The path must be an unencoded absolute `/api/v1/...` path; complete URLs, fragments, dot segments, empty segments, and trailing slashes are rejected. Query text in the path is preserved, and each repeatable `--query key=value` appends another value in order. Query names and values are passed to the API without semantic validation.

The method and path must match the CLI's explicit allowlist of current issue-tracker routes. Numeric route IDs must be positive `int64` values. This is fail-closed: a newly added server route is unavailable until the CLI allowlist is updated. `POST /api/v1/attachments` is temporarily excluded while raw multipart support is unavailable; attachment `PATCH` is not allowed.

`--header` may be repeated. Header names are case-insensitive, and the last value wins. Transport-managed headers, including `Host`, `Content-Length`, `Transfer-Encoding`, `Connection`, `Trailer`, `Upgrade`, and `Proxy-Connection`, are rejected.

`--data` accepts a literal value, `@file`, or `-` for standard input and is available only for `POST`, `PUT`, and `PATCH`. The body is not validated as JSON. A request with a body defaults to `Content-Type: application/json` unless the header is supplied explicitly.

The command does not prompt before write or delete operations, does not follow redirects, and uses a 10-second HTTP timeout. It writes response bytes unchanged to standard output, including binary data and HTTP error bodies; `--output` does not transform them. Exit status is `0` for HTTP `2xx`, `1` for HTTP `3xx`-`5xx` or transport failures, and `2` for usage, input, and allowlist errors.

## Version

Print the `tq` version and the build commit.

```sh
tq version
```

Release builds installed from a versioned module or GitHub Release print the tag version when Go build metadata includes it. Local builds fall back to `dev`.

## Configuration

`tq config` reports the version, build profile, `TQ_HOME` override, resolved home directory, configuration file path, and resolved configuration values. Pass `--output json` for machine-readable output. It does not print the configuration file's raw YAML.

## Update

Install `tq` and the sibling service executables from a GitHub Release, migrate local databases, and restart local services.

```sh
tq update
```

The command prints the current version and the target release before stopping services. It then asks for confirmation because the update stops and restarts local services. Pass `-y` to skip the confirmation prompt.

```sh
tq update -y
```

By default, `tq update` installs the latest formal release. Pass `--tag` to install a specific release or prerelease tag.

`tq update` is unavailable for binaries with a non-empty build profile because generic release artifacts do not retain that profile.

```sh
tq update --tag v0.2.0-rc.1
```

The update flow stops services, installs the release artifacts into the fixed user install location, verifies the newly installed `tq version`, applies migrations, and starts services again. If any step fails, later steps are not run.

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
| `--dependency IDS` | no | Set dependencies from a comma-separated list of issue IDs. Empty values are rejected. |
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
| `--dependency IDS` | Replace the full dependency set with a comma-separated list of issue IDs. Empty values are rejected; use `--clear-dependencies` to remove all dependencies. |
| `--clear-dependencies` | Remove all dependencies. Cannot be combined with `--dependency`. |
| `--attach PATH` | Upload a PNG, JPEG, GIF, or WebP image and append a Markdown image reference to the description. |

Attachment references use `![filename](attachment://<id>)`. The issue-tracker serves those images through the attachment content API, and the Web UI renders them from Markdown.

### `issue watch`

Poll the ready queue and emit one JSON object per line for monitoring and agent dispatch. `event` records contain newly observed queued issues; transient API failures produce `error` records without stopping the loop. The command always uses this JSON-line protocol and ignores the global `--output` setting.

| Flag | Default | Description |
|---|---:|---|
| `--interval SECONDS` | `30` | Polling interval. Must be positive. |
| `--seen-ttl SECONDS` | `900` | Suppress re-emitting an issue for this period. Must exceed `--interval`. |
| `--verbose` | disabled | Also emit polling details as `info` records. |

### Issue shortcuts

Shortcut commands update one field without requiring `issue update` flags:

```text
tq issue close <id>
tq issue cancel <id>
tq issue ready <id>
tq issue draft <id>
tq issue rename <id> <title>
tq issue edit <id> <description>
```

The status shortcuts set `done`, `cancelled`, `ready`, and `backlog`, respectively.

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

Start issue-tracker, orchestrator, and web as host-local background processes. Before starting any service process, the command opens the local issue-tracker and orchestrator databases and checks for pending migrations. If any migration is pending, it exits immediately with guidance to run `tq migrate`. Otherwise, it starts issue-tracker first, waits for its health endpoint, then starts orchestrator and web. Logs are appended under `$TQ_HOME/system/log/`.

```sh
TQ_HOME=./.tasq go run ./cmd/tq service start
```

Default service ports:

| Service | Port | Log |
|---|---:|---|
| issue-tracker | `37651` | `$TQ_HOME/system/log/issue-tracker.log` |
| orchestrator | `37652` | `$TQ_HOME/system/log/orchestrator.log` |
| web | `37653` | `$TQ_HOME/system/log/web.log` |

If any default port is already in use, `tq service start` proposes a distinct OS-selected loopback port for every service and asks for confirmation. Enter `y` or `yes` to continue; declined or non-interactive confirmation leaves services stopped. Use `tq service start -y` to accept the proposed ports without a prompt. Tasq rechecks the proposed ports after confirmation and fails rather than choosing another set if a port was claimed.

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

## Logs and Web UI

Show the last 1,000 lines of a service log:

```sh
tq logs issue-tracker
tq logs orchestrator
tq logs web
```

Use `-n LINES` to change the number of lines and `-f` to follow appended output. `tracker` is an alias for `issue-tracker`. The command reads files below `$TQ_HOME/system/log/` and does not support `--output json`.

Open the running Web UI in the default browser:

```sh
tq web
```

The command reads the Web URL from local service state and fails if the Web UI is not running.

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

Remove a project by key. By default, `project remove` prints an irreversible-operation warning, lists the project and descendant data that will be deleted, and requires typing the exact project key before deletion starts. The deletion removes the project and descendant data such as issues, comments, attachments, workflow overrides, and run data.

```sh
make run-tq ARGS="project remove tasq"
```

Pass `-y` to skip the prompt for agents and scripts.

```sh
make run-tq ARGS="project remove -y tasq"
```

If the project has running runs, the command fails before deletion and prints the API-provided reason.

## Workflows

### `workflow add`

Add or replace the database workflow override for a project.

```sh
make run-tq ARGS="workflow add --project tasq --file WORKFLOW.md"
```

### `workflow remove`

Remove the database workflow override for a project. After removal, workflow resolution falls back to the project `WORKFLOW.md` file or the global workflow fallback.

```sh
make run-tq ARGS="workflow remove --project tasq"
```

### `workflow show`

Show the resolved `WORKFLOW.md` content for a project.

```sh
make run-tq ARGS="workflow show --project tasq"
```

The command uses the same source order as workflow resolution:

1. `WORKFLOW.md` under the registered project location.
2. The stored project workflow from the issue-tracker API.
3. The global `$TQ_HOME/WORKFLOW.md`.

Text output prints a `# Source: ...` header followed by the resolved `WORKFLOW.md` content. Use `--json` or global `--output json` for structured output:

```sh
make run-tq ARGS="workflow show --project tasq --json"
```

## Migrations

### `migrate`

Apply all pending SQLite migrations for the local issue-tracker and orchestrator databases under `$TQ_HOME`.

```sh
make run-tq ARGS="migrate"
```

The command runs without starting services and writes migration state to each database's `schema_migrations` table.

### `migrate down`

Roll back one applied migration per local database.

```sh
make run-tq ARGS="migrate down"
```

### `migrate status`

List applied and pending migrations for each local database.

```sh
make run-tq ARGS="migrate status"
```

Use JSON output for scripts:

```sh
make run-tq ARGS="--output json migrate status"
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
cancelled
duplicate
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
