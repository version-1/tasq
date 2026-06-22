# ADR-0007: Dev Process Logs を TQ Home 配下に保存する

## Context

Tasq のローカル開発では、ADR-0002 で定義した単一の開発コンテナ内で issue-tracker、
orchestrator、Web プロセスを起動する。Makefile はこれまでバックグラウンドプロセスログを
リポジトリワークスペースの `.tmp/dev-logs/` 配下へ書き出していた。

この場所は初回の dev-container 移行では見つけやすかったが、実行時状態が 2 か所に
分かれる問題があった。Databases とサービス discovery state は既に `$TQ_HOME/system/` 配下に
あり、プロセスログだけが `$TQ_HOME` の外にあった。そのためクリーンアップ、確認、実行時状態の
所有範囲に一貫性がなかった。

## Decision

開発サービスログは次の場所へ書き込む。

```text
$TQ_HOME/system/log/
```

Makefile ではホスト側とコンテナ側の展開を分けて扱う。

- ホスト向けのログコマンドは `$(TQ_HOME)/system/log/*.log` を追跡する。
- 開発コンテナのプロセスコマンドは `${TQ_HOME}/system/log/*.log` へ書き込む。

ログディレクトリは実行時状態であり、ソース管理対象のプロジェクト内容ではない。

## Alternatives

### `.tmp/dev-logs/` 配下に残す

これは ADR-0002 の初期実装詳細を維持し、ログを Tasq 実行時データから
視覚的に分離できる。しかしサービスログは運用上の実行時状態なので、他の
`$TQ_HOME/system/` 状態と同じ場所に置く方がよいと判断した。

### `$TQ_HOME/system/data/` 配下に置く

すべての実行時成果物を 1 つの subtree に置けるが、追記型のプロセスログと SQLite
databases や attachment data が混ざる。`system/log/` を分けることで、運用ログの
確認とクリーンアップをしやすくする。

### Docker Compose logs だけを使う

ファイルログ管理は不要になるが、既定の開発ワークフローは 1 つの開発コンテナ内で複数の
バックグラウンドプロセスを起動する。プロセスごとのファイルの方が追跡しやすく、bug report にも
添付しやすい。

## Consequences

`make run-logs` は `$TQ_HOME/system/log/` のログを追跡する。

既定の repository-local setup では、ホスト上のログは `.tasq/system/log/` 配下に現れる。
開発コンテナ内では、同じ論理的な場所が `/workspace/.tasq/system/log/` になる。

Tasq の実行時状態のクリーンアップやアーカイブは、ログ、databases、state discovery files を 1 つの
root から扱える。既存のローカル `.tmp/dev-logs/` files は disk に残る可能性があるが、新しい開発プロセスの出力はそこへ書き込まない。

## Notes

この ADR はプロセスログの保存場所に対する後続変更を記録する。単一の開発コンテナへ移行した
履歴上の意思決定記録である ADR-0002 は意図的に書き換えない。
