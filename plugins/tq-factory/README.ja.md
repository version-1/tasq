# TQ Factory plugin

English counterpart: [README.md](README.md).

TQ Factory は、Tasq の `tq` CLI を作業管理に使うプロジェクト向けのコーディングエージェント用 workflow plugin です。外部エージェントまたは人間が実作業を行い、課題とコメントで範囲を限定した作業を管理します。Tasq の従来 orchestrator を起動せず、エージェントの常駐実行や内蔵モデルによるドキュメント生成も行いません。

## ローカル開発での導入

この package には、portable な root `plugin.json`、Codex 互換 manifest、Claude manifest が含まれます。ローカルの Codex client で試す場合は、この package をローカルまたは repository marketplace に追加し、その source から install します。Codex は repository marketplace として `$REPO_ROOT/.agents/plugins/marketplace.json` を利用できます。その後、複数の課題にまたがるコーディング目標に `tq-factory` を使います。この package 自体は marketplace catalog の追加や公開を行いません。

marketplace の形式とローカル client での利用可否は、[公式の plugin packaging guide](https://developers.openai.com/plugins/build/plugins)を参照してください。

対象プロジェクトには次が必要です。

- `tq` から到達できる issue-tracker が起動していること。
- 登録済みの project key があること（`tq project list`）。
- 対象の課題とコメントを読み書きする権限があること。

ローカル既定値以外の tracker を使う場合は、`TQ_API_URL` を設定するか `--api-url` を指定します。書き込み前に `tq --output json issue list --project <project-key>` で接続先を確認します。

## 最小の流れ

1. project key と具体的な目標を渡し、エージェントに `tq-factory` を使うよう依頼します。
2. エージェントは Tasq に目標課題と子課題を作成または再開します。
3. 外部エージェントは、割り当てられた子課題を分離した workspace で進め、`tq comment add` で進捗と引き継ぎの根拠を記録します。
4. 独立した reviewer が検証結果を残します。権限を持つ担当者が、検証済みの課題を完了にし、目標の一覧を更新します。

正確なコマンド形式と状態の扱いは、[skill](skills/tq-factory/SKILL.md)とリンク先の workflow reference を参照してください。

## オフライン検証シナリオ

`tq` を呼び出さずに、[scenario](skills/tq-factory/evals/offline-scenario.md)を確認できます。架空のエージェント実行環境を前提にせず、安全な課題状態と引き継ぎを判断できるかを検証します。
