---
id: install
title: Installation
sidebar_position: 2
---

# Installation

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

## Install a Prerelease

After downloading the installer, install the latest prerelease with:

```sh
TASQ_RELEASE_CHANNEL=prerelease sh install.sh
```

This requires the GitHub CLI (`gh`) to select the latest prerelease. To install
a specific formal release or prerelease tag, set `TASQ_VERSION`; this does not
require `gh`:

```sh
TASQ_VERSION=v0.4.0-rc.1 sh install.sh
```

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
