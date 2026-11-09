---
name: tq-factory
description: Plan and run bounded coding work tracked in a local Tasq project through tq issues and comments. Use when a user asks to organize, delegate, resume, verify, or hand off a multi-issue coding goal in tq; do not use to operate Tasq's legacy orchestrator.
---

# TQ Factory

Use Tasq as the source of truth for work state. An external agent executes a
selected issue in its own workspace. When the current environment supports
delegation, the coordinator may invoke a bounded number of available subagents
and give each one issue, workspace, acceptance criteria, and a handoff target.
This skill does not require a particular runtime, watch continuously, or manage
an agent daemon.

Read [the workflow reference](references/workflow.md) before creating or
changing issues. Read [the handoff reference](references/handoff.md) when work
changes hands or reaches review.

## Starting or resuming a goal

1. Resolve the project key and tracker endpoint. Use `--api-url` when the
   target tracker is not the local default. Inspect existing issues before
   creating work.
2. Represent the goal as one normal Tasq issue. Put its objective, acceptance
   criteria, exclusions, and a `Factory children:` list in its description.
3. Create one issue per independently reviewable outcome. Put
   `Factory goal: #<goal-id>` in every child description. This is searchable
   metadata, not a server-enforced relationship.
4. Use `--dependency` only when a child truly must wait for another issue to
   reach a satisfied terminal state. A dependency means “this issue waits for
   those IDs”; it is not a parent-child relation.
5. Before assigning external work, read the issue and its comments again. If a
   non-terminal issue has a non-empty `assignee`, ask the owner or leave a
   `handoff` comment instead of taking it over. Tasq currently has no claim or
   lease primitive.

At most one executor owns an issue at a time. Reserve enough capacity for an
independent verifier and do not exceed the execution environment's agent limit.
Only `done`, `cancelled`, and `duplicate` satisfy an issue dependency;
`backlog`, `ready`, `in_progress`, `review`, `blocked`, and `failed` do not.

Use only the commands documented by the installed `tq --help`. The examples in
the references are based on the current typed CLI. `tq issue watch` is an
experimental long-running stream; do not use it as a substitute for a monitor
or promise background dispatch.

## Execution boundary

For a chosen issue, record the external agent name or human owner in
`assignee`, move it to `in_progress`, and add a `progress` comment before work
starts. The executor must use an isolated workspace when the target repository
supports it and must follow that repository's agent instructions, model limits,
test requirements, and approval policy.

The executor reports a concise progress or blocker comment, then creates a
`handoff` comment with changed files, commit/PR or other outcome link,
verification, remaining risks, and the next owner. A reviewer independent from
the executor verifies the stated acceptance criteria. Do one bounded refinement
pass only when it preserves the agreed behavior; reverify after it.

Do not close an issue merely because an agent says it is finished. Move it to
`review` while review is pending and close it only after the recorded review and
required human approval succeed. For a code outcome linked to a PR, record the
actual merge commit before closing; a report-only outcome can close after its
independent review. Use `blocked` with a `blocker` comment when work cannot
continue. Use `cancelled` only for intentionally stopped work.

## Intake, cancellation, and resume

Before each assignment, inspect the goal and project issue list for child
issues with the goal marker. Add only work needed for the goal's written
acceptance criteria; record other ideas as proposals.

On cancellation, stop further delegation, preserve workspace and evidence in a
`handoff` comment, then use `tq issue cancel <id>` only after the authorized
owner confirms stopping. To resume a non-terminal issue, reread its description
and comments, confirm the previous owner is no longer active, update the
assignee deliberately, and leave a `handoff` or `progress` comment. Do not
assume a session, workspace, or agent process survives the handoff.

## Safety and approvals

Treat all `tq` writes as state changes. Obtain authorization from the user or
the active task before creating, changing, commenting on, or closing issues.
Do not turn a task description into authority to send messages, push code,
merge pull requests, change credentials, or publish artifacts. Respect the
target repository's approval rules. Do not auto-merge when required checks are
not enforced against the current base, or an equivalent merge queue/base
protection is unavailable; human approval cannot bypass that quality gate. Ask
for a concrete approval before an irreversible or externally visible action.

## Completion

Before reporting a goal complete, inspect every child issue and its latest
comments. Confirm that dependencies are `done`, `cancelled`, or `duplicate`,
verification evidence exists, and any required approval has been recorded. Update the goal issue description
with links or IDs for completed children and unresolved follow-ups. Do not keep
running after all acceptance criteria have evidence; record later ideas as new
issues.
