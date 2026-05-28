# Workflow

## Worktree Usage

Create worktrees under sequential directories from `.worktrees/1` to `.worktrees/n` at the repository root, and use them as the working directories for tasks.

When working on multiple tasks at the same time, assign an unused number to each task and keep it separate from existing worktrees.

Example:

```sh
git worktree add .worktrees/1 <branch>
git worktree add .worktrees/2 <branch>
```

Before starting work, check the existing numbers under `.worktrees/` and use the next available number.
