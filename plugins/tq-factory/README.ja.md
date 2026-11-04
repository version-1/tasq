# TQ Factory plugin

English counterpart: [README.md](README.md).

TQ Factory は、Tasq の `tq` CLI で作業を管理するプロジェクト向けに、コーディングエージェントの進め方を提供するプラグインです。外部エージェントまたは人間が実作業を担い、課題とコメントを使って範囲を限定した作業を管理します。Tasq の従来の orchestrator は起動せず、エージェントを常駐実行したり、内蔵モデルでドキュメントを生成したりもしません。

## ローカル開発での導入

このパッケージには、移植可能なルートの `plugin.json`、Codex 互換の manifest、Claude 用 manifest が含まれます。ローカルの Codex クライアントで試す場合は、このパッケージをローカルまたはリポジトリの marketplace に登録し、その配布元からインストールします。Codex では、`$REPO_ROOT/.agents/plugins/marketplace.json` をリポジトリ用 marketplace として利用できます。インストール後、複数の課題にまたがるコーディング目標に `tq-factory` を使います。このパッケージ自体が marketplace catalog を追加したり、公開したりすることはありません。

marketplace の形式とローカルクライアントで利用できるかどうかは、[公式の plugin packaging guide](https://developers.openai.com/plugins/build/plugins)を参照してください。

対象プロジェクトには次が必要です。

- `tq` から到達できる issue-tracker が起動していること。
- 登録済みのプロジェクトキーがあること（`tq project list`）。
- 対象の課題とコメントを読み書きする権限があること。

ローカルの既定値以外の tracker を使う場合は、`TQ_API_URL` を設定するか `--api-url` を指定します。書き込む前に `tq --output json issue list --project <project-key>` を実行し、接続先を確認します。

## 最小の流れ

1. プロジェクトキーと具体的な目標を渡し、エージェントに `tq-factory` を使うよう依頼します。
2. エージェントは Tasq に目標課題と子課題を作成または再開します。
3. 外部エージェントは、割り当てられた子課題を分離したワークスペースで進め、`tq comment add` で進捗と引き継ぎの根拠を記録します。
4. 実装担当とは独立した reviewer が検証結果を残します。権限を持つ担当者が、検証済みの課題を完了にし、目標の一覧を更新します。

正確なコマンド形式と状態の扱いは、[skill](skills/tq-factory/SKILL.md)と、そこから参照するワークフローの説明を確認してください。

## オフライン検証シナリオ

`tq` を呼び出さずに、[scenario](skills/tq-factory/evals/offline-scenario.md)を確認できます。架空のエージェント実行環境を前提にせず、安全な課題の状態と引き継ぎを判断できるかを検証します。
