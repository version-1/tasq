---
id: agent-tutorial
title: Agent Tutorial
sidebar_position: 3
---

# Agent Tutorial

Use this tutorial after completing [QuickStart](pathname:///getting-started/quickstart).
The goal is to let Codex or Claude Code create a task plan, register that plan
as a Tasq issue, and move the issue through the agent queue until a GitHub pull
request exists.

This tutorial assumes:

- `tq` is installed and on `PATH`.
- `tq service start` is already running.
- A project has already been registered with `tq project add`.
- Codex or Claude Code is authenticated for the local repository and GitHub.

## 1. Ask the Agent to Create a Plan

Start Codex or Claude Code in the repository you registered with Tasq. Ask it to
inspect the codebase and write a short implementation plan before creating any
issue.

Example prompt:

```md
Create a concise implementation plan for adding <task>. Do not edit files yet.
Include the goal, concrete steps, verification commands, and any risks.
```

Review the plan before continuing. The plan should be small enough for one pull
request and specific enough that another agent can execute it without guessing.

## 2. Ask the Agent to Create a Tasq Issue

After the plan looks correct, ask the agent to create a Tasq issue from it.

Example prompt:

```md
Create a Tasq issue for this plan with `tq issue create`. Use the registered
project key, include the plan in the issue description, and report the issue ID.
```

The agent should run a command similar to:

```sh
tq issue create \
  --project tasq-demo \
  --title "Add <task>" \
  --description "<plan>"
```

## 3. Inspect the Issue in the Web UI

Open the issue in the Web UI and confirm that the title, description, and plan
are correct before making it executable.

If your installed `tq` supports issue-specific Web navigation, use:

```sh
tq issue web <issue-id>
```

Otherwise, open the Web UI and select the issue:

```sh
tq web
```

## 4. Mark the Issue Ready

When the issue description is ready for execution, move it into the ready queue.

```sh
tq issue ready <issue-id>
```

At this point the issue is eligible for the orchestrator and agent runner.

## 5. Wait for Agent Progress

Keep the issue detail page open in the Web UI. The issue should move from
`ready` to `in_progress` when an agent starts working on it.

Use the Web UI activity and comments to follow progress. If the run becomes
blocked, use the run context shown in the Web UI to recover it from the CLI.

## 6. Confirm the GitHub Pull Request

When the agent completes the task, confirm that a GitHub pull request was
created and linked from the issue activity or comments.

Review the pull request, check the reported verification commands, and continue
with your normal review and merge process.
