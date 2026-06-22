# ADR-0003: Makefile target で開発コンテナコマンドの意図を分離する

## Context

ADR-0002 でローカル開発は単一の `dev` コンテナへ移行した。その後も Makefile には曖昧なコマンド境界が残っていた。ある target はホスト向けの entry point で、別の target は Docker Compose 操作、さらに別の target は `dev` コンテナ内でプロセスを実行するものだったが、名前だけではそれらの役割が分かりにくかった。

この曖昧さにより、予想外の挙動が起きた。既存コンテナに入るだけ、または `tq` を実行するだけのコマンドが、別プロセスの起動、rebuild、待機、停止まで行う可能性があった。これにより、日常的なコマンドの挙動を推論しづらくなり、失敗原因も実行したコマンドと関係ないものに見えやすくなった。

Makefile はローカル開発のプロセス管理インターフェースであり続ける。そのため、コマンド名と target dependency は意図を明確に伝える必要がある。

## Decision

Makefile target をコマンドの意図ごとに分ける。

- `dev-*` target はホスト向けのローカル開発 entry point とする。
- `dc-*` target は Docker Compose と dev-container 操作を扱う。
- `run-*` target は、すでに起動している `dev` コンテナ内でプロセスやコマンドを実行する。

`dc-shell` と `dc-exec` は既存の `dev` container に attach する。コンテナの起動、rebuild、recreate は行わない。コンテナが存在しない場合は、環境状態を変更せずに失敗する。

`run-tq` は `dev` コンテナ内で `tq` コマンドを実行することだけを責務にする。issue-tracker の起動、issue-tracker の待機、既存プロセスの停止、サービスライフサイクル管理は行わない。issue-tracker が起動していない場合、`run-tq` は下位の API 接続エラーで失敗する。

`run-orchestrator`、`run-web`、`run-tui` のようなサービス指向の target は、そのワークフローがサービス依存関係を必要とするため、issue-tracker の readiness を確認してよい。この readiness の挙動は `run-tq` には共有しない。

Makefile help output は prefix taxonomy を説明し、target を section ごとに表示する。

## Alternatives

### 互換 alias を残す

旧 target name を残すと短期的な移行負荷は下がる。しかし曖昧なコマンドの意味も残る。既定のローカル開発インターフェースでは意図を見えるようにするべきなので、互換 alias は残さない。

### すべての `run-*` target で service を保証する

一部のコマンドは便利になるが、単純なコマンドがプロセス状態を変更するようになる。また、失敗が要求されたコマンドによるものかサービス起動によるものか分かりづらくなる。Readiness を管理するのは、サービスワークフローとして動作する target だけにする。

### Compose service name を直接使う

開発者が `docker compose` を直接呼ぶこともできる。しかしコマンド知識が散らばり、ローカルワークフローの一貫性が落ちる。Makefile を文書化されたインターフェースとして維持する。

## Consequences

コマンドインターフェースはより明示的になる。開発者は環境ライフサイクルのコマンドと、既存の `dev` コンテナに attach するだけ、または `dev` コンテナ内で実行するだけのコマンドを区別できる。

`run-tq` は暗黙の挙動を減らす。issue-tracker がすでに起動している必要はあるが、無関係なサービスプロセスには触らない。これにより診断で使いやすくなり、意図しない restart を避けられる。

旧コマンドや開発習慣の一部は新しい prefix へ移行する必要がある。Makefile reference は新しい名前と example を記録する。

失敗はより直接的になる。`dc-shell` が失敗するなら `dev` コンテナが起動していない。`run-tq` が接続エラーで失敗するなら、`dev` コンテナ内から issue-tracker に到達できない。

## Notes

この ADR は ADR-0002 以降に導入した Makefile の挙動を精緻化する。ADR-0002 の単一コンテナ構成や、ADR-0001 のホストローカル project path モデルは置き換えない。
