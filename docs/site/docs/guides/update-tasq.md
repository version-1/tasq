---
id: update-tasq
title: Update Tasq
sidebar_position: 5
---

# Update Tasq

Use `tq update` to replace the installed Tasq CLI and local service binaries
with GitHub Release artifacts. By default, the command installs the latest
formal release.

## Update to the Latest Formal Release

Check the current version, run the update, and confirm the installed version:

```sh
tq version
tq update
tq version
```

Before making changes, `tq update` displays the current and target versions and
asks for confirmation.

:::warning Service interruption and automatic migrations

The update temporarily stops the issue-tracker, orchestrator, and Web services.
It installs the release artifacts, verifies the newly installed `tq` binary,
applies database migrations, and then restarts the services.

Run the command only when a short interruption to the local services is
acceptable. The `-y` option skips the confirmation for this interruption and
the automatic migration.

:::

## Skip Confirmation

Pass `-y` to run the same update without the confirmation prompt:

```sh
tq update -y
```

This option changes only the confirmation behavior. The command still stops
and restarts the services and applies database migrations.

## Install a Specific Version

Pass `--tag` with a formal release tag to install that version:

```sh
tq update --tag v0.3.5
```

Prerelease tags are also supported:

```sh
tq update --tag v0.4.0-rc.1
```

After either command completes, verify the installed version:

```sh
tq version
```

For the complete command synopsis, see the
[CLI Reference](pathname:///reference/cli-reference#runtime-commands).
