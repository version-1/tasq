---
id: quickstart
title: QuickStart
sidebar_position: 2
---

# QuickStart

This path starts Tasq from a fresh checkout and creates the first issue.

## Build the CLI

```sh
git clone https://github.com/version-1/tasq.git
cd tasq
make build-tq
export PATH="$PWD/bin:$PATH"
tq version
```

## Initialize Local State

Tasq stores local state under `TQ_HOME`. If it is not set, Tasq uses `~/.tasq`.

```sh
export TQ_HOME="$PWD/.tasq"
tq migrate
tq service start
tq service status
```

`tq service start` launches the issue-tracker, orchestrator, and Web server on local loopback ports and writes discovery state to `$TQ_HOME/system/state.json`.

## Register a Project

Register the current repository so issues can be scoped to a project.

```sh
tq project add --key tasq .
tq project check tasq
```

## Create and Move an Issue

```sh
tq issue create \
  --project tasq \
  --title "Write onboarding notes" \
  --description "Capture the first setup path for new contributors."
```

```sh
tq issue list --project tasq
tq issue get 1
tq issue ready 1
tq issue update 1 --status in_progress
tq comment add 1 --type progress --body "Started work."
tq issue update 1 --status review
```

## Open the Web UI

```sh
tq web
```

The Web UI expects the local service to already be running. If it cannot open, check `tq service status` and inspect logs with `tq logs web`.
