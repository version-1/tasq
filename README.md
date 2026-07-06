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

## Developer Documentation

For local development, verification, command references, and repository workflow, see [docs/development.md](docs/development.md).

Japanese counterpart: [README.ja.md](README.ja.md).
