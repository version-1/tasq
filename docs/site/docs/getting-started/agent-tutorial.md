---
id: agent-tutorial
title: Agent Tutorial
sidebar_position: 3
---

# Agent Tutorial

In this tutorial, you will register a small repository, ask a coding agent to
plan a TypeScript and React TODO app, create a Tasq issue from that plan, and
follow the work through to a GitHub pull request.

Complete [Install](pathname:///getting-started/install) first.

## Prerequisites

- The `tq` command is installed and available on your `PATH`.
- `tq service start` is running. Confirm it with `tq service status`.
- Codex CLI is installed and authenticated for GitHub, or Claude Code is
  installed and authenticated. You can also perform the `tq` commands yourself
  when neither agent is available.
- Git and a GitHub account are available.

## 1. Register the Tutorial Project

### Fork and Clone `tasq-todo`

Fork [version-1/tasq-todo](https://github.com/version-1/tasq-todo) on GitHub,
then clone your fork. Replace `<your-account>` with your GitHub account name.

```sh
git clone https://github.com/<your-account>/tasq-todo.git
cd tasq-todo
```

You can use your own repository instead. Use a project key that is unique and
easy to recognize in Tasq.

### Add the Project to Tasq

Register the repository and validate its workflow.

```sh
tq project add --key tasq-todo .
tq project check tasq-todo
```

`tq project add` links the current repository to the project key. It also
creates a default `WORKFLOW.md` only when the repository does not already have
one.

## 2. Understand `WORKFLOW.md`

`WORKFLOW.md` is the project-level contract that tells Tasq and its agents how
to run work in that repository. It combines runtime settings with the task
instructions given to the agent.

The [tasq-todo `WORKFLOW.md`](https://github.com/version-1/tasq-todo/blob/main/WORKFLOW.md)
demonstrates both parts:

- Its front matter configures polling, the `.worktrees/agents` workspace root,
  agent limits, and the Codex app-server command.
- Its body gives every agent the issue title, description, and a required flow:
  inspect scope, create an isolated branch or worktree, make a focused change,
  verify it, create a pull request, leave a handoff comment, and move the issue
  to `review`.

Copy this file into your own project as a starting point, then replace commands,
test expectations, and workspace settings with ones that fit your repository.
Use [Workflow Configuration](pathname:///guides/workflow-configuration) for
workflow resolution and override details.

## 3. Plan Work and Create an Issue

### Install the `tasq-cli` Skill for Codex

The optional `tasq-cli` skill gives Codex a focused reference for project,
issue, workflow, service, and log commands. Install it once, then restart Codex
so it loads the new skill.

```sh
python "${CODEX_HOME:-$HOME/.codex}/skills/.system/skill-installer/scripts/install-skill-from-github.py" \
  --repo version-1/tasq \
  --path .agents/skills/tasq-cli
```

Claude Code is a valid alternative. It can use the same `tq` commands without
this optional skill; include the command-oriented request below in your prompt
and refer to the [tq CLI reference](pathname:///reference/cli-reference) when
needed.

### Create a Plan in Plan Mode

Start Codex in plan mode in the `tasq-todo` directory. With Claude Code, use
its plan mode when available; otherwise, use the same prompt and explicitly
ask it not to edit files until you approve the plan.

Copy this prompt into the agent. It is intentionally in this tutorial for now;
it can move to `tasq-todo/docs/plan.md` later.

```md
Create a concise implementation plan for a TypeScript and React TODO app.

Do not edit files yet. Inspect the repository and propose a plan that includes:

- a component structure for creating, completing, and filtering TODO items;
- state and data-flow decisions;
- accessible interaction details;
- focused tests and verification commands; and
- any risks or assumptions.

Keep the plan small enough for one pull request.
```

Review the plan. Adjust the goal or scope before you create an issue.

### Ask the Agent to Create the Tasq Issue

After approving the plan, send this follow-up prompt. The agent should use the
installed `tq` command rather than creating an issue only in prose.

```md
Create a Tasq issue with `tq issue create`.

Use project key `tasq-todo`, title it "Build a TypeScript and React TODO app",
and put the approved plan in the issue description. Report the issue ID when
you finish.
```

Confirm the issue exists:

```sh
tq issue list --project tasq-todo
```

## 4. Inspect and Queue the Issue in the Web UI

Open the local Web UI:

```sh
tq web
```

Use the project and issue list to select `tasq-todo`, then open the issue you
just created. Check the title, description, plan, status, comments, and
activity before allowing an agent to run it.

When the issue is ready, move it to `ready` in the UI. You can make the same
transition from the CLI:

```sh
tq issue ready <issue-id>
```

## 5. Follow Agent Execution

Tasq runs only ready issues whose dependencies have been resolved. An issue
created with `--dependency <issue-id>` remains out of the runnable queue until
that dependency reaches a terminal state. This lets you split a larger change
into safe, ordered pieces instead of asking one agent to do everything at once.

Keep the issue detail page open. Use its Activity and comments to follow status
changes, run events, worktree information, and the agent's handoff.

### Recover a Blocked Codex Run

If a Codex run becomes blocked, open the issue Activity tab and copy the thread
ID from the latest run. From the repository checkout, resume that exact session:

```sh
codex resume <thread-id>
```

For a more autonomous setup that reduces repeated approval waits, follow
[Codex Autonomy Setup](pathname:///guides/codex-autonomy-setup). For recovery
details and alternatives such as `codex resume --last`, see
[Recover a Blocked Session](pathname:///guides/recover-blocked-session).

## 6. Confirm the Pull Request

When the agent completes the issue, confirm that the issue activity or comments
link to a GitHub pull request. Review the change and its reported verification
before merging.

You do not need to wait for every issue in Tasq before applying this workflow
to your own project. Start with one completed pull request, then register more
projects as needed. Tasq becomes especially useful when it tracks several
projects and their independent queues at the same time.
