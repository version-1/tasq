---
id: development-setup
title: Development Setup
sidebar_position: 1
---

# Development Setup

Tasq development can run in Docker Compose or directly on the host. Compose keeps services in one long-lived `dev` container, while host-only operation uses `tq service start`.

## Compose Environment

The Compose workflow runs the issue-tracker, orchestrator, and Web server inside the `dev` container. Runtime state defaults to `/workspace/.tasq` in the container.

```sh
make dev-up
make dev-ports
```

Recommended development commands include:

- `make run-issue-tracker`
- `make run-orchestrator`
- `make run-web`
- `make run-tui`
- `make run-migrate`
- `make dev-codex-login`

## Host-Only Environment

For host-only local operation, build `tq`, set `TQ_HOME`, apply migrations, and start services.

```sh
go build -o ./bin/tq ./cmd/tq
export PATH="$PWD/bin:$PATH"
export TQ_HOME="$PWD/.tasq"
tq migrate
tq service start
```

The services use loopback ports and write discovery state under `$TQ_HOME/system/state.json`.

## Codex in the Dev Container

The dev container uses `CODEX_HOME=/home/codex/.codex`, backed by the `codex-home` named volume. Run `make dev-codex-login` once to authenticate with device auth inside the container.
