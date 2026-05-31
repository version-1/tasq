# tasq

Local-first issue tracker and task orchestrator for managing executable work, assigning it to coding agents, and observing progress from a Web UI or TUI.

## Components

- Issue Tracker: Go REST API backed by SQLite. It owns issues, work items, and UI summaries.
- Orchestrator: Go worker backed by SQLite. It claims executable work and records run state.
- `tq`: Go CLI for agents and workflow tools to create, read, list, and update issues through the issue-tracker API.
- Web UI: Next.js client for the issue-tracker API.
- TUI: Go terminal client for the same issue-tracker API.

For the full architecture, see [docs/design.md](docs/design.md).

## Quick Start

Use Docker Compose through `make` for local development:

```sh
make web-up
```

This starts the issue-tracker, orchestrator, OpenAPI UI, and Web UI. Docker Compose assigns host ports automatically, and the command prints the assigned URLs.

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
make dev-build-app
```

## Documentation

- [WORKFLOW.md](WORKFLOW.md): repository workflow, task flow, documentation update rules, and component workflow links.
- [docs/design.md](docs/design.md): system architecture and service boundaries.
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

- SQLite files are created under `.data/` in the repository and are ignored by git.
- Compose stores Go module/build caches and `web/node_modules` in named Docker volumes.
- The orchestrator reads `WORKFLOW.md` for Symphony-oriented runtime settings.
- The Web UI calls the issue-tracker API through `NEXT_PUBLIC_ISSUE_TRACKER_URL` when served from a different origin.
- `tq` resolves the issue-tracker API URL from `--api-url`, `TQ_API_URL`, or `http://localhost:8080`.

## tq CLI

List issues:

```sh
go run ./cmd/tq --api-url http://localhost:8080 issue list
```

Get an issue through the default `TQ_API_URL`:

```sh
TQ_API_URL=http://localhost:8080 go run ./cmd/tq issue get 1
```

Create an issue:

```sh
go run ./cmd/tq issue create \
  --api-url http://localhost:8080 \
  --title "Wire Symphony workflow" \
  --description "Define the first workflow contract" \
  --status ready \
  --priority high
```

Use `--output json` for machine-readable output:

```sh
go run ./cmd/tq --api-url http://localhost:8080 --output json issue list
```
