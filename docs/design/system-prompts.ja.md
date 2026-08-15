# System Prompts

このドキュメントでは、オーケストレーターが Codex runner のターンに注入する、Tasq 所有のプロンプト文字列を記録します。プロジェクト固有の作業指示は、引き続き有効な `WORKFLOW.md` から取得します。

オーケストレーターは、「ワークフローで定義した課題の変更はコーディングエージェントが行う」という Symphony の境界を保ちながら、エージェントが Issue Tracker を同期できるようにこれらのプロンプトを注入します。

## Tasq コマンドの解決

オーケストレーターは、課題を割り当てる前にビルドプロファイルから Tasq CLI コマンドを決定します。

| ビルドプロファイル | コマンド |
|---|---|
| 空（本番） | `tq` |
| `dev` | `tqdev` |

その他のプロファイルはエラーになります。`tq service start` とオーケストレーターの単体起動は、選択したコマンドを `PATH` から解決できることを要求します。

開発ビルドのプロンプトでは、ターン固有の指示より前に次の文言を追加します。

```text
Use the `tqdev` command instead of `tq`.
When using the `tasq-cli` skill, interpret every `tq` command as `tqdev`.
```

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

Tasq は、開発ビルド用のコマンド解決ガイダンス、タスク作業プロンプトの順に追加し、最後に `{{ issue.id }}` や `{{ tq.command }}` などの変数を展開します。`tasq.task_work_prompt: false` を設定すると、初回ターンでは両方の注入を無効にしますが、継続時の動作は変わりません。

承認理由に関する文言はエージェントの振る舞いを案内するものであり、実行時検証や Tasq の承認拒否方針は変更しません。Pull Request Artifact の指示は、エージェントが PR を作成または更新した場合にだけ適用します。

## タスク再開時のプロンプト

再開時に `Task.ResumeThreadID` が設定されている場合、runner は既存の Codex thread を再開し、完全なワークフロープロンプトは再送しません。完了済みの作業を繰り返さず同じタスクを続けるため、継続用の指示だけを送信します。

再開時に注入されるプロンプトは次のとおりです。

```text
First run `<tasq-command> issue update <issue-id> --status in_progress` to keep the issue tracker synchronized. Then continue the same task in this live thread without repeating completed work, and stop when it is ready for handoff. Before requesting approval for a command execution or file change, provide a non-empty, specific reason. The reason must identify what needs approval, the target scope (the command and working directory or the file paths), why approval is required, and the expected effect. Do not send a null, empty, or vague reason such as only saying that approval is required. If this continuation creates or updates a pull request, register the primary PR before handoff with `<tasq-command> artifact set <issue-id> --type pull_request <pr-url>`. On success, add the handoff comment, then move the issue to `review`; on failure, retry reasonably, then leave a blocker comment and do not move to `review` if it remains unresolved. Otherwise, artifact registration is not required.
```

`<tasq-command>` には [Tasq コマンドの解決](#tasq-コマンドの解決)で決定したコマンドを使い、`<issue-id>` には現在の課題 ID を設定します。開発ビルドでは、上記と同じ 2 行の読み替えガイダンスを先頭へ追加します。

有効な複数ターン実行の後続ターンでも、同じ継続プロンプトを使います。既存の再開条件または継続条件が次のターンを選んだ場合だけ送信し、継続が無効な場合に余分なターンを追加することはありません。担当する変更要求の指示がある場合は、この注意事項の後ろに追加します。
