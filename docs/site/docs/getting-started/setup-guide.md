---
id: setup-guide
title: Setup Guide
sidebar_position: 4
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

For the shortest Tasq install path, see [Install](./install). For command
behavior and API resolution details, see [tq CLI](./concepts/tq-cli).

## Add a Project Workflow

Add a `WORKFLOW.md` file to the repository root so agents know the expected
route for every Tasq task in that project. Keep it operational: how to inspect
the issue, how to make changes, which commands verify the work, and when to
hand the result back.

Example:

```md
# WORKFLOW.md

## Task Intake

1. Run `tq issue get <issue-id>` and read the title, description, comments, and
   attachments before editing files.
2. Run `git status --short` and inspect the current branch.
3. Restate the goal, scope, and verification commands before making changes.

## Implementation

1. Keep changes scoped to the issue.
2. Update tests or documentation when behavior or setup instructions change.
3. Add progress comments with `tq comment add <issue-id> --type progress` when
   the task takes more than one focused step.

## Verification

1. Run the smallest relevant tests first.
2. For docs-site changes, run `npm run build` from `docs/site`.
3. Report any command that could not be run and why.

## Handoff

1. Move the issue to review only after verification passes.
2. Summarize changed files, verification results, and remaining risks.
3. If requested, create a pull request and link it from the issue comments.
```

When the workflow should live in the repository, commit `WORKFLOW.md` with the
project. If you need a machine-local override instead, store it through Tasq:

```sh
tq workflow add --project tasq-demo --file WORKFLOW.md
tq workflow show --project tasq-demo
```

For resolution order and override behavior, see
[Workflow Configuration](../guides/workflow-configuration).

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

<a id="verify-command-permission-coverage"></a>

## Verify Command Permission Coverage

Before autonomous work, use the repository's `tasq-permission-check` skill to
check the Tasq lifecycle commands against the Rules file you configured. It
uses `codex execpolicy check` and does not run the listed lifecycle commands,
so it finds missing Rules without creating projects, issues, commits, or pull
requests.

Install the skill from the Tasq repository if it is not already available to
your agent:

```sh
python "${CODEX_HOME:-$HOME/.codex}/skills/.system/skill-installer/scripts/install-skill-from-github.py" \
  --repo version-1/tasq \
  --path .agents/skills/tasq-permission-check
```

Then ask Codex to run `tasq-permission-check` for the Rules file used by your
project. For a direct check in a Tasq checkout, run:

```sh
bash .agents/skills/tasq-permission-check/scripts/check-execpolicy.sh \
  --rules codex/rules/tasq-dev.rules
```

Every result other than `allow` is a permission gap. Add only the narrow Rule
needed for that command, then run the check again. See the
[skill instructions](https://github.com/version-1/tasq/blob/main/.agents/skills/tasq-permission-check/SKILL.md)
for the lifecycle command list and reporting format.
