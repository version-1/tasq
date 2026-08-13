# System Prompts

このドキュメントでは、オーケストレーターが Codex runner のターンに注入する、Tasq 所有のプロンプト文字列を記録します。プロジェクト固有の作業指示は、引き続き有効な `WORKFLOW.md` から取得します。

オーケストレーターは、「ワークフローで定義した課題の変更はコーディングエージェントが行う」という Symphony の境界を保ちながら、エージェントが Issue Tracker を同期できるようにこれらのプロンプトを注入します。

## タスク開始時のプロンプト

初回実行で `Task.ResumeThreadID` が空の場合、runner は有効なワークフロープロンプトを展開します。`tasq.task_work_prompt` が `false` に設定されていない限り、Tasq はテンプレート変数を展開する前に、既定の task-work プロンプトを先頭へ追加します。

開始時に注入されるプロンプトは次のとおりです。

````text
Use `tq` to keep the issue tracker synchronized:

If the `tasq-cli` skill is available, use it as the preferred guidance for tracker operations.

- Prefer typed `tq` commands such as `tq issue`, `tq comment`, and `tq artifact` for issue tracker operations.
- Use `tq api` only when the issue tracker operation has no typed `tq` command.
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
tq issue update {{ issue.id }} --status in_progress
tq comment add {{ issue.id }} --author codex --type progress --body "Started work."

# Meaningful progress milestone
tq comment add {{ issue.id }} --author codex --type progress --body "Implemented the change; running verification."

# Blocked (use instead of the review handoff)
tq comment add {{ issue.id }} --author codex --type blocker --body "Blocked: explain the blocker and what is needed."
tq issue update {{ issue.id }} --status blocked

# Ready for review
tq artifact set {{ issue.id }} --type pull_request <pr-url>
tq comment add {{ issue.id }} --author codex --type handoff --body "PR: <url>; verification: <summary>."
tq issue update {{ issue.id }} --status review
```

Tasq CLI のすべての操作には `{{ tq.command }}` を使います。managed service では service を起動した
CLI の永続 snapshot を `"$TQ_EXECUTABLE"` として展開し、orchestrator の直接起動では `tq` に
fallback します。`PATH` 上の別 executable に置き換えてはいけません。
````

`{{ issue.id }}` などのテンプレート変数は、このプロンプトを先頭へ追加した後に展開されます。`tasq.task_work_prompt: false` を設定すると、Issue Tracker と Artifact に関する指示を含む開始プロンプト全体が無効になりますが、継続時の動作は変わりません。

Pull Request Artifact は、現在レビューを依頼している主要 PR を表します。同じ Artifact を再設定すると以前の URL が置き換わり、補助 PR はハンドオフコメントの補足情報として残します。PR を作成または更新した場合にだけ Artifact 登録が必要です。登録に失敗した場合は合理的な回数だけ再試行し、解決できなければブロッカーコメントを残して `review` へ移動してはいけません。

## タスク再開時のプロンプト

再開時に `Task.ResumeThreadID` が設定されている場合、runner は既存の Codex thread を再開し、完全なワークフロープロンプトは再送しません。完了済みの作業を繰り返さず同じタスクを続けるため、継続用の指示だけを送信します。

再開時に注入されるプロンプトは次のとおりです。

```text
First run `<tasq-command> issue update <issue-id> --status in_progress` to keep the issue tracker synchronized. Then continue the same task in this live thread without repeating completed work, and stop when it is ready for handoff. If this continuation creates or updates a pull request, register the primary PR before handoff with `<tasq-command> artifact set <issue-id> --type pull_request <pr-url>`. On success, add the handoff comment, then move the issue to `review`; on failure, retry a reasonable number of times, then leave a blocker comment and do not move to `review` if it remains unresolved. Otherwise, artifact registration is not required.
```

`<tasq-command>` は managed run では `"$TQ_EXECUTABLE"`、それ以外では `tq` になります。`<issue-id>` はターン開始前に現在の課題 ID で埋め込まれます。同じ継続プロンプトは、有効な複数ターン実行の後続ターンでも使われます。既存の再開条件または後続ターン条件によって継続ターンが選ばれた場合だけ送信され、継続が無効な場合に余分なターンを追加することはありません。担当する変更要求の指示がある場合は、従来どおりこの注意事項の後ろに追加されます。
