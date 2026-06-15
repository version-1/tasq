# Setup Guide

This guide covers setup that happens outside `tq` but affects how smoothly agents can work on Tasq projects.

Use it when preparing a local machine, Codex profile, or agent runner for repeated Tasq work. Keep the settings scoped to trusted local projects and avoid granting broad access globally.

## Goals

- Let Codex update Git metadata for each managed project, including worktrees.
- Mark each local Tasq project or worktree as trusted so project-local Codex settings can load.
- Allow routine task commands without repeated prompts.
- Sign in to Codex with ChatGPT subscription access for interactive local work.

## Codex Authentication

For normal local development, sign in with ChatGPT so Codex uses subscription access and the active ChatGPT workspace controls.

```sh
codex login
```

For remote or headless terminals, use device authentication when browser login is not practical.

```sh
codex login --device-auth
```

Avoid API key authentication for ordinary interactive Tasq work unless you intentionally want usage billed through the OpenAI Platform account instead of ChatGPT subscription access.

## Trust Local Projects

Codex only loads project-local `.codex/` settings, hooks, and rules for trusted projects. Add every local Tasq checkout or worktree that agents will use.

```toml
# ~/.codex/config.toml

[projects."/Users/YOU/src/tasq"]
trust_level = "trusted"

[projects."/Users/YOU/src/tasq/.worktrees/agents/issue-56"]
trust_level = "trusted"
```

Use absolute paths. If an agent runner creates disposable worktrees, the runner should add the concrete worktree path it starts.

## Allow Git Metadata Writes

Worktree checkouts often keep their Git metadata outside the workspace directory. A command like `git rebase` may need to write under the parent repository's `.git/worktrees/<name>` directory, even when the code workspace itself is writable.

For each managed project, resolve the Git metadata locations:

```sh
git rev-parse --path-format=absolute --git-common-dir
git rev-parse --path-format=absolute --git-dir
```

Then allow the required Git metadata path in Codex configuration. For a simple checkout, the project `.git` directory is usually enough. For a linked worktree, include the parent repository's `.git` directory or the specific `.git/worktrees/<name>` directory if your runner can compute it.

Direct `workspace-write` configuration:

```toml
# ~/.codex/config.toml

sandbox_mode = "workspace-write"

[sandbox_workspace_write]
writable_roots = [
  "/Users/YOU/src/tasq/.git",
  "/Users/YOU/src/tasq/.git/worktrees/issue-56",
]
```

Permission profile configuration is better when you want a reusable policy:

```toml
# ~/.codex/config.toml

default_permissions = "tasq-workspace"

[permissions.tasq-workspace.workspace_roots]
"/Users/YOU/src/tasq" = true
"/Users/YOU/src/tasq/.worktrees/agents/issue-56" = true
"/Users/YOU/src/tasq/.git" = true

[permissions.tasq-workspace.filesystem]
":minimal" = "read"

[permissions.tasq-workspace.filesystem.":workspace_roots"]
"." = "write"
"**/*.env" = "deny"
```

Prefer exact paths over broad globs for writable locations. If several agents share one parent repository, granting the whole parent `.git` directory is simpler but broader; granting only `.git/worktrees/<name>` is narrower but may still need parent ref/log writes for some Git operations.

## Command Rules

Use Codex rules for commands that are expected during normal Tasq tasks. Rules are evaluated outside the sandbox, so keep `allow` entries narrow and use `prompt` for commands that mutate remote state.

Create a user-level rules file:

```text
~/.codex/rules/default.rules
```

Example starting point:

```python
prefix_rule(
    pattern = ["tq", "issue", ["get", "list"]],
    decision = "allow",
    justification = "Reading Tasq issues is part of normal task setup",
)

prefix_rule(
    pattern = ["tq", "comment", "list"],
    decision = "allow",
    justification = "Reading Tasq issue comments is part of normal task setup",
)

prefix_rule(
    pattern = ["make", ["help", "dev-docs-build"]],
    decision = "allow",
    justification = "Repository verification commands for docs-site tasks",
)

prefix_rule(
    pattern = ["git", ["status", "diff", "log", "show", "rev-parse", "merge-base", "branch"]],
    decision = "allow",
    justification = "Read-only Git inspection commands",
)

prefix_rule(
    pattern = ["gh", "pr", ["status", "view", "checks", "diff"]],
    decision = "allow",
    justification = "Read-only GitHub PR inspection commands",
)

prefix_rule(
    pattern = ["~/.codex/bin/safe-git-push"],
    decision = "prompt",
    justification = "Pushing branches changes remote state and should stay explicit",
)

prefix_rule(
    pattern = ["~/.codex/bin/safe-gh-edit"],
    decision = "prompt",
    justification = "Editing GitHub issues or PRs changes remote state and should stay explicit",
)
```

Restart Codex after changing rules.

Test a rule before relying on it:

```sh
codex execpolicy check --pretty \
  --rules ~/.codex/rules/default.rules \
  -- tq issue get 56
```

## Runner Checklist

When an agent runner creates a worktree, it should prepare these values before starting Codex:

1. Workspace path, such as `/Users/YOU/src/tasq/.worktrees/agents/issue-56`.
2. Git common directory from `git rev-parse --path-format=absolute --git-common-dir`.
3. Worktree Git directory from `git rev-parse --path-format=absolute --git-dir`.
4. A trusted project entry for the workspace path.
5. Writable access for the workspace and the Git metadata path needed by Git commands.

This avoids the common failure where file edits work but `git rebase`, `git checkout`, or `git commit` is blocked because Git metadata lives outside the workspace.

## What Not To Allow Globally

Do not globally allow destructive or broad commands such as:

- `rm`
- `git reset`
- `git checkout`
- `git push`
- `gh pr edit`
- shell wrappers like `bash -lc` or `zsh -lc`

Use narrow wrappers, explicit prompt rules, or per-task approval for those commands.
