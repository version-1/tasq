# Symphony 仕様

このディレクトリには、本リポジトリで使用する Symphony service specification のローカルコピーを格納しています。

## 出典

- Upstream repository: https://github.com/openai/symphony
- Upstream document: https://github.com/openai/symphony/blob/main/SPEC.md
- `SPEC.md` に使用した raw source: https://raw.githubusercontent.com/openai/symphony/main/SPEC.md
- 取得日: 2026-05-31

## ファイル

- [SPEC.md](SPEC.md): upstream からコピーした英語版の仕様書。
- [SPEC.ja.md](SPEC.ja.md): 英語版と併せて保守する日本語訳。
- [DEVIATIONS.md](DEVIATIONS.md): upstream 仕様からの Tasq 固有の差分。
- [DEVIATIONS.ja.md](DEVIATIONS.ja.md): Tasq 固有の差分の日本語訳。
- [CODEX_APP_SERVER.md](CODEX_APP_SERVER.md): Tasq の Codex app-server transport と JSON-RPC contract。
- [CODEX_APP_SERVER.ja.md](CODEX_APP_SERVER.ja.md): Codex app-server contract の日本語訳。
- [WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md): Tasq がサポートする workflow front matter フィールドと prompt template ガイド。
- [WORKFLOW_CONTRACT.ja.md](WORKFLOW_CONTRACT.ja.md): workflow contract の日本語訳。
- [ENTITY_MAPPING.md](ENTITY_MAPPING.md): Symphony SPEC のドメインモデルと Tasq エンティティの対応関係。
- [ENTITY_MAPPING.ja.md](ENTITY_MAPPING.ja.md): Tasq エンティティ対応関係の日本語訳。

## リポジトリでの扱い

本リポジトリで Symphony 関連の orchestration、workflow、workspace、agent-runner、tracker、observability の挙動を実装する場合は、[SPEC.md](SPEC.md) を governing contract として扱います。

upstream の `SPEC.md` が更新された場合は、このディレクトリ内の両方のファイルを更新し、日本語訳を同期させてください。
