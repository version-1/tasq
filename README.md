# tasq

Local-first issue tracker and task orchestrator for managing executable work, assigning it to coding agents, and observing progress from a Web UI or TUI.

## Components

- Issue Tracker: Go REST API backed by SQLite. It owns issues, comments, projects, workspaces, and UI summaries.
- Orchestrator: Go service backed by SQLite. It records run state and runner events for runtime inspection.
- `tq`: Go CLI for agents and workflow tools to create, read, list, and update issues through the issue-tracker API.
- Web UI: Next.js client for the issue-tracker API.
- TUI: Go terminal client for the same issue-tracker API.

For the full architecture, see [docs/design.md](docs/design.md).
For local configuration, see [docs/configuration.md](docs/configuration.md).

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

- [WORKFLOW.md](WORKFLOW.md): repository workflow, task flow, documentation update rules, and component workflow links.
- [docs/design.md](docs/design.md): system architecture and service boundaries.
- [docs/references/makefile.md](docs/references/makefile.md): Makefile targets, variables, and local development command reference.
- [cmd/issue-tracker/WORKFLOW.md](cmd/issue-tracker/WORKFLOW.md): issue-tracker development workflow.
- [cmd/orchestrator/WORKFLOW.md](cmd/orchestrator/WORKFLOW.md): orchestrator development workflow.
- [web/WORKFLOW.md](web/WORKFLOW.md): Web UI development workflow.
- [web/docs/design.md](web/docs/design.md): Web UI structure and styling conventions.
- [docs/openapi/issue-tracker.yml](docs/openapi/issue-tracker.yml): issue-tracker OpenAPI contract.
- [docs/symphony/README.md](docs/symphony/README.md): Symphony documentation index.
- [docs/symphony/SPEC.md](docs/symphony/SPEC.md): Symphony orchestration and runner specification.
- [docs/symphony/DEVIATIONS.md](docs/symphony/DEVIATIONS.md): intentional deviations from the Symphony specification.

Japanese counterpart: [README.ja.md](README.ja.md).

## Notes

- Runtime state and SQLite files are created under `.tasq/` in the repository and are ignored by git.
- Compose stores Go module/build caches, `web/node_modules`, and Codex login state in named Docker volumes.
- The orchestrator reads `WORKFLOW.md` for Symphony-oriented runtime settings.
- The Web UI calls the issue-tracker API through `NEXT_PUBLIC_ISSUE_TRACKER_URL` when served from a different origin.
- `tq` resolves the issue-tracker API URL from `$TQ_HOME/system/state.json` when run through `make run-tq`.
- Run `make dev-codex-login` once to authenticate Codex with device auth and persist credentials in the `codex-home` Docker volume.

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
