# tasq

AI coding agent task manager.

tasq helps you turn implementation work into a visible queue, start local services for that queue, and inspect progress from both the `tq` CLI and the Web UI.

[![Watch the Tasq introduction video](docs/site/static/img/tasq-thumbnail.png)](https://github.com/user-attachments/assets/8c4fdc9c-c70b-4f86-8e0a-323f8880ffb7)

Japanese counterpart: [README.ja.md](README.ja.md).

## Why Tasq?

Parallel AI-agent work shifts the bottleneck from writing code to coordinating tasks, workspaces, and reviews. Tasq keeps that work visible through an issue queue, isolated workspaces, local services, and both a CLI and Web UI.

For the problem, workflow, and complete product overview, see the [Tasq documentation site](https://version-1.github.io/tasq/).

## Features

- Issue queue for agent-sized tasks, priorities, dependencies, and comments.
- `tq` CLI for creating tasks, updating state, adding progress comments, and scripting workflows.
- Local issue-tracker, orchestrator, and Web UI services backed by SQLite.
- Web UI for scanning projects, issues, comments, queue status, and service state.
- Project registration so issues stay tied to a real local repository path.
- Release archives that include the runtime binaries needed for a binary-only local setup.

## Install

Install the latest formal release archive with the reviewed installer:

```sh
curl -fsSLO https://raw.githubusercontent.com/version-1/tasq/main/scripts/install.sh
less install.sh
sh install.sh
```

For installation directories, `TQ_HOME`, service startup, and updates, see [Installation](https://version-1.github.io/tasq/getting-started/install).

## Documentation

Visit the [Tasq documentation site](https://version-1.github.io/tasq/) for the tutorial, guides, concepts, and reference.
