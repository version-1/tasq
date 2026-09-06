# Compose 開発環境での `tq`

`tq` のコマンド、フラグ、出力、終了ステータスに関する利用者向けの正典は、[CLI リファレンス](../site/i18n/ja/docusaurus-plugin-content-docs/current/reference/cli-reference.md)です。インストール済みの `tq` を実行する場合は、そちらを参照してください。

## 実行方法

ローカルの Compose 開発では、起動済みの `dev` コンテナ内にある CLI を Makefile ラッパーで実行します。`tq` の後ろに続く引数だけを渡してください。

```sh
make run-tq ARGS="issue list"
```

先に `make dev-up` で環境を起動します。このラッパーは、開発サービスの状態から Issue Tracker API を解決します。実行に必要な条件と関連する開発用ターゲットは、[Makefile リファレンス](makefile.ja.md#主な開発ターゲット)を参照してください。

## 関連ドキュメント

- [CLI リファレンス](../site/i18n/ja/docusaurus-plugin-content-docs/current/reference/cli-reference.md) — 利用者向けの正典となるコマンドリファレンス。
- [Tasq 運用](../design/operations.ja.md) — ローカルサービスの動作と検証。
- [Makefile リファレンス](makefile.ja.md) — 開発用コマンドと変数。
