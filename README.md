# tasq

Local-first issue tracker and task orchestrator for managing executable work, assigning it to coding agents, and observing progress from a Web UI or TUI.

## Components

- Issue Tracker: Go REST API backed by SQLite. It owns issues, comments, projects, workspaces, and UI summaries.
- Orchestrator: Go service backed by SQLite. It records run state and runner events for runtime inspection.
- `tq`: Go CLI for agents and workflow tools to create, read, list, and update issues through the issue-tracker API.
- Web UI: Go-served Vite + React client for the issue-tracker API.
- TUI: Go terminal client for the same issue-tracker API.

For the full architecture, see [docs/design.md](docs/design.md).
For local configuration, see [docs/design/configuration.md](docs/design/configuration.md).

## Quick Start

Use Docker Compose through `make` for local development:

```sh
make dev-up
```

This starts the `dev` container and OpenAPI UI, then launches the issue-tracker, orchestrator, and Web UI inside the `dev` container. Docker Compose assigns host ports automatically, and the command prints the assigned URLs.

Show the URLs again:

```sh
make dev-ports
```

Stop the environment:

```sh
make dev-down
```

List all available development commands:

```sh
make help
```

### Linux/WSL2 Sandbox Prerequisite

Codex uses Bubblewrap for Linux sandboxing. The dev image installs `bubblewrap`,
but Linux and WSL2 hosts must also allow unprivileged user namespace creation for
Codex sandboxed commands to work reliably. If Codex reports `bwrap: No
permissions to create a new namespace`, treat it as a host or Docker runtime
capability issue, not just a missing package in the image.

## Verification

Run the standard Compose-backed checks:

```sh
make dev-test
```

Run the broader build check before handing off changes that affect both Go services and the Web UI:

```sh
make dev-build
```

## Documentation

- [docs/development.md](docs/development.md): repository workflow, task flow, documentation update rules, and component workflow links.
- [WORKFLOW.md](WORKFLOW.md): Symphony runtime workflow contract used by the orchestrator.
- [docs/design.md](docs/design.md): system architecture and service boundaries.
- [docs/design/deployment.md](docs/design/deployment.md): release tag creation, GitHub Actions, and GoReleaser deployment flow.
- [docs/references/makefile.md](docs/references/makefile.md): Makefile targets, variables, and local development command reference.
- [cmd/issue-tracker/WORKFLOW.md](cmd/issue-tracker/WORKFLOW.md): issue-tracker development workflow.
- [cmd/orchestrator/WORKFLOW.md](cmd/orchestrator/WORKFLOW.md): orchestrator development workflow.
- [cmd/web/WORKFLOW.md](cmd/web/WORKFLOW.md): Web UI development workflow.
- [docs/design/web.md](docs/design/web.md): Web UI structure and styling conventions.
- [docs/openapi/issue-tracker.yml](docs/openapi/issue-tracker.yml): issue-tracker OpenAPI contract.
- [docs/symphony/README.md](docs/symphony/README.md): Symphony documentation index.
- [docs/symphony/SPEC.md](docs/symphony/SPEC.md): Symphony orchestration and runner specification.
- [docs/symphony/DEVIATIONS.md](docs/symphony/DEVIATIONS.md): intentional deviations from the Symphony specification.

Japanese counterpart: [README.ja.md](README.ja.md).

## Notes

- Runtime state and SQLite files are created under `.tasq/` in the repository and are ignored by git.
- Compose stores Go module/build caches, `cmd/web/frontend/node_modules`, Codex login state, and GitHub CLI login state in named Docker volumes.
- The orchestrator reads `WORKFLOW.md` for Symphony-oriented runtime settings and the per-issue agent prompt.
- The Web UI calls local backends through the Go server proxy paths `/tracker/*` and `/orchestrator/*`.
- `tq` resolves the issue-tracker API URL from `$TQ_HOME/system/state.json` when run through `make run-tq`.
- Run `make dev-codex-login` once to authenticate Codex with device auth and persist credentials in the `codex-home` Docker volume.
- Run `make dev-gh-login` once to authenticate GitHub CLI, configure Git to use `gh` as its HTTPS credential helper, and persist credentials in the `gh-config` Docker volume. Use an HTTPS Git remote for pushes from the dev container.
- Use `make dev-codex-status` and `make dev-gh-status` to confirm the dev container is authenticated before running agent workflows that need Codex or GitHub access.

## tq CLI

List issues:

```sh
make run-tq ARGS="issue list"
```

Get an issue:

```sh
make run-tq ARGS="issue get 1"
```

Create an issue:

```sh
make run-tq ARGS='issue create --title "Wire Symphony workflow" --description "Define the first workflow contract" --status ready --priority high'
```

Use issue shortcuts for common status and text updates:

```sh
make run-tq ARGS="issue ready 1"
make run-tq ARGS="issue close 1"
make run-tq ARGS='issue rename 1 "Clarify workflow contract"'
make run-tq ARGS='issue edit 1 "Updated description"'
```

Use `--output json` for machine-readable output:

```sh
make run-tq ARGS="--output json issue list"
```
