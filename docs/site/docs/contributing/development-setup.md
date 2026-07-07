---
id: development-setup
title: Development Setup
sidebar_position: 1
---

# Development Setup

Tasq development can run in Docker Compose or directly on the host. Prefer
Compose when working on service integration, Web UI behavior, Codex/GitHub CLI
authentication, or anything that should match the shared project workflow. Use
host-only mode for quick CLI checks, release binary smoke tests, or debugging a
local installed `tq`.

For the full repository workflow, see
[docs/development.md](https://github.com/version-1/tasq/blob/main/docs/development.md).

## Compose Environment

The Compose workflow uses one long-lived `dev` container. `make dev-up` builds
the container, starts issue-tracker, orchestrator, and Web processes inside it,
starts the OpenAPI UI, and prints the assigned local URLs.

Docker Compose assigns host ports by default, so do not assume fixed ports in
Compose mode. Run `make dev-ports` whenever you need the current URLs.

```sh
make dev-up
make dev-ports
```

Common commands:

- `make dev-open` opens the Web UI and OpenAPI UI in a browser.
- `make run-tq ARGS="issue list"` runs `tq` inside the dev container.
- `make run-issue-tracker`
- `make run-orchestrator`
- `make run-web`
- `make run-migrate`
- `make run-logs` follows service logs.
- `make dev-down` stops Compose services.

## Host-Only Environment

For host-only local operation, build `tq`, put the local binary first on
`PATH`, set a repository-local `TQ_HOME`, apply migrations, and start services.

```sh
make build-tq
export PATH="$PWD/bin:$PATH"
export TQ_HOME="$PWD/.tasq"
tq migrate
tq service start
```

Host services use fixed loopback ports and write discovery state under
`$TQ_HOME/system/state.json`.

| Service | Port |
| --- | ---: |
| issue-tracker | `37651` |
| orchestrator | `37652` |
| web | `37653` |

Use host-only mode carefully when Compose services are already running. Stop one
mode before starting the other if ports or runtime state conflict.

## Codex in the Dev Container

The dev container uses `CODEX_HOME=/home/codex/.codex`, backed by the
`codex-home` named volume. Run `make dev-codex-login` once to authenticate with
device auth inside the container, then verify it with `make dev-codex-status`.

GitHub CLI credentials are stored in the `gh-config` named volume. Run
`make dev-gh-login` once before workflows that create branches, push, or open
pull requests from inside the dev container. Verify with `make dev-gh-status`.

On Linux and WSL2 hosts, Codex sandboxed commands need unprivileged user
namespaces. If Codex reports a Bubblewrap namespace error, treat it as a host or
Docker runtime capability issue.

## Documentation Work

The docs site lives under `docs/site`. Use the repository wrappers when you want
dependencies installed automatically:

```sh
make dev-docs
make dev-docs-ja
make dev-docs-build
```

When editing documentation, keep the English and Japanese pages synchronized.
English docs are the primary source of truth for development and design
decisions.
