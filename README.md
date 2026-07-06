# tasq

AI coding agent task manager.

tasq helps you turn implementation work into a visible queue, start local services for that queue, and inspect progress from both the `tq` CLI and the Web UI.

Japanese counterpart: [README.ja.md](README.ja.md).

## Problem

AI coding agents make it possible to work on several implementation tasks at the same time. The bottleneck moves from writing code to coordinating parallel work.

Teams still need to know which tasks exist, which ones are ready, what is running, what needs review, and which local workspace belongs to which task. Running every agent in one checkout also increases branch-switching and file-conflict risk.

## Solution

tasq gives agent work a product surface: an issue tracker, local services, a CLI, and a Web UI that keep task state and project context in one place.

![Tasq task queue to parallel agent workspaces](docs/site/static/img/agent-task-queue.svg)

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
version="$(curl -fsSLI -o /dev/null -w '%{url_effective}' https://github.com/version-1/tasq/releases/latest | sed 's#.*/##')"
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in x86_64) arch="amd64" ;; aarch64) arch="arm64" ;; esac
archive="tasq_${version#v}_${os}_${arch}.tar.gz"
tmp_dir="$(mktemp -d)"
curl -fsSL "https://github.com/version-1/tasq/releases/download/${version}/${archive}" -o "${tmp_dir}/${archive}"
tar -xzf "${tmp_dir}/${archive}" -C "${tmp_dir}"
install_dir="${HOME}/.local/bin"
mkdir -p "$install_dir"
cp "${tmp_dir}/tq" "${tmp_dir}/issue-tracker" "${tmp_dir}/orchestrator" "${tmp_dir}/web" "$install_dir/"
chmod 0755 "$install_dir/tq" "$install_dir/issue-tracker" "$install_dir/orchestrator" "$install_dir/web"
```

Make sure the install directory is on your `PATH`:

```sh
export PATH="${HOME}/.local/bin:${PATH}"
tq version
```

## Getting Started

Tasq stores machine-local runtime data under `TQ_HOME`. If `TQ_HOME` is not set, it uses `~/.tasq`.

Initialize local databases and start the issue-tracker, orchestrator, and Web UI with explicit local ports:

```sh
export TQ_HOME="${HOME}/.tasq"
export TQ_API_URL="http://127.0.0.1:47651"
tq migrate
issue-tracker -addr 127.0.0.1:47651 &
tracker_pid=$!
orchestrator -issue-tracker http://127.0.0.1:47651 -port 47652 &
orchestrator_pid=$!
web -addr 127.0.0.1:47653 \
  -tracker-url http://127.0.0.1:47651 \
  -orchestrator-url http://127.0.0.1:47652 &
web_pid=$!
sleep 1
```

This starts the local services on:

| Service | Port |
| --- | ---: |
| issue-tracker | `47651` |
| orchestrator | `47652` |
| web | `47653` |

Register a local repository as a project:

```sh
tq project add --key tasq-demo .
```

Create a task and move it into the queue:

```sh
tq issue create \
  --project tasq-demo \
  --title "Try Tasq from binaries" \
  --description "Create the first issue through the tq CLI."
tq issue list --project tasq-demo
```

Open the Web UI at [http://127.0.0.1:47653](http://127.0.0.1:47653) and confirm that the project and issue are visible.

Stop the local services when you are done:

```sh
kill "$web_pid" "$orchestrator_pid" "$tracker_pid"
```

If one of these ports is already in use, choose another loopback port and update `TQ_API_URL`, `-issue-tracker`, `-tracker-url`, and `-orchestrator-url` consistently.

## Documentation

- [Design documentation](docs/design.md)
- [Release binary startup notes](docs/design/release-binary-startup.md)
- [CLI reference](docs/site/docs/reference/cli-reference.md)

## Development

For repository workflow, local development, and verification, see [docs/development.md](docs/development.md).
