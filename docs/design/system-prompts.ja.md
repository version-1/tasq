# System Prompts

このドキュメントでは、orchestrator が Codex runner の turn に注入する Tasq 所有の prompt 文字列を記録します。プロジェクト固有の作業指示は、引き続き有効な `WORKFLOW.md` から取得します。

orchestrator は、Symphony の境界である「workflow が定義した ticket mutation は coding agent が行う」を保ちつつ、agent が issue tracker を同期できるようにこれらの prompt を注入します。

## タスク開始時の Prompt

初回 run では、`Task.ResumeThreadID` が空の場合、runner は有効な workflow prompt を render します。`tasq.task_work_prompt` が `false` に設定されていない限り、Tasq は template expansion の前に既定の task-work prompt を先頭へ追加します。

開始時に注入される prompt は次のとおりです。

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

`{{ issue.id }}` などの template variable は、この prompt を先頭へ追加した後に render されます。

## タスク再開時の Prompt

再開 run では、`Task.ResumeThreadID` が設定されている場合、runner は既存の Codex thread を resume し、完全な workflow prompt は再送しません。完了済みの作業を繰り返さず同じタスクを続けるため、continuation guidance だけを送信します。

再開時に注入される prompt は次のとおりです。

```text
First run `tq issue update <issue-id> --status in_progress` to keep the issue tracker synchronized. Then continue the same task in this live thread. Do not repeat completed work. Stop when the workflow is ready for handoff.
```

`<issue-id>` は turn 開始前に現在の task の issue ID で埋め込まれます。同じ continuation prompt は、multi-turn run の後続 turn でも使われます。
