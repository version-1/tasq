# Makefile Reference

The repository Makefile is the primary entry point for local development. It wraps Docker Compose, starts the dev container, runs service processes inside that container, and exposes the assigned local URLs.

Use `make help` to print the prefix guide and sectioned target list generated from the Makefile comments.

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
| `WEB_ISSUE_TRACKER_URL` | empty | Reserved for compatibility; the Go Web server uses proxy configuration instead of a browser build-time API origin. |
| `RELEASE_BRANCH` | `main` | Branch required by the formal release target. |
| `RELEASE_REMOTE` | `origin` | Remote that receives release tags. |
| `RELEASE_REPO` | `version-1/tasq` | GitHub repository used when installing `tq` from release assets. |
| `TQ_INSTALL_DIR` | `$HOME/.local/bin` | Directory where release install targets place the `tq` binary. |
| `TQ_INSTALL_NAME` | `tq` | Installed command name for release install targets. |
| `AIR_VERSION` | `v1.52.3` | Air version used to run Go services in watch mode. |

Example with fixed ports:

```sh
ISSUE_TRACKER_PORT=8080 ORCHESTRATOR_PORT=8081 OPENAPI_PORT=8082 WEB_PORT=3000 make dev-up
```

## Main Development Targets

| Target | Purpose |
|---|---|
| `make dev-up` | Start the `dev` container and OpenAPI UI, start issue-tracker, orchestrator, and Web in the background, then print URLs. |
| `make dev-restart` | Stop Compose services and run `dev-up` again. |
| `make dev-down` | Stop Compose services. |
| `make dev-reset-db CONFIRM=1` | Stop Compose, remove local SQLite files under `.tasq/system/data/`, and start the dev environment again. |
| `make dev-openapi` | Start only the OpenAPI UI Compose service and print ports. |
| `make dev-open` | Open the Web UI and OpenAPI UI in a browser. |
| `make dev-ports` | Print the currently assigned issue-tracker, orchestrator, OpenAPI, and Web URLs. |

`dev-up` uses automatic host port assignment by default. It prints the assigned URLs but does not open a browser. Run `make dev-open` when you want to open the browser explicitly.

## Container Targets

Use `dc-*` targets for Docker Compose service status and commands that operate on the dev container itself.

| Target | Purpose |
|---|---|
| `make dc-ready` | Wait until required dev container tools and volumes are writable by the `codex` user. |
| `make dc-ps` | Show Docker Compose service status. |
| `make dc-shell` | Open a shell in the running dev container as the `codex` user. |
| `make dc-exec CMD="..."` | Run an arbitrary command inside the dev container as the `codex` user. |

Useful examples:

```sh
make dc-ps
make dc-shell
make dc-exec CMD="go test ./internal/config"
```

## Runtime Targets

Use `run-*` targets for processes and commands that run inside an already-running dev container.

| Target | Purpose |
|---|---|
| `make run-all` | Start issue-tracker, orchestrator, and Web inside the running dev container. |
| `make run-stop` | Stop Air-managed service processes inside the dev container without stopping the container. |
| `make run-issue-tracker` | Start only the issue-tracker process inside the running dev container. |
| `make run-is` | Alias for `run-issue-tracker`. |
| `make run-orchestrator` | Start issue-tracker, then start the orchestrator process. |
| `make run-or` | Alias for `run-orchestrator`. |
| `make run-web` | Start issue-tracker, then start the Web process. |
| `make run-w` | Alias for `run-web`. |
| `make run-tui` | Run the TUI interactively inside the dev container. |
| `make run-tq ARGS="..."` | Run the installed `tq $(ARGS)` binary inside the running dev container without changing service processes. |
| `make run-ps` | Show dev processes running inside the dev container. |
| `make run-logs` | Follow `$TQ_HOME/system/log/*.log`. |

Useful examples:

```sh
make run-is
make run-or
make run-w
make run-tq ARGS="issue list"
make run-logs
```

## Verification Targets

| Target | Purpose |
|---|---|
| `make dev-test` | Run `go test ./...`, install Web dependencies, and run Web typecheck inside the dev container. |
| `make dev-build` | Run `go test ./...`, install Web dependencies, and run the Web production build inside the dev container. |

## Release Targets

| Target | Purpose |
|---|---|
| `make prerelease` | Create and push a prerelease tag through `scripts/release.sh`. |
| `make release version=v0.1.1` | Create and push a formal release tag through `scripts/release.sh`. |
| `make install-tq` | Install `tq` from the latest formal release into `$HOME/.local/bin`. |
| `make install-tq version=v0.1.0` | Install `tq` from a specific release tag. |
| `make install-tq-prerelease` | Install `tq` from the latest prerelease. |
| `make install-tq-prerelease version=v0.1.0-pre.1` | Install `tq` from a specific prerelease tag. |

See [Deployment Flow](../design/deployment.md) for the full tag, GitHub Actions, and GoReleaser flow.

## Authentication Targets

| Target | Purpose |
|---|---|
| `make dev-codex-login` | Run `codex login --device-auth` inside the dev container and persist credentials in the `codex-home` Docker volume. |
| `make dev-codex-status` | Show Codex authentication status inside the dev container. |
| `make dev-gh-login` | Run `gh auth login` and `gh auth setup-git` inside the dev container, then persist credentials in the `gh-config` Docker volume. |
| `make dev-gh-status` | Show GitHub CLI authentication status inside the dev container. |

Use device auth for container logins when a browser redirect points to a localhost callback inside
the container that is not reachable from the host browser.

Authentication targets only execute commands in an existing `dev` container. They do not build or
start the container; run `make dev-up` first when the dev container does not exist.

Examples:

```sh
make dev-codex-login
make dev-codex-status
make dev-gh-login
make dev-gh-status
```

## Operational Notes

The `dev` container is long-lived. Service processes are ordinary child processes launched with `docker compose exec`; they are not separate Compose services. Use `make run-stop` to stop only those child processes, or `make dev-down` to stop the Compose services.

The Makefile runs container commands as the `codex` user. At container startup, Compose prepares
writable named volumes for Go module cache, Go build cache, Web `node_modules` under
`cmd/web/frontend`, Codex credentials, and GitHub CLI credentials.

The Go Web server proxies browser API calls to the issue-tracker and orchestrator over container-local URLs. Restart the Web process with `make run-web` after changing frontend code, proxy configuration, or backend ports.
