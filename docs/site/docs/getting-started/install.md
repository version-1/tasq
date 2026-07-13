---
id: install
title: Install
sidebar_position: 2
---

# Install

Install Tasq and start its local services before following the [Agent Tutorial](pathname:///getting-started/agent-tutorial).

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

## Continue with the Tutorial

Your CLI and local services are ready. Continue to the [Agent Tutorial](pathname:///getting-started/agent-tutorial) to register a project, prepare an agent workflow, and follow an issue through to a pull request.
