---
id: codex-autonomy-setup
title: Codex Autonomy Setup
sidebar_position: 3
---

# Codex Autonomy Setup

Codex works best when the repository explains what it may do, the local
workspace is trusted, and routine verification commands can run without
interrupting the session.

Use this guide to prepare a Tasq checkout for long-running local Codex work.
If a session still becomes blocked despite this setup, use
[Recover a Blocked Session](./recover-blocked-session) to resume the exact
Codex thread from Tasq activity.

For Codex product details, see the official Codex documentation for
[custom instructions](https://developers.openai.com/codex/customization),
and [permissions](https://developers.openai.com/codex/enterprise/permissions).

## 1. Keep Repository Instructions Current

Codex reads repository instructions from `AGENTS.md` when a session starts.
Keep this file operational and concise:

- Preferred language and tone.
- Branch, commit, push, and PR rules.
- Required verification commands.
- Examples of wrapper scripts that make approved commands safer, when your
  workflow uses them.
- Documentation synchronization rules.

When a workflow becomes too long for `AGENTS.md`, move the details into
versioned docs and link to them from `AGENTS.md`.

## 2. Maintain Codex Configuration

Codex autonomy depends on local configuration as much as repository
instructions. Keep `~/.codex/config.toml` current when you add a Tasq checkout,
create a long-lived worktree, or introduce a toolchain that writes outside the
repository.

First, trust each checkout or worktree that Codex should operate in:

```toml
# ~/.codex/config.toml

[projects."/Users/YOU/src/tasq"]
trust_level = "trusted"
```

Then make workspace-write permissions explicit. For older Codex configuration,
use `sandbox_workspace_write.writable_roots` to allow the project, the Tasq
project `.git` directory, and tool cache directories that routine commands
need:

```toml
# ~/.codex/config.toml

sandbox_mode = "workspace-write"
approval_policy = "on-request"

[sandbox_workspace_write]
writable_roots = [
  "/Users/YOU/src/tasq",
  "/Users/YOU/src/tasq/.git",
  "/Users/YOU/Library/Caches",
  "/Users/YOU/.cache",
  "/Users/YOU/.npm",
  "/Users/YOU/.pnpm-store",
  "/Users/YOU/Library/pnpm/store",
  "/Users/YOU/go/pkg/mod",
  "/Users/YOU/.cargo",
  "/Users/YOU/.gradle",
  "/Users/YOU/.m2",
]
```

For current Codex permissions, prefer a named profile so the same filesystem
rules can be reused across sessions. Start from the
[Minimal Codex Permissions](pathname:///getting-started/setup-guide#minimal-codex-permissions)
profile in the Setup Guide, then add one `workspace_roots` entry per cache
directory your workflow writes to:

```toml
# ~/.codex/config.toml

[permissions.tasq_workspace.workspace_roots]
"/Users/YOU/Library/Caches" = true
"/Users/YOU/.cache" = true
"/Users/YOU/.npm" = true
"/Users/YOU/.pnpm-store" = true
"/Users/YOU/Library/pnpm/store" = true
"/Users/YOU/go/pkg/mod" = true
"/Users/YOU/.cargo" = true
"/Users/YOU/.gradle" = true
"/Users/YOU/.m2" = true
```

Adjust the cache list to the languages, CLIs, and SDKs used by your local
workflow. Common examples include npm, pnpm, Go, Cargo, Gradle, Maven, Python
package caches, and SDK-specific cache directories. Keep the paths narrow:
grant the exact cache directory instead of the whole home directory.

Tasq uses Git worktrees for agent work. When you add a project to Tasq, also
allow writes to that project's root `.git` path. A worktree checkout can store
its Git metadata as a file that points back to the root repository, and Codex
may need write access to the root `.git` directory for branch, commit, stash,
and worktree operations.

## 3. Allow Routine Local Commands

Codex can be more autonomous when low-risk read and verification commands do not
require repeated prompts. Keep the allowlist narrow and leave destructive
operations, history rewriting, credential changes, and direct remote writes
behind prompts or explicit wrapper scripts.

Use Codex rules for repeatable command decisions. See the official
[Codex rules documentation](https://developers.openai.com/codex/rules) for the
current `.rules` file format, rule fields, and `codex execpolicy check`.

Useful low-risk command groups for Tasq work include:

- `git status`, `git diff`, `git log`, and `git show`.
- `rg`, `sed -n`, `cat`, `find`, and `ls`.
- `go test ./...`, `npm run typecheck`, and focused frontend test commands.
- `make dev-docs-build` or `cd docs/site && npm run build` for documentation
  changes.
- Read-only `gh pr view`, `gh pr diff`, and `gh pr checks`.

Wrapper scripts are optional examples, not required Tasq commands. If your
workflow provides scripts that validate arguments and limit side effects, you
can route approved remote writes through those scripts instead of allowing the
raw command directly:

```sh
~/.codex/bin/safe-git-push
~/.codex/bin/safe-gh-edit pr <number> --body-file <file>
```

For user-local defaults, place rules in `~/.codex/rules/default.rules`. For
project-local Tasq defaults, place them in `.codex/rules/default.rules` and keep
the project trusted so Codex loads the project `.codex/` layer.

```python
# ~/.codex/rules/default.rules or .codex/rules/default.rules

prefix_rule(
    pattern = ["git", ["status", "diff", "log", "show"]],
    decision = "allow",
    justification = "Read-only Git inspection is safe for routine Tasq work.",
    match = [
        "git status --short",
        "git diff -- docs/site",
        "git log --oneline -5",
        "git show --stat",
    ],
)

prefix_rule(
    pattern = [["rg", "find", "ls", "cat"]],
    decision = "allow",
    justification = "Repository inspection commands are safe when scoped by the sandbox.",
    match = [
        "rg TODO docs",
        "find docs/site -maxdepth 2 -type f",
        "ls docs/site",
        "cat AGENTS.md",
    ],
)

prefix_rule(
    pattern = ["npm", "run", ["build", "typecheck", "test"]],
    decision = "allow",
    justification = "Tasq frontend and docs verification commands are expected.",
    match = [
        "npm run build",
        "npm run typecheck",
        "npm run test",
    ],
)

prefix_rule(
    pattern = ["go", "test"],
    decision = "allow",
    justification = "Go tests are routine local verification.",
    match = [
        "go test ./...",
        "go test ./internal/...",
    ],
)

prefix_rule(
    pattern = ["/Users/YOU/.codex/bin/safe-git-push"],
    decision = "prompt",
    justification = "Remote writes should stay explicit even when routed through a wrapper script.",
)
```

After editing rules, restart Codex so the updated files are loaded. To validate a
rule before relying on it, run:

```sh
codex execpolicy check --pretty \
  --rules ~/.codex/rules/default.rules \
  -- git status --short
```

## 4. Define the Goal and Workflow Path

For larger tasks, give Codex both a clear task goal and a documented path for
the work. The prompt should define the outcome for the current run, while
`WORKFLOW.md` should describe the repeatable steps an agent follows in the
repository.

Use the task goal to state:

- The target branch or PR.
- Files or product area in scope.
- What should not change.
- Required tests or docs build commands.
- Whether Codex should commit, push, or create a PR.

Use `WORKFLOW.md` to make the agent route explicit:

- How to inspect the current state before editing.
- Which implementation steps to take, in order.
- Which commands verify the work.
- Which actions require human approval.
- What to do when the expected route is blocked.

Keep `WORKFLOW.md` close to the work it governs. A repository-wide workflow can
live at the root, while a product area or package can have a narrower workflow
document beside its code. Link the relevant workflow from `AGENTS.md` when the
agent should always consider it.

If the task is long-running, ask Codex to keep a goal and update it as work
progresses. The combination of a clear task goal and an explicit workflow path
reduces dead ends and makes recovery easier when the terminal disconnects or
the session needs to be resumed later. For recovery steps, see
[Recover a Blocked Session](./recover-blocked-session).

## 5. Keep Autonomy Bounded

Autonomy should reduce waiting, not bypass review. Keep these actions explicit:

- Deleting files or branches.
- Rewriting Git history.
- Editing credentials, global configuration, or managed policy.
- Pushing to a remote without a prompt, review step, or wrapper that limits the
  command.
- Editing GitHub issues or PRs without a prompt, review step, or wrapper that
  limits the command.
- Running commands outside the trusted workspace.

When Codex asks for approval, review the command, working directory, and likely
side effects before allowing it.
