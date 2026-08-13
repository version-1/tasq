# System Prompts

This document records the Tasq-owned prompt text that the orchestrator injects into Codex runner turns. Project-specific work instructions still come from the effective `WORKFLOW.md`.

The orchestrator injects these prompts so agents can keep the issue tracker synchronized while preserving the Symphony boundary where workflow-defined ticket mutations are performed by the coding agent.

## Task Start Prompt

On a first run, when `Task.ResumeThreadID` is empty, the runner renders the effective workflow prompt. Unless `tasq.task_work_prompt` is set to `false`, Tasq prepends the default task-work prompt before template expansion.

The injected start prompt is:

````text
Use `{{ tq.command }}` to keep the issue tracker synchronized:

If the `tasq-cli` skill is available, use it as the preferred guidance for tracker operations.

- Use `{{ tq.command }}` for every Tasq CLI operation, including commands shown elsewhere as `tq`; it identifies the CLI that started the managed services.
- Prefer typed commands such as `{{ tq.command }} issue`, `{{ tq.command }} comment`, and `{{ tq.command }} artifact` for issue tracker operations.
- Use `{{ tq.command }} api` only when the issue tracker operation has no typed command.
- Do not call the issue tracker API directly with `curl`, `wget`, or a custom HTTP script. This restriction applies only to the issue tracker API, not to other services or local endpoint verification.
- When work starts, move the issue to `in_progress` and leave a progress comment.
- Add progress comments at meaningful milestones during the work.
- If work is blocked, leave a blocker comment that explains why, then move the issue to `blocked`.
- If you create or update a pull request, register the primary PR being submitted for review as the issue's `pull_request` artifact before handoff. `tq artifact set` is an upsert, so replace the artifact URL if the primary PR is recreated. Mention any supporting PRs in the handoff comment instead of registering them as the primary artifact.
- After successful registration, leave a handoff comment with the PR URL and verification summary, then move the issue to `review`. If registration fails, retry a reasonable number of times; if it remains unresolved, leave a blocker comment and do not move to `review`. Skip artifact registration when no pull request was created or updated.
- Always pass `--author codex` when posting comments.
- Run only the commands for the current lifecycle stage; the examples below are not a single script.

```sh
# Start
{{ tq.command }} issue update {{ issue.id }} --status in_progress
{{ tq.command }} comment add {{ issue.id }} --author codex --type progress --body "Started work."

# Meaningful progress milestone
{{ tq.command }} comment add {{ issue.id }} --author codex --type progress --body "Implemented the change; running verification."

# Blocked (use instead of the review handoff)
{{ tq.command }} comment add {{ issue.id }} --author codex --type blocker --body "Blocked: explain the blocker and what is needed."
{{ tq.command }} issue update {{ issue.id }} --status blocked

# Ready for review
{{ tq.command }} artifact set {{ issue.id }} --type pull_request <pr-url>
{{ tq.command }} comment add {{ issue.id }} --author codex --type handoff --body "PR: <url>; verification: <summary>."
{{ tq.command }} issue update {{ issue.id }} --status review
```

Use `{{ tq.command }}` for every Tasq CLI operation. Managed services render it as
`"$TQ_EXECUTABLE"`, a persistent snapshot of the CLI that started the services; direct
orchestrator launches fall back to `tq`. Do not substitute another executable from `PATH`.
````

Template variables such as `{{ issue.id }}` are rendered after this prompt is prepended. Setting `tasq.task_work_prompt: false` disables this entire start prompt, including its tracker and artifact instructions; it does not change continuation behavior.

The pull-request artifact represents the primary PR currently submitted for review. Setting it again replaces the prior URL, while supporting PRs remain handoff-comment context. Artifact registration is conditional on creating or updating a PR. A registration failure must be retried reasonably and, if unresolved, reported as a blocker without a transition to `review`.

## Task Resume Prompt

On a resumed run, when `Task.ResumeThreadID` is set, the runner resumes the existing Codex thread and does not resend the full workflow prompt. It sends only continuation guidance so the agent continues the same task without repeating completed work.

The injected resume prompt is:

```text
First run `<tasq-command> issue update <issue-id> --status in_progress` to keep the issue tracker synchronized. Then continue the same task in this live thread without repeating completed work, and stop when it is ready for handoff. If this continuation creates or updates a pull request, register the primary PR before handoff with `<tasq-command> artifact set <issue-id> --type pull_request <pr-url>`. On success, add the handoff comment, then move the issue to `review`; on failure, retry a reasonable number of times, then leave a blocker comment and do not move to `review` if it remains unresolved. Otherwise, artifact registration is not required.
```

`<tasq-command>` uses `"$TQ_EXECUTABLE"` in managed runs and `tq` otherwise. `<issue-id>` is filled from the current task issue ID before the turn starts. The same continuation prompt is also used for later turns in an enabled multi-turn run. It is sent only when the runner's existing resume or later-turn conditions select a continuation turn; the runner does not add an extra turn when continuation is disabled. Assigned change-request guidance, when present, remains appended after this reminder.
