---
id: quickstart
title: QuickStart
sidebar_position: 2
---

# QuickStart

This path installs Tasq, starts the local services, registers one project, and
creates the first issue.

Use [Setup Guide](pathname:///getting-started/setup-guide) instead when you are
preparing Codex permissions or a repeatable agent development environment.

## Install the CLI

Install the latest formal release and make sure `tq` is on your `PATH`.

```sh
curl -fsSLO https://raw.githubusercontent.com/version-1/tasq/main/scripts/install.sh
less install.sh
sh install.sh
export PATH="${HOME}/.local/bin:${PATH}"
tq version
```

Review the installer before running it. The release archive installs the `tq`
CLI plus the local `issue-tracker`, `orchestrator`, and `web` service binaries.

## Start Local Services

Tasq stores machine-local runtime data under `TQ_HOME`. If `TQ_HOME` is not set,
Tasq uses `~/.tasq`.

```sh
export TQ_HOME="${HOME}/.tasq"
tq migrate
tq service start
tq service status
```

`tq service start` launches the issue-tracker, orchestrator, and Web server on
fixed local loopback ports and writes discovery state to
`$TQ_HOME/system/state.json`.

| Service | Port |
| --- | ---: |
| issue-tracker | `37651` |
| orchestrator | `37652` |
| web | `37653` |

## Register a Project

Register a local repository so issues can be scoped to a project. From inside
the repository you want to track, run:

```sh
tq project add --key tasq-demo .
tq project check tasq-demo
```

## Create the First Issue

```sh
tq issue create \
  --project tasq-demo \
  --title "Write onboarding notes" \
  --description "Capture the first setup path for new contributors."
```

Confirm that the issue exists, then move it into the queue.

```sh
tq issue list --project tasq-demo
tq issue ready 1
tq issue update 1 --status in_progress
tq comment add 1 --type progress --body "Started work."
tq issue update 1 --status review
```

## Open the Web UI

```sh
tq web
```

The Web UI expects the local services to already be running. If it cannot open,
check service state and logs:

```sh
tq service status
tq logs web
```

## Stop or Restart Services

Use `service stop` when you are done. Run `service start` again to restart the
same local environment.

```sh
tq service stop
tq service start
```
