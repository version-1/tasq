---
id: recover-blocked-session
title: Recover a Blocked Session
sidebar_position: 4
---

# Recover a Blocked Session

When Codex becomes blocked, loses terminal connectivity, or needs to continue
from another shell, resume the saved session instead of starting from scratch.

Start from the Tasq Web UI so you resume the exact Codex thread tied to the
issue.

## Find the Thread ID

1. Open the issue in the Web UI.
2. Open the **Activity** tab.
3. Find the latest run row.
4. Copy the thread ID from that run.

The thread ID identifies the Codex session that was working on the issue. Use
it when you resume from the CLI.

## Resume from the CLI

Run `codex resume` from the repository checkout and pass the thread ID copied
from the Activity tab:

```sh
codex resume <thread id>
```

If the Web UI does not have a thread ID for the run, fall back to the Codex CLI
session picker:

```sh
codex resume
```

You can also resume the latest session from the current working directory:

```sh
codex resume --last
```

## Re-anchor the Session

After resuming, send a short recovery prompt that states what changed while the
session was blocked:

```text
Continue the previous task. First run git status, inspect the latest diff, and
then proceed from the current repository state.
```

Resumed sessions keep the prior transcript and plan history, so Codex can use
the earlier context while still re-checking the current working tree before
editing.
