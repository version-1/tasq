# ADR-0003: Makefile target で dev container command の意図を分離する

## Context

ADR-0002 で local development は単一の `dev` container へ移行した。その後も Makefile には曖昧な command boundary が残っていた。ある target は host-facing な entry point で、別の target は Docker Compose operation、さらに別の target は dev container 内で process を実行するものだったが、名前だけではそれらの役割が分かりにくかった。

この曖昧さにより、予想外の挙動が起きた。既存 container に入るだけ、または `tq` を実行するだけの command が、別 process の起動、rebuild、待機、停止まで行う可能性があった。これにより、日常的な command の挙動を推論しづらくなり、失敗原因も実行した command と関係ないものに見えやすくなった。

Makefile は local development の process-management surface であり続ける。そのため、command name と target dependency は意図を明確に伝える必要がある。

## Decision

Makefile target を command intent ごとに分ける。

- `dev-*` target は host-facing な local development entry point とする。
- `dc-*` target は Docker Compose と dev-container operation を扱う。
- `run-*` target は、すでに起動している dev container 内で process や command を実行する。

`dc-shell` と `dc-exec` は既存の `dev` container に attach する。container の起動、rebuild、recreate は行わない。container が存在しない場合は、environment state を変更せずに失敗する。

`run-tq` は dev container 内で `tq` command を実行することだけを責務にする。issue-tracker の起動、issue-tracker の待機、既存 process の停止、service lifecycle management は行わない。issue-tracker が起動していない場合、`run-tq` は underlying API connection error で失敗する。

`run-orchestrator`、`run-web`、`run-tui` のような service-oriented target は、その workflow が service dependency を必要とするため issue-tracker readiness を確認してよい。この readiness behavior は `run-tq` には共有しない。

Makefile help output は prefix taxonomy を説明し、target を section ごとに表示する。

## Alternatives

### Compatibility alias を残す

旧 target name を残すと短期的な migration friction は下がる。しかし曖昧な command meaning も残る。Default local development surface では意図を見えるようにするべきなので、compatibility alias は残さない。

### すべての `run-*` target で service を保証する

一部の command は便利になるが、単純な command が process state を変更するようになる。また、失敗が requested command によるものか service startup によるものか分かりづらくなる。Readiness を管理するのは、service workflow として動作する target だけにする。

### Compose service name を直接使う

Developer が `docker compose` を直接呼ぶこともできる。しかし command knowledge が散らばり、local workflow の一貫性が落ちる。Makefile を documented interface として維持する。

## Consequences

Command surface はより明示的になる。Developer は environment lifecycle command と、既存 dev container に attach するだけ、または dev container 内で実行するだけの command を区別できる。

`run-tq` は magic を減らす。issue-tracker がすでに起動している必要はあるが、無関係な service process には触らない。これにより diagnostics で使いやすくなり、意図しない restart を避けられる。

旧 command や開発習慣の一部は新しい prefix へ移行する必要がある。Makefile reference は新しい名前と example を記録する。

Failure はより直接的になる。`dc-shell` が失敗するなら dev container が起動していない。`run-tq` が connection error で失敗するなら、dev container 内から issue-tracker に到達できない。

## Notes

この ADR は ADR-0002 以降に導入した Makefile behavior を refine する。ADR-0002 の single-container topology や、ADR-0001 の host-local project path model は置き換えない。
