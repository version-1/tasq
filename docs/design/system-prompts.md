# System Prompts

This document records the Tasq-owned prompt text that the orchestrator injects into Codex runner turns. Project-specific work instructions still come from the effective `WORKFLOW.md`.

The orchestrator injects these prompts so agents can keep the issue tracker synchronized while preserving the Symphony boundary where workflow-defined ticket mutations are performed by the coding agent.

## Task Start Prompt

On a first run, when `Task.ResumeThreadID` is empty, the runner renders the effective workflow prompt. Unless `tasq.task_work_prompt` is set to `false`, Tasq prepends the default task-work prompt before template expansion.

The injected start prompt is:

````text
Use `tq` to keep the issue tracker synchronized:

If the `tasq-cli` skill is available, use it as the preferred guidance for tracker operations.

- When work starts, move the issue to `in_progress` and leave a progress comment.
- Add progress comments at meaningful milestones during the work.
- If work is blocked, leave a blocker comment that explains why, then move the issue to `blocked`.
- When the pull request is ready for review, leave a handoff comment with the PR and verification summary, then move the issue to `review`.
- Always pass `--author codex` when posting comments.
- Run only the commands for the current lifecycle stage; the examples below are not a single script.

```sh
# Start
tq issue update {{ issue.id }} --status in_progress
tq comment add {{ issue.id }} --author codex --type progress --body "Started work."

# Meaningful progress milestone
tq comment add {{ issue.id }} --author codex --type progress --body "Implemented the change; running verification."

# Blocked (use instead of the review handoff)
tq comment add {{ issue.id }} --author codex --type blocker --body "Blocked: explain the blocker and what is needed."
tq issue update {{ issue.id }} --status blocked

# Ready for review
tq comment add {{ issue.id }} --author codex --type handoff --body "PR: <url>; verification: <summary>."
tq issue update {{ issue.id }} --status review
```

Run the installed `tq` binary from `PATH`. Do not use `go run ./cmd/tq` for
tracker synchronization.
````

Template variables such as `{{ issue.id }}` are rendered after this prompt is prepended.

## Task Resume Prompt

On a resumed run, when `Task.ResumeThreadID` is set, the runner resumes the existing Codex thread and does not resend the full workflow prompt. It sends only continuation guidance so the agent continues the same task without repeating completed work.

The injected resume prompt is:

```text
First run `tq issue update <issue-id> --status in_progress` to keep the issue tracker synchronized. Then continue the same task in this live thread. Do not repeat completed work. Stop when the workflow is ready for handoff.
```

`<issue-id>` is filled from the current task issue ID before the turn starts. The same continuation prompt is also used for later turns in a multi-turn run.
