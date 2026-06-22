# ADR-0002: ローカル開発では単一の開発コンテナを使う

## Context

これまでの Tasq ローカル開発では、issue-tracker、orchestrator、Web、Go tooling を別々の Compose service として起動していた。この構成はサービスごとの境界を見やすくする一方で、コマンドをホストで実行するかコンテナで実行するかによって、`localhost`、service name、`TQ_HOME`、`system/state.json` の意味が変わっていた。

特にエージェント開発では混乱しやすい。orchestrator は `codex app-server` を起動し、`tq`、TUI、Web、issue-tracker、orchestrator は同じ issue-tracker エンドポイントを安定して解決する必要がある。ホストとコンテナで同じ `TQ_HOME` を共有すると、一方のネットワーク名前空間でしか使えないアドレスが状態に保存される可能性がある。

また、Codex には分離境界が必要である。エージェント runner をホストで直接動かすと、ファイルシステムや認証情報がさらされる範囲を推論しづらくなる。

## Decision

ローカル開発では、単一の `dev` コンテナと standalone な `openapi` documentation UI service を使う。

`dev` コンテナ内で issue-tracker、orchestrator、Web、`tq`、TUI、Codex CLI を同じコンテナ名前空間上で動かす。次を使う。

- `TQ_HOME=/workspace/.tasq`
- `CODEX_HOME=/home/codex/.codex`
- 固定の非 root ユーザー
- Go cache、Web `node_modules`、Codex 認証情報用の named volume

Codex の認証は `dev` コンテナ内で `codex login --device-auth` を実行して行う。認証情報は `codex-home` named volume に保存する。Device auth により、コンテナ内にだけ存在する localhost callback へブラウザーがリダイレクトして失敗する問題を避ける。Docker image には認証情報を含めず、既定のワークフローではホストの Codex 認証情報をマウントしない。

プロセス管理は Makefile に残す。Makefile は issue-tracker、orchestrator、Web を `dev` コンテナ内で起動し、バックグラウンドログを `.tmp/dev-logs/` に保存し、TUI は対話型コマンドとして扱う。

既定の development Compose file から、issue-tracker、orchestrator、Web、Go tools の分割 service を削除する。

## Alternatives

### 分割 Compose services を維持する

Service ごとにコンテナを分ける構成は分かりやすく、Compose logs も扱いやすい。しかし既定の開発ワークフローとしてはホストとコンテナ間のアドレス変換が残り、`TQ_HOME` の状態共有が複雑になるため採用しない。

### すべてホストで動かす

`localhost` とホストパスは自然になるが、Codex 周りのコンテナ分離境界がなくなり、ローカル依存関係の再現性も落ちる。高度な手動ワークフローとしては可能だが既定にはしない。

### Process manager を使う

supervisord、overmind、foreman などで `dev` コンテナ内の複数プロセスを管理できる。初回移行では Makefile ベースのプロセス管理で十分なため、追加の実行時依存関係は導入しない。

## Consequences

ローカル開発は単一のネットワーク名前空間を使うため、`state.json` が `127.0.0.1` を指しても、`dev` コンテナ内の issue-tracker、orchestrator、`tq`、TUI から同じ意味で解決できる。

`dev` コンテナは Codex の分離境界でもある。`codex-home` volume は秘密情報を含むローカル状態として扱う。この volume を削除すると再 login が必要になる。

既定の `make tq` は `dev` コンテナ内で実行される。エンドポイントの一貫性は改善するが、project path を永続化するコマンドでは注意が必要である。コンテナのワークスペースパスが見える可能性があるため、project record の永続モデルとしては ADR-0001 のホストローカルパス方針を維持する。

プロセスライフサイクルは Makefile が所有する。二重起動対策は狭いプロセスパターンに依存し、バックグラウンドログは `.tmp/dev-logs/` に保存される。

旧 Compose service name に依存していた script や開発習慣は、dev-container target へ移行する必要がある。

## ADR-0001 との関係

ADR-0001 は project と workspace path の永続的な製品モデルを記録している。つまり、それらはコンテナ実行時パスではなくホストローカルの絶対パスである。この ADR は既定のローカル開発構成を変更するが、その永続モデルは置き換えない。

そのため、新しい dev-container workflow で project-path 永続化コマンドを既定の `make tq` flow に含める前に、ホストを考慮したパス戦略が必要である。それまでは、project-path コマンドは `/workspace` をそのまま永続化するのではなく、ホストを考慮するワークフローとして扱う。
