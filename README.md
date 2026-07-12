# tasq

AI coding agent task manager.

tasq helps you turn implementation work into a visible queue, start local services for that queue, and inspect progress from both the `tq` CLI and the Web UI.

[![Watch the Tasq introduction video](docs/site/static/img/tasq-thumbnail.png)](https://github.com/user-attachments/assets/8c4fdc9c-c70b-4f86-8e0a-323f8880ffb7)

[Watch the Tasq introduction video](https://github.com/user-attachments/assets/8c4fdc9c-c70b-4f86-8e0a-323f8880ffb7).

Japanese counterpart: [README.ja.md](README.ja.md).

## Problem

AI coding agents make it possible to work on multiple implementation tasks at the same time. The bottleneck moves from writing code to managing parallel work.

### Human Context Switching

Agents can run in parallel, but humans still need to track which tasks were assigned, which agents are running, how far each task has progressed, and what should be reviewed next.

### Workspace Conflicts

Running multiple agents in one repository checkout can create branch switching issues, unfinished-change conflicts, and overlapping file edits.

### Repeated Setup Work

Each agent task often needs the same preparation steps: create a branch, create a worktree, verify dependencies, and run the right setup commands.

## Solution

tasq gives agent work a product surface: an issue tracker, local services, a CLI, and a Web UI that keep task state and project context in one place.

![Tasq agent issue flow](docs/site/static/img/agent-issue-flow.png)

Tasks move through a reviewable workflow:

```text
backlog -> ready -> in_progress -> review -> done
```

## Features

- Issue queue for agent-sized tasks, priorities, dependencies, and comments.
- `tq` CLI for creating tasks, updating state, adding progress comments, and scripting workflows.
- Local issue-tracker, orchestrator, and Web UI services backed by SQLite.
- Web UI for scanning projects, issues, comments, queue status, and service state.
- Project registration so issues stay tied to a real local repository path.
- Release archives that include the runtime binaries needed for a binary-only local setup.

## Install

Download the latest GitHub Release archive for your platform, extract it, and place all four binaries on your `PATH`.

Each release tarball contains:

- `tq`: the CLI you run directly.
- `issue-tracker`: the local REST API service.
- `orchestrator`: the local run-state service.
- `web`: the local Web UI server with embedded frontend assets.

Install the latest formal release:

```sh
curl -fsSLO https://raw.githubusercontent.com/version-1/tasq/main/scripts/install.sh
less install.sh
sh install.sh
```

Review the installer before running it. The installer verifies the downloaded release archive against the release `checksums.txt`, then verifies the installed `tq` binary matches the extracted release binary. A successful install prints `verified installed tq sha256: ...`.

Make sure the install directory is on your `PATH`:

```sh
export PATH="${HOME}/.local/bin:${PATH}"
tq version
```

## Getting Started

For setup and usage instructions, see the [Tasq documentation site](https://version-1.github.io/tasq/).

## Documentation

For guides, concepts, and reference material, visit the [Tasq documentation site](https://version-1.github.io/tasq/).
