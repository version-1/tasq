---
id: getting-started
title: Getting Started
sidebar_position: 1
---

# Getting Started

This guide takes you from a fresh checkout to a running local Tasq service and your first issue.

## Install

Clone the repository and build the `tq` command:

```sh
git clone https://github.com/version-1/tasq.git
cd tasq
go build -o ./bin/tq ./cmd/tq
```

Add the binary to your `PATH` or call it through `./bin/tq`.

```sh
export PATH="$PWD/bin:$PATH"
tq version
```

## Initialize Local State

Tasq stores local service data under `$TQ_HOME`. If `TQ_HOME` is not set, Tasq uses the default home directory for the current user.

Apply database migrations before starting services:

```sh
tq migrate
```

Start the issue-tracker, orchestrator, and Web UI services:

```sh
tq service start
tq service status
```

Register the current repository as a Tasq project. The project key is inferred from the directory name unless you pass `--key`.

```sh
tq project add --key tasq .
tq project check tasq
```

## Create Your First Issue

Create an issue in the project:

```sh
tq issue create \
  --project tasq \
  --title "Write onboarding notes" \
  --description "Capture the first setup path for new contributors."
```

List issues:

```sh
tq issue list --project tasq
```

Inspect an issue:

```sh
tq issue get 1
```

Move the issue through the workflow:

```sh
tq issue ready 1
tq issue update 1 --status in_progress
tq comment add 1 --type progress --body "Started work."
tq issue update 1 --status review
```

Open the Web UI when the local service is running:

```sh
tq web
```

## Useful Defaults

- `tq` resolves the issue-tracker URL from `--api-url`, `TQ_API_URL`, `$TQ_HOME/system/state.json`, or `http://localhost:37651`.
- Use `--output json` when scripts or agents need structured output.
- Service logs are written under `$TQ_HOME/system/log/` and can be viewed with `tq logs`.
