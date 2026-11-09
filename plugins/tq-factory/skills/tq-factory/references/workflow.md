# TQ factory workflow

Use an explicit project key in every create command. `tq` resolves its tracker
endpoint as `--api-url`, then `TQ_API_URL`, then local service state, then
`http://localhost:37651`; pass `--api-url` for a deliberate non-default target.

## Goal and work-item shape

Create a goal first. The `Factory children:` section is maintained by the
coordinator as a readable index; it does not replace checking each child.

```sh
tq issue create --project my-project --title "Reduce import failures" \
  --description $'## Goal\n\nMake imports recoverable.\n\n## Acceptance criteria\n\n- A failed import records a useful error.\n- Retrying preserves earlier files.\n\n## Exclusions\n\n- No new storage provider.\n\n## Factory children\n\n- Pending.'
```

Create a child with enough context for a new session to work safely. Capture
the goal ID returned by `tq` and replace `<goal-id>` below.

```sh
tq issue create --project my-project --title "Record import retry failure" \
  --description $'Factory goal: #<goal-id>\n\n## Outcome\n\nStore an actionable failure reason.\n\n## Acceptance criteria\n\n- The API exposes the reason.\n- Existing successful imports keep working.\n\n## Handoff\n\nNo work has started.' \
  --priority high
```

Edit the goal description to add `- #<child-id> Record import retry failure`.
Use `tq issue get <id>` to inspect details before every state-changing step.

## Dependencies and assignment

Use dependencies only for execution order. To make issue `42` wait for issue
`41`, create it with `--dependency 41`, or replace its dependency set later:

```sh
tq issue update 42 --dependency 41
```

This replaces the full dependency list. Use `--clear-dependencies` only when
the issue can proceed without any prerequisite. Do not infer ownership from a
dependency: use `assignee` plus comments.

```sh
tq issue update 42 --assignee codex --status in_progress
tq comment add 42 --author codex --type progress \
  --body "Started in worktree .worktrees/42; acceptance criteria reviewed."
```

If `assignee` is non-empty or recent comments show active work, do not replace
it silently. Leave a `handoff` comment or ask the current owner to resolve the
conflict. There is no atomic claim operation.

## State, progress, and interruption

Use the built-in state values deliberately:

| Situation | Action |
| --- | --- |
| Ready for an executor | `tq issue ready <id>` |
| Work starts | `tq issue update <id> --status in_progress` |
| Needs independent review | `tq issue update <id> --status review` |
| Cannot proceed | `tq issue update <id> --status blocked` and a `blocker` comment |
| Intentional stop | `tq issue cancel <id>` and a `handoff` comment |
| Verified completion | `tq issue close <id>` |

Use `progress` comments for short updates and `handoff` for durable context.
Always pass `--author` explicitly. If the local service is unavailable, report
the failure and stop; do not create a shadow state file or assume an update was
accepted.

## Intake and priority

Before starting another child, run `tq issue list --project <key>` and inspect
the goal's child references. Work that has `ready` status and no active
dependencies is a candidate, but priority, owner, collision risk, and available
review capacity still determine selection. Add a new child only if it is needed
for the goal's written acceptance criteria; otherwise create a separate issue
or report it as a proposal.

Use `tq --output json issue list --project <key>` only when an agent needs to
parse the response. Do not use `tq issue watch` as a background service from a
one-shot agent session.

## Cancel and resume

Cancellation preserves evidence before state changes. Ask the authorized owner,
write a `handoff` comment that records the workspace, current outcome, and
restart condition, then use `tq issue cancel <id>`. A cancelled issue satisfies
a dependency but is not automatically reopened. To resume work, create a new
ready issue or explicitly restore the intended status after the owner confirms
the scope; record why in a handoff comment.
