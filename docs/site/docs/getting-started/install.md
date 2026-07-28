---
id: install
title: Installation
sidebar_position: 2
---

# Installation

Install Tasq and start its local services before following the [Agent Tutorial](pathname:///getting-started/agent-tutorial).

## Install the CLI

Choose `TQ_HOME` before installing. The installer puts the public `tq` command
on your `PATH` and installs the private service binaries under the selected
home.

```sh
curl -fsSLO https://raw.githubusercontent.com/version-1/tasq/main/scripts/install.sh
less install.sh
export TQ_HOME="${HOME}/.tasq"
TQ_HOME="$TQ_HOME" sh install.sh
export PATH="${HOME}/.local/bin:${PATH}"
tq version
```

Review the installer before running it. It installs `tq` in `~/.local/bin` by
default, and installs `issue-tracker`, `orchestrator`, and `web` in
`$TQ_HOME/system/bin`. The service binaries are managed by `tq service start`;
running them directly is not a supported distribution interface.

## Start Local Services

Tasq stores machine-local runtime data and service binaries under `TQ_HOME`. If
`TQ_HOME` was not set during installation, it uses `~/.tasq`. Use the same home
when starting services.

```sh
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

## Update Tasq

Check the installed version, update to the latest formal release, and verify
the new version:

```sh
tq version
tq update
tq version
```

:::warning Service interruption and automatic migrations

`tq update` temporarily stops the local services, installs the release,
applies database migrations, and restarts the services. Do not start an update
while uninterrupted local service access is required.

:::

See [Update Tasq](pathname:///guides/update-tasq) for confirmation-free updates
and installing a specific formal or prerelease version.

## Continue with the Tutorial

Your CLI and local services are ready. Continue to the [Agent Tutorial](pathname:///getting-started/agent-tutorial) to register a project, prepare an agent workflow, and follow an issue through to a pull request.
