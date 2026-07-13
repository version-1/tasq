# Tasq ライフサイクルで必要なコマンド

次の代表コマンドを、この順番で `codex execpolicy check` に渡します。コマンドはポリシー検査用であり、実際の副作用を起こしてはいけません。

| 手順 | 代表コマンド | 必要な理由 |
| --- | --- | --- |
| トラッカーを確認する | `tq project list` | 登録済みプロジェクトを確認する。 |
| 課題を確認する | `tq issue get 1` | 割り当てられた課題を読む。 |
| 作業を開始する | `tq issue update 1 --status in_progress` | エージェントが開始したことを記録する。 |
| 引き継ぎを記録する | `tq comment add 1 --author codex --body "started"` | 進捗や blocker を発信元付きで残す。 |
| 作業を完了する | `tq issue update 1 --status review` | 課題を review に移す。 |
| リポジトリを確認する | `git status --short` | 変更前に作業ツリーを確認する。 |
| ベースを更新する | `git fetch origin` | 最新のリモート参照を取得する。 |
| 作業場所を作成する | `git worktree add .worktrees/tasq/1-example -b agent/1-example` | 課題作業を分離する。 |
| 変更をステージする | `git add README.md` | 対象を絞った変更を準備する。 |
| 変更をコミットする | `git commit -m "Implement issue 1"` | 課題の結果を保存する。 |
| ブランチを公開する | `git push origin HEAD` | 通常のタスクブランチを公開する。 |
| PR を作成する | `gh pr create --title "Implement issue 1" --body ""` | pull request を作成する。 |
| PR を確認する | `gh pr checks 1` | pull request の検証状況を確認する。 |
| 作業場所を閉じる | `git worktree remove .worktrees/tasq/1-example` | PR 作成と review への移動後にだけクリーンアップする。 |
| マージ済みブランチを削除する | `git branch -d agent/1-example` | マージ済みのローカルタスクブランチを削除する。 |
| 既存 PR を更新する（任意） | `gh pr edit 1 --title "Implement issue 1"` | 既存 PR を更新するワークフローでだけ必要。 |

force push、履歴の書き換え、認証情報の変更、広すぎるファイルシステムアクセス、対象を絞らないリモート変更は、自律実行の基本範囲外です。`forbidden` のままにするか、明示的なユーザー承認を必要としてください。
