---
id: setup-guide
title: Setup Guide
sidebar_position: 3
---

# Setup Guide

This guide covers setup outside `tq` that makes repeated local agent work predictable.

Keep these settings scoped to trusted local projects. Avoid granting broad global permissions for commands that can change Git history, remote state, credentials, or system configuration.

## Codex Authentication

For normal interactive work, sign in with ChatGPT so Codex uses subscription access and the active ChatGPT workspace controls.

```sh
codex login
```

For remote terminals, use device authentication.

```sh
codex login --device-auth
```

Use API key authentication only when you intentionally want usage billed through an OpenAI Platform account instead of ChatGPT subscription access.

## Trust Local Projects

Codex loads project-local `.codex/` settings only for trusted projects. Add every checkout or worktree that agents will use.

```toml
# ~/.codex/config.toml

[projects."/Users/YOU/src/tasq"]
trust_level = "trusted"

[projects."/Users/YOU/src/tasq/.worktrees/agents/issue-57"]
trust_level = "trusted"
```

Use absolute paths. If an agent runner creates disposable worktrees, add the concrete worktree path before starting Codex.

## Allow Git Metadata Writes

Linked worktrees often keep Git metadata outside the workspace directory. Commands such as `git rebase`, `git commit`, and `git checkout` may need access to the parent repository `.git` directory.

Resolve the required paths from each worktree:

```sh
git rev-parse --path-format=absolute --git-common-dir
git rev-parse --path-format=absolute --git-dir
```

Then allow the workspace and the required Git metadata path in the Codex permission profile.

```toml
[permissions.tasq-workspace.workspace_roots]
"/Users/YOU/src/tasq" = true
"/Users/YOU/src/tasq/.worktrees/agents/issue-57" = true
"/Users/YOU/src/tasq/.git" = true
```

Prefer exact paths. Granting the whole parent `.git` directory is simpler for many worktrees, while granting only a specific `.git/worktrees/<name>` path is narrower.

## Command Rules

Allow routine read and verification commands narrowly. Keep remote writes and destructive operations behind prompts or safe wrappers.

Useful low-risk rules usually cover:

- `tq issue get`, `tq issue list`, and `tq comment list`.
- `git status`, `git diff`, `git log`, and `git show`.
- `make dev-docs-build` for documentation changes.
- Read-only `gh pr view`, `gh pr diff`, and `gh pr checks`.

Do not globally allow broad commands such as `rm`, `git reset`, `git push`, direct `gh pr edit`, or shell wrappers with arbitrary arguments.
