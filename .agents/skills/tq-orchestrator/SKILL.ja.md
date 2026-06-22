---
name: tq-orchestrator
description: Monitor tool で tq issue tracker から ready な issue を polling し、 subagents に任せるスキル
---

# 制約

本スキルは Claude Code エージェントで実行されることを想定している。
Claude Code 以外での実行を行った場合は エラーを返す。

# 目的

tasq のオーケストレータ層として issue tracker から ready な issue を polling し、 subagents に任せるスキルを提供する。

1. Monitor tool で issue tracker の状況を監視して、ready な issue が存在し、issue を subagents に任せる手順
2. Workflow Template の定義

# 禁止事項

- issue を subagents に委譲せずに、スキルを実行したエージェント自身が issue を解決することは禁止する。

# 手順

1. ユーザに polling の間隔（秒）を確認する。（デフォルト: 30）
2. **Monitor tool** で `tq issue watch --interval <秒>` を実行する。tq issue tracker から ready な issue を polling し、stdout に1行1 JSON エンベロープを出力する（1行が Monitor の1通知になる）。
3. Ready な issue を検出すると event エンベロープを出力する（下記「出力プロトコル」参照）。同じ issue は `--seen-ttl` が経過するまで再出力されない。
4. event エンベロープを受け取ったら、`body.id` から issue id を読み取り、Workflow Template をもとに作業を subagents に引き渡す
  - subagents への引き渡しが完了したら、issue のステータスを `tq issue update <issue_id> --status in_progress` で in_progress に変更する。
  - subagent の名前を `tq comment add <issue_id> --author claude-code --body "<comment>"` で issue にコメントとして残す。

# 出力プロトコル

> [!NOTE]
> `tq issue watch` は experimental なコマンドです。最悪削除しても CLI の他機能に影響しないよう、意図的に独立させてあります。

`tq issue watch` はグローバルの `--output` フラグを無視し、常に1行1 JSON を次の3種いずれかで出力する:

- `{"type":"error","body":"<message>"}` — 一時的な polling 失敗。ループは継続する。常に出力。
- `{"type":"event","eventType":"issue-ready","body":<issue>}` — ready な issue を検出。`body` は issue の全フィールド。常に出力。
- `{"type":"info","body":"<message>"}` — 起動時の設定値と各周回のサマリ。`--verbose` 指定時のみ出力。

フラグ:
- `--interval <秒>`: polling 間隔（デフォルト 30）。
- `--seen-ttl <秒>`: 同じ issue の再出力を抑制する期間（デフォルト 900。`--interval` より大きいこと）。
- `--verbose`: info エンベロープも出力する。

ループは停止されるまで（SIGINT・kill・Monitor の timeout）回り続け、自発的には終了しない。

# Workflow Template

1. Issue を `tq issue get <issue_id>` で取得する。
2. Confirm the task scope from the issue title and description above.
3. Before branch creation, run `git fetch origin`, then create the task branch from `origin/main` or the repository's default branch.
4. Create or switch to an isolated task branch and worktree before editing.
  - git worktree を使用して `<project_root>/.worktrees/tasq/<issue_id>-<issue-summary-title>` に作業用ワークツリーを作成して作業を開始する。
  - 作成した worktree/branch 名を `tq comment add <issue_id> --author claude-code --body "<comment>"` で issue にコメントとして残す。
5. Make focused changes that satisfy the issue.
6. Run the narrowest useful verification first, then broaden checks when shared behavior is affected.
7. Commit the change and create or update a pull request.
  - 作成した pull request のタイトル・URL を issue に `tq comment add <issue_id> --author claude-code --body "<comment>"` でコメントとして残す。
8. 未対応事項など引き継ぐ内容がある場合は、`tq comment add <issue_id> --author claude-code --body "<comment>"` でコメントを残す.
9. 作業が完了したら issue のステータスを `tq issue update <issue_id> --status review` で review に変更する。
10. 作業が完了したら、作業に使用した Worktree/Branch を削除する。
   - Pull Request が作成できていること, ステータスを review に変更していることを確認して実行する
   - 上記条件を満たしていない場合は、作業用ワークツリーの削除を行わない。

## エラー時やコマンド実行許可が必要な場合の対応

タスクの実行中のエラーや実行許可が必要で作業が継続できない場合は、
  1. 継続不能な理由を `tq comment add <issue_id> --author claude-code --body "<comment>"` で issue にコメントを残す
  2. `tq issue update <issue_id> --status blocked` で issue のステータスを blocked に変更する

