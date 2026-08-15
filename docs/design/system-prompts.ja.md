# System Prompts

このドキュメントでは、オーケストレーターが Codex runner のターンに注入する、Tasq 所有のプロンプト文字列を記録します。プロジェクト固有の作業指示は、引き続き有効な `WORKFLOW.md` から取得します。

オーケストレーターは、「ワークフローで定義した課題の変更はコーディングエージェントが行う」という Symphony の境界を保ちながら、エージェントが Issue Tracker を同期できるようにこれらのプロンプトを注入します。

## タスク開始時のプロンプト

初回実行で `Task.ResumeThreadID` が空の場合、runner は有効なワークフロープロンプトを展開します。`tasq.task_work_prompt` が `false` に設定されていない限り、Tasq はテンプレート変数を展開する前に、既定の task-work プロンプトを先頭へ追加します。

開始時に注入されるプロンプトは次のとおりです。

````text
Use `{{ tq.command }}` to keep the issue tracker synchronized. If the `tasq-cli` skill is available, use it as the preferred guidance for tracker operations.

Tracker access:
- Use `{{ tq.command }}` for every Tasq CLI operation, including commands shown elsewhere as `tq`. Do not substitute another executable from `PATH` or use `go run ./cmd/tq`.
- Prefer typed commands such as `{{ tq.command }} issue`, `{{ tq.command }} comment`, and `{{ tq.command }} artifact`. Use `{{ tq.command }} api` only when no typed command supports the operation.
- Do not call the issue tracker API directly with `curl`, `wget`, or a custom HTTP script. This restriction does not apply to other services or local endpoint verification.

Lifecycle:
- At start, move the issue to `in_progress` and leave a progress comment. Add more progress comments at meaningful milestones.
- Before requesting approval for a command execution or file change, provide a non-empty, specific reason. The reason must identify what needs approval, the target scope (the command and working directory or the file paths), why approval is required, and the expected effect. Do not send a null, empty, or vague reason such as only saying that approval is required.
- If blocked, leave a blocker comment that explains the blocker and what is needed, then move the issue to `blocked`.
- If you create or update a pull request, register the primary PR as the issue's `pull_request` artifact before handoff. Setting the artifact again replaces its URL; mention supporting PRs only in the handoff comment. If registration fails, retry reasonably; if it remains unresolved, leave a blocker comment and do not move to `review`. Skip registration when no pull request was created or updated.
- After any required artifact registration succeeds, leave a handoff comment with the PR URL and verification summary, then move the issue to `review`.
- Always pass `--author codex` when posting comments.
- Run only the commands for the current lifecycle stage; the examples below are alternatives, not a single script.

```sh
# Start
{{ tq.command }} issue update {{ issue.id }} --status in_progress
{{ tq.command }} comment add {{ issue.id }} --author codex --type progress --body "Started work."

# Meaningful progress milestone
{{ tq.command }} comment add {{ issue.id }} --author codex --type progress --body "Implemented the change; running verification."

# Blocked (use instead of the review handoff)
{{ tq.command }} comment add {{ issue.id }} --author codex --type blocker --body "Blocked: explain the blocker and what is needed."
{{ tq.command }} issue update {{ issue.id }} --status blocked

# Ready for review (include the artifact command only when a PR was created or updated)
{{ tq.command }} artifact set {{ issue.id }} --type pull_request <pr-url>
{{ tq.command }} comment add {{ issue.id }} --author codex --type handoff --body "PR: <url>; verification: <summary>."
{{ tq.command }} issue update {{ issue.id }} --status review
```

````

`{{ issue.id }}` などのテンプレート変数は、このプロンプトを先頭へ追加した後に展開されます。`tasq.task_work_prompt: false` を設定すると、Issue Tracker と Artifact に関する指示を含む開始プロンプト全体が無効になりますが、継続時の動作は変わりません。

本番ビルドでは `{{ tq.command }}` を `tq` として展開します。開発ビルドでは `tqdev` として展開し、既定のタスク作業プロンプトより前に次の 2 行を追加します。

```text
Use the `tqdev` command instead of `tq`.
When using the `tasq-cli` skill, interpret every `tq` command as `tqdev`.
```

この開発用ガイダンスも既定のタスク作業プロンプト注入の一部であるため、初回ターンで `tasq.task_work_prompt: false` を設定すると無効になります。

承認理由に関する契約を追加することで、承認待ちでブロックされたときのコメントから、運用者が必要な対応を判断できるようにします。runner は Codex app-server から受け取った request payload をすでに保持しているため、エージェントに対して、操作内容、対象範囲、必要性、想定される影響を `reason` に含めるよう指示します。この変更では、実行時の検証や Tasq の承認拒否方針は変更しません。

Pull Request Artifact は、現在レビューを依頼している主要 PR を表します。同じ Artifact を再設定すると以前の URL が置き換わり、補助 PR はハンドオフコメントの補足情報として残します。PR を作成または更新した場合にだけ Artifact 登録が必要です。登録に失敗した場合は合理的な回数だけ再試行し、解決できなければブロッカーコメントを残して `review` へ移動してはいけません。

## タスク再開時のプロンプト

再開時に `Task.ResumeThreadID` が設定されている場合、runner は既存の Codex thread を再開し、完全なワークフロープロンプトは再送しません。完了済みの作業を繰り返さず同じタスクを続けるため、継続用の指示だけを送信します。

再開時に注入されるプロンプトは次のとおりです。

```text
First run `<tasq-command> issue update <issue-id> --status in_progress` to keep the issue tracker synchronized. Then continue the same task in this live thread without repeating completed work, and stop when it is ready for handoff. Before requesting approval for a command execution or file change, provide a non-empty, specific reason. The reason must identify what needs approval, the target scope (the command and working directory or the file paths), why approval is required, and the expected effect. Do not send a null, empty, or vague reason such as only saying that approval is required. If this continuation creates or updates a pull request, register the primary PR before handoff with `<tasq-command> artifact set <issue-id> --type pull_request <pr-url>`. On success, add the handoff comment, then move the issue to `review`; on failure, retry reasonably, then leave a blocker comment and do not move to `review` if it remains unresolved. Otherwise, artifact registration is not required.
```

`<tasq-command>` は本番ビルドでは `tq`、開発ビルドでは `tqdev` になります。開発ビルドの継続プロンプトには、上記と同じ 2 行のコマンドおよび `tasq-cli` の読み替えガイダンスも先頭へ追加します。`<issue-id>` はターン開始前に現在の課題 ID で埋め込まれます。同じ継続プロンプトは、有効な複数ターン実行の後続ターンでも使われます。既存の再開条件または後続ターン条件によって継続ターンが選ばれた場合だけ送信され、継続が無効な場合に余分なターンを追加することはありません。担当する変更要求の指示がある場合は、従来どおりこの注意事項の後ろに追加されます。

再開した thread と後続 turn には完全な開始プロンプトが再送されないため、承認理由に関する同じ契約を継続用プロンプトにも記載します。
