# Tasq 設計

Tasq の設計ドキュメントは、関心ごとに分割しています。

- [Architecture](design/architecture.ja.md): 所有境界、コンポーネント、依存関係、状態の所有者、現在の MVP の動作。
- [API](design/api.ja.md): issue-tracker API サーフェスと CLI コマンドの対応関係。
- [Configuration](design/configuration.ja.md): ローカルホームディレクトリ、実行時状態、Compose 開発設定。
- [Schema](design/schema.ja.md): エンティティフィールド仕様と検証ルール。
- [Status](design/status.ja.md): 課題 status、派生的なキュー状態、想定されるワークフロー遷移。
- [Operations](design/operations.ja.md): ローカル開発環境、検証、未決定事項。
- [Deployment](design/deployment.ja.md): release tag 作成、GitHub Actions、GoReleaser のデプロイフロー。

Web UI の構造とスタイル規約は [design/web.ja.md](design/web.ja.md) を参照してください。
