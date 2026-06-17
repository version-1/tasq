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

1. ユーザに polling の間隔を確認する。（デフォルト: 30 秒）
2. **Monitor tool** で `./scripts/poller.sh` を実行し、tq issue tracker から ready な issue を取得する。
3. Ready な issue が存在する場合、`scripts/poller.sh` が通知 json を出力する
4. json 出力を受け取ったら、issue id を json から抽出し、Workflow Template をもとに作業を subagents に引き渡す
  - subagents への引き渡しが完了したら、issue のステータスを `tq issue update <issue_id> --status in_progress` で in_progress に変更する。
  - subagent の名前を `tq comment add <issue_id> --author claude-code --body "<comment>"` で issue にコメントとして残す。

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

