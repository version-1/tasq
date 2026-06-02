# Makefile Reference

The repository Makefile is the primary entry point for local development. It wraps Docker Compose, starts the dev container, runs service processes inside that container, and exposes the assigned local URLs.

Use `make help` to print the target list generated from the Makefile comments.

## Configuration Variables

| Variable | Default | Description |
|---|---|---|
| `COMPOSE` | `docker compose` | Docker Compose command. Override when using a wrapper or alternate Compose binary. |
| `BROWSER_OPEN` | `open` | Browser opener used by `dev-open`. Set to another command, or to a no-op command in headless environments. |
| `TQ_HOME` | `./.tasq` | Repository-local Tasq runtime state on the host. The dev container uses `/workspace/.tasq`. |
| `ISSUE_TRACKER_PORT` | empty | Host port for issue-tracker. Empty means Docker Compose assigns a free port. |
| `ORCHESTRATOR_PORT` | empty | Host port for orchestrator. Empty means Docker Compose assigns a free port. |
| `OPENAPI_PORT` | empty | Host port for OpenAPI UI. Empty means Docker Compose assigns a free port. |
| `WEB_PORT` | empty | Host port for Web UI. Empty means Docker Compose assigns a free port. |
| `WEB_ISSUE_TRACKER_URL` | empty | Issue-tracker URL passed to the Web UI. Empty means the Makefile resolves the assigned issue-tracker port. |
| `AIR_VERSION` | `v1.52.3` | Air version used to run Go services in watch mode. |

Example with fixed ports:

```sh
ISSUE_TRACKER_PORT=8080 ORCHESTRATOR_PORT=8081 OPENAPI_PORT=8082 WEB_PORT=3000 make dev-up
```

## Main Development Targets

| Target | Purpose |
|---|---|
| `make dev-up` | Start the `dev` container and OpenAPI UI, start issue-tracker, orchestrator, and Web in the background, print URLs, then open OpenAPI and Web in a browser. |
| `make dev-up-d` | Alias for `dev-up`. |
| `make dev-up-forward` | Start OpenAPI UI and run issue-tracker, orchestrator, and Web in the foreground inside the dev container. |
| `make dev-restart` | Stop Compose services and run `dev-up` again. |
| `make dev-down` | Stop Compose services. |
| `make dev-rebuild-schema CONFIRM=1` | Stop Compose, remove local SQLite files under `.tasq/system/data/`, and start the dev environment again. |

`dev-up` uses automatic host port assignment by default. Always run `make dev-ports` after startup when you need the current URLs.

## Service Process Targets

| Target | Purpose |
|---|---|
| `make dev-start-processes` | Start issue-tracker, orchestrator, and Web inside the already-running dev container. |
| `make dev-stop-processes` | Stop Air and Next.js processes inside the dev container without stopping the container. |
| `make issue-tracker-up` | Start the dev container if needed and start only the issue-tracker process. |
| `make orchestrator-up` | Start issue-tracker, then start the orchestrator process. |
| `make web-up` | Start issue-tracker, then start the Web process. |
| `make openapi-up` | Start only the OpenAPI UI Compose service and print ports. |
| `make tui-up` | Run the TUI interactively inside the dev container. |

Background process logs are written under `.tmp/dev-logs/`.

## Inspection Targets

| Target | Purpose |
|---|---|
| `make dev-ps` | Show Docker Compose service status. |
| `make dev-ports` | Print the currently assigned issue-tracker, orchestrator, OpenAPI, and Web URLs. |
| `make dev-logs` | Follow `.tmp/dev-logs/*.log`. |
| `make dev-shell` | Open a shell in the running dev container as the `codex` user. |
| `make dev-exec CMD="..."` | Run an arbitrary command inside the dev container as the `codex` user. |
| `make dev-ready` | Wait until required dev container tools and volumes are writable by the `codex` user. |

Useful examples:

```sh
make dev-ports
make dev-shell
make dev-exec CMD="go test ./internal/config"
```

## Verification Targets

| Target | Purpose |
|---|---|
| `make dev-test` | Run `go test ./...`, install Web dependencies, and run Web typecheck inside the dev container. |
| `make dev-build-app` | Run `go test ./...`, install Web dependencies, and run the Web production build inside the dev container. |
| `make codex-check` | Confirm Codex CLI and `codex app-server` are available inside the dev container. |

## CLI And Codex Targets

| Target | Purpose |
|---|---|
| `make tq ARGS="..."` | Start issue-tracker if needed and run `go run ./cmd/tq $(ARGS)` inside the dev container. |
| `make codex-login` | Run `codex login --device-auth` inside the dev container and persist credentials in the `codex-home` Docker volume. |

Use device auth for container login because browser redirects to a localhost callback inside the container are not reachable from the host browser.

Examples:

```sh
make tq ARGS="issue list"
make codex-login
make codex-check
```

## Operational Notes

The `dev` container is long-lived. Service processes are ordinary child processes launched with `docker compose exec`; they are not separate Compose services. Use `make dev-stop-processes` to stop only those child processes, or `make dev-down` to stop the Compose services.

The Makefile runs container commands as the `codex` user. At container startup, Compose prepares writable named volumes for Go module cache, Go build cache, Web `node_modules`, and Codex credentials.

`NEXT_PUBLIC_ISSUE_TRACKER_URL` is resolved when the Web process starts. If the issue-tracker host port changes after Web has started, restart the Web process with `make web-up` or restart everything with `make dev-up`.
