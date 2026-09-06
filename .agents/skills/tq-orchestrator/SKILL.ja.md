---
name: tq-orchestrator
description: Monitor ツールで Tasq の Issue Tracker を監視し、ready の課題をサブエージェントへ割り当てるスキル
---

# 目的

Tasq の Issue Tracker から ready の課題を監視し、プロジェクトのワークフローに従うサブエージェントへ各課題を割り当てる。

# 制約

- このスキルは Claude Code エージェント内でのみ実行する。それ以外の環境ではエラーを返す。
- オーケストレーターは課題の解決を必ずサブエージェントへ委譲し、自身では解決しない。
- プロジェクトの `WORKFLOW.md` を優先する。プロジェクトのワークフローに規定がない場合に限り、後述の既定ワークフローを適用する。

# 監視と割り当て

1. 監視間隔を秒単位でユーザーに確認する（既定値: 30）。
2. **Monitor ツール**で `tq issue watch --interval <seconds>` を実行する。
3. `issue-ready` イベントを受信するたびに、`body.id` から課題IDを読み取り、`tq issue get <issue_id>` で課題を取得する。
4. プロジェクトの `WORKFLOW.md` に従い、規定されていない部分だけを後述の既定ワークフローで補って、課題をサブエージェントへ割り当てる。
5. 割り当てに成功したら、次の処理を行う。
   - `tq issue update <issue_id> --status in_progress` で課題の状態を `in_progress` に変更する。
   - `tq comment add <issue_id> --author claude-code --body "<comment>"` でサブエージェント名を記録する。

# 委譲後の既定ワークフロー

1. 課題のタイトルと説明から作業範囲を確認する。
2. 編集前に、最新のデフォルトブランチを基点とする専用ブランチとワークツリーを作成する。
   - `git fetch origin` を実行する。
   - `origin/main` またはリポジトリのデフォルトブランチから作業ブランチを作成する。
   - `<project_root>/.worktrees/tasq/<issue_id>-<issue-summary-title>` にワークツリーを作成する。
   - `tq comment add <issue_id> --author claude-code --body "<comment>"` でワークツリー名とブランチ名を記録する。
3. 課題を満たす変更に絞って実装する。最小限の有効な検証から始め、共有部分に影響する場合は検証範囲を広げる。
4. 変更をコミットし、Pull Request を作成または更新する。
5. Pull Request を用意できたら、次の処理を行う。
   - `tq comment add <issue_id> --author claude-code --body "<comment>"` でタイトルと URL を記録する。
   - `tq artifact set <issue_id> --type pull_request <pr_url>` で URL を artifact として登録する。
   - 未解決事項や引き継ぎ事項があればコメントを残す。
   - `tq issue update <issue_id> --status review` で課題の状態を `review` に変更する。
6. Pull Request が存在し、課題が `review` になっていることを確認してから、作業用のワークツリーとブランチを削除する。それ以外の場合はワークツリーを保持する。

エラーやコマンド実行権限の不足など、何らかの理由で課題を完了できない場合は、次の処理を行う。

1. `tq comment add <issue_id> --author claude-code --body "<comment>"` で阻害理由を説明する。
2. `tq issue update <issue_id> --status blocked` で課題の状態を `blocked` に変更する。
3. 後続対応のため、作業用のワークツリーを保持する。

# 出力プロトコル

> [!NOTE]
> `tq issue watch` は実験的なコマンドであり、CLI の他の機能から意図的に分離されている。

`tq issue watch` はグローバルの `--output` フラグを無視し、常に1行につき1個の JSON オブジェクトを出力する。

- `{"type":"error","body":"<message>"}` — 一時的な監視エラー。ループは継続する。
- `{"type":"event","eventType":"issue-ready","body":<issue>}` — ready の課題。`body` に課題の全データが含まれる。
- `{"type":"info","body":"<message>"}` — 起動設定または監視結果の概要。`--verbose` を指定した場合のみ出力する。

フラグ:

- `--interval <seconds>`: 監視間隔（既定値: 30）。
- `--seen-ttl <seconds>`: 同じ課題のイベントを再出力しない期間（既定値: 900。`--interval` より大きい値が必要）。
- `--verbose`: info エンベロープを出力する。

ループは SIGINT、kill、Monitor のタイムアウトのいずれかで停止されるまで継続し、自動では終了しない。
