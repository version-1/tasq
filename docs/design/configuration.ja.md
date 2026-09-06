# Configuration

English counterpart: [configuration.md](configuration.md).

`TQ_HOME`、ディレクトリ構成、`config.yaml`、`state.json`、解決順序については [設定リファレンス](../site/i18n/ja/docusaurus-plugin-content-docs/current/reference/configuration.md)（サイト版）を参照してください。このドキュメントでは、開発時固有の設定である build profile と Compose dev container だけを扱います。

## Build Profile

開発用バイナリには build profile を埋め込めます。profile が空の場合は既定の home として `~/.tasq` を使い、`dev` のような profile は `~/.tasq-dev` に解決されます。profile は英小文字、数字、ハイフンだけで構成します。明示的に設定した `TQ_HOME` は常に build profile より優先されます。

直接サービスを起動する場合と managed startup で同じ runtime state を使うため、`tq`、`issue-tracker`、`orchestrator`、`web` には同じ profile を埋め込む必要があります。service port は共有のままなので、profile 分離によってサービスを同時起動できるようにはなりません。

## Compose Dev Container

既定の Compose 開発ワークフローでは、`dev` container 内でツールを実行し、
`TQ_HOME=/workspace/.tasq` を使います。`tq`、TUI、issue-tracker、orchestrator はすべて、その container 内の同じ実行時状態を読みます。

Codex の認証情報は `TQ_HOME` とは分離します。dev container では
`CODEX_HOME=/home/codex/.codex` を使い、`codex-home` named volume に保存します。
container 内で一度 `make dev-codex-login` を実行し、device auth で認証します。`codex-home`
volume を削除するとログイン状態も削除されます。device auth を使うことで、container 内にだけ存在する localhost callback へブラウザがリダイレクトして失敗する問題を避けます。

リポジトリ管理の Codex rules は `codex/rules/` に置き、dev container 内の
`/home/codex/.codex/rules` へ読み取り専用でマウントします。認証情報、個人用の上書き設定、
生成された承認判断、その他の秘密情報を含む Codex state は、リポジトリではなく
`codex-home` volume に保持します。
