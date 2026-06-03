# ADR-0007: Dev Process Logs を TQ Home 配下に保存する

## Context

Tasq local development は ADR-0002 で定義した single dev container 内で issue-tracker、
orchestrator、Web processes を起動する。Makefile はこれまで background process logs を
repository workspace の `.tmp/dev-logs/` 配下へ書き出していた。

この場所は initial dev-container migration では見つけやすかったが、runtime state が 2 か所に
分かれる問題があった。Databases と service discovery state は既に `$TQ_HOME/system/` 配下に
あり、process logs だけが `$TQ_HOME` の外にあった。そのため cleanup、inspection、runtime-state
ownership の一貫性が弱かった。

## Decision

Development service logs は次の場所へ書き込む。

```text
$TQ_HOME/system/log/
```

Makefile では host 側と container 側の展開を分けて扱う。

- Host-facing log commands は `$(TQ_HOME)/system/log/*.log` を follow する。
- Dev-container process commands は `${TQ_HOME}/system/log/*.log` へ書き込む。

Log directory は runtime state であり、source-controlled project content ではない。

## Alternatives

### `.tmp/dev-logs/` 配下に残す

これは ADR-0002 の initial implementation detail を維持し、logs を Tasq runtime data から
視覚的に分離できる。しかし service logs は operational runtime state なので、他の
`$TQ_HOME/system/` state と同じ場所に置く方がよいと判断した。

### `$TQ_HOME/system/data/` 配下に置く

すべての runtime artifacts を 1 つの subtree に置けるが、append-only process logs と SQLite
databases や attachment data が混ざる。`system/log/` を分けることで、operational logs の
inspection と cleanup をしやすくする。

### Docker Compose logs だけを使う

File log management は不要になるが、default dev workflow は 1 つの dev container 内で複数の
background processes を起動する。Process ごとの file の方が follow しやすく、bug report にも
添付しやすい。

## Consequences

`make run-logs` は `$TQ_HOME/system/log/` の logs を follow する。

Default の repository-local setup では、host 上の logs は `.tasq/system/log/` 配下に現れる。
Dev container 内では、同じ logical location が `/workspace/.tasq/system/log/` になる。

Tasq runtime state の cleanup や archive は、logs、databases、state discovery files を 1 つの
root から扱える。既存の local `.tmp/dev-logs/` files は disk に残る可能性があるが、新しい dev
process output はそこへ書き込まない。

## Notes

この ADR は process-log location に対する後続変更を記録する。Single dev container へ移行した
historical decision record である ADR-0002 は意図的に書き換えない。
