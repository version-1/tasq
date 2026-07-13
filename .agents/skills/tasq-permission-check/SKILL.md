---
name: tasq-permission-check
description: Verify Codex execution-policy coverage before autonomous Tasq issue work. Use this skill whenever a user asks to configure, audit, troubleshoot, or confirm Codex command permissions, Rules, requestApproval behavior, blocked autonomous runs, or execpolicy checks for tq workflows. It inventories the Tasq lifecycle commands, checks them sequentially against a Rules file, and reports every command that would not run without approval.
compatibility: Requires `codex execpolicy check`, `jq`, and a Codex Rules file.
---

# Tasq Permission Check

## Purpose

Confirm that the Commands required for autonomous Tasq issue work are allowed
by a specific Codex Rules file before an agent starts work. This catches a
missing Rule before it turns an otherwise routine run into `requestApproval` or
a blocked issue.

## Safety Boundary

This is a policy check, not a live workflow rehearsal. Do not create projects,
issues, comments, worktrees, commits, pushes, or pull requests merely to test
permissions. The bundled script passes representative command tokens to
`codex execpolicy check`; it does not execute those commands.

Do not edit Rules, change Codex configuration, or approve a command as part of
this skill unless the user separately asks for that change.

## Inputs

Ask for the Rules file when it is not clear. Prefer the project-provided
`codex/rules/tasq-dev.rules` when it exists; otherwise use the path supplied by
the user. The result applies only to the Rules file that was checked, not to an
unknown managed or user-level policy layer.

## Procedure

1. Read [the required command matrix](references/required-commands.md). It
   separates commands that only inspect state from commands that mutate local
   or remote state during a normal Tasq lifecycle.
2. Confirm that `codex` and `jq` are available. If either is missing, report
   the missing dependency and stop.
3. Run the bundled checker from the repository root:

   ```sh
   bash .agents/skills/tasq-permission-check/scripts/check-execpolicy.sh \
     --rules codex/rules/tasq-dev.rules
   ```

4. Read every result in order. `allow` means the checked Rules file predicts no
   `requestApproval` for that command. Any other decision (`prompt`,
   `forbidden`, or no matching rule) is a gap; do not run the corresponding
   real command just to test it.
5. For each gap, report the exact representative command, decision, and its
   lifecycle purpose. Recommend the user review
   `codex/rules/tasq-dev.rules` and add only the narrow rule that their workflow
   needs.
6. If the user asks for a live follow-up, run only non-mutating commands such
   as `tq version`, `tq project list`, `git status --short`, and `gh pr status`.
   Check their policy first. Never use live project creation, issue updates,
   comments, worktree creation/removal, commits, pushes, or PR creation as a
   permission probe.

## Report Format

Use this structure:

```md
# Tasq Permission Check

- Rules file: `<path>`
- Result: PASS | GAPS FOUND

## Allowed Commands

| Lifecycle step | Representative command | Decision |
| --- | --- | --- |

## Permission Gaps

| Lifecycle step | Representative command | Decision | Required action |
| --- | --- | --- | --- |

## Notes

- `allow` is a policy prediction that no `requestApproval` is expected for the
  checked command.
- The checker did not execute any mutating lifecycle command.
```

When every command is allowed, state that the checked policy covers the listed
Tasq lifecycle. When there are gaps, state that autonomous execution is not yet
ready and leave the Rules change to the user unless they request it.
