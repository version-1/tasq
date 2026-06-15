# tasq

AI coding agent task manager.

tasq is a CLI tool for running multiple AI coding-agent tasks in parallel with Claude Code, Codex, and similar agents.

It creates an isolated workspace for each task and supports the workflow from task registration, agent execution, state management, review, and integration.

## Problem

AI coding agents make it possible to work on multiple implementation tasks at the same time. The bottleneck moves from writing code to managing parallel work.

### Human context switching

Agents can run in parallel, but humans still need to track the state of each task:

- Which tasks were assigned.
- Which agents are running.
- How far each task has progressed.
- What should be reviewed next.

This increases coordination cost and makes it harder to keep a reliable overview of active work.

### Workspace conflicts

Running multiple agents in one repository checkout can create conflicts:

- Branch switching interrupts other work.
- Unfinished changes overlap.
- Multiple agents may edit the same files from the same workspace.

### Repeated setup work

Each agent task often needs the same preparation steps:

- Create a branch.
- Create a worktree.
- Install or verify dependencies.
- Run the right setup and verification commands.

Repeating that setup for every task slows down parallel execution.

## Solution

tasq manages executable tasks as a queue and creates agent-ready workspaces for tasks that are ready to run.

![Tasq task queue to parallel agent workspaces](docs/site/static/img/agent-task-queue.svg)

The goal is not just faster code generation. The goal is to reduce the management cost introduced by parallel agent work.

## Features

### Task queue

Track implementation tasks as they move through the workflow.

```sh
tasq add "implement user login"
```

```text
TODO
READY
RUNNING
DONE
```

### Isolated workspace

Create a Git worktree for each task so agents can work independently.

```text
project/
├── main
└── .worktrees/
    ├── task-a
    ├── task-b
    └── task-c
```

### Parallel agent execution

Run multiple agents at the same time without sharing one mutable checkout.

```text
task-a -> Codex A
task-b -> Codex B
task-c -> Codex C
```

### Review workflow

Review and integrate agent output per task.

```text
RUNNING
   |
   v
REVIEW
   |
   v
MERGED
```

tasq provides task management, workspace isolation, and agent execution support for AI coding-agent workflows.

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
- The orchestrator resolves each project's `WORKFLOW.md` for Symphony-oriented runtime settings and the per-issue agent prompt, with `$TQ_HOME/WORKFLOW.md` as the fallback.
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
