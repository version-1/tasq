---
id: setup-guide
title: Setup Guide
sidebar_position: 3
---

# Setup Guide

This guide covers the minimum local setup outside `tq` that Tasq expects before
you run Codex-backed agent work repeatedly.

Keep these settings scoped to trusted local projects. Avoid granting broad global permissions for commands that can change Git history, remote state, credentials, or system configuration.

## Codex Authentication

Codex authentication is a prerequisite for Tasq agent runs that use Codex. This
guide assumes the Codex CLI is already authenticated with the account and
workspace you intend to use.

For login methods, device authentication, API-key authentication, credential
storage, and CI/CD authentication, see the official
[Codex Authentication](https://developers.openai.com/codex/auth) documentation.

## Agent-Friendly `tq` Usage

The `tq` command is often used by agents such as Codex and Claude Code rather
than by a human typing every command directly. Install Tasq so the `tq` command
is available on `PATH`, then install or expose the `tasq-cli` agent guide in the
agent environment.

With `tasq-cli` available, agents have a compact reference for common `tq`
commands, issue lookup, comments, status transitions, and service operations.
That reduces ambiguity when an agent needs to inspect a task, update progress,
or add a note before handing work back.

For the shortest Tasq install path, see [QuickStart](./quickstart). For command
behavior and API resolution details, see [tq CLI](./concepts/tq-cli).

## Minimal Codex Permissions

Detailed autonomy setup belongs in
[Codex Autonomy Setup](../guides/codex-autonomy-setup). For this setup guide,
start with the smallest `~/.codex/config.toml` entry that trusts the Tasq
checkout and grants workspace-write access to the checkout plus the repository
Git metadata.

```toml
# ~/.codex/config.toml

default_permissions = "tasq_workspace"

[projects."/Users/YOU/src/tasq"]
trust_level = "trusted"

[permissions.tasq_workspace]
description = "Tasq checkout with Git metadata writes enabled."
extends = ":workspace"

[permissions.tasq_workspace.filesystem.":workspace_roots"]
"." = "write"

[permissions.tasq_workspace.workspace_roots]
"/Users/YOU/src/tasq" = true
"/Users/YOU/src/tasq/.git" = true
```

Use exact absolute paths for your local checkout. If you run agents from
additional worktrees, cache directories, or SDK tool directories, add only the
specific paths required by that workflow. The full checklist for worktrees,
cache writes, command rules, and recovery paths is in
[Codex Autonomy Setup](../guides/codex-autonomy-setup).
