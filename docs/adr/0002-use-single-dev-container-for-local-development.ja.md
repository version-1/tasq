# ADR-0002: Local development では単一 dev container を使う

## Context

これまでの Tasq local development は、issue-tracker、orchestrator、Web、Go tooling を別々の Compose service として起動していた。この構成は service ごとの境界を見やすくする一方で、command を host で実行するか container で実行するかによって、`localhost`、service name、`TQ_HOME`、`system/state.json` の意味が変わっていた。

特に agent 開発では混乱しやすい。orchestrator は `codex app-server` を起動し、`tq`、TUI、Web、issue-tracker、orchestrator は同じ issue-tracker endpoint を安定して解決する必要がある。host と container で同じ `TQ_HOME` を共有すると、一方の network namespace でしか使えない address が state に保存される可能性がある。

また、Codex には isolation boundary が必要である。agent runner を host で直接動かすと、filesystem や credential exposure の範囲を推論しづらくなる。

## Decision

Local development では、単一の `dev` container と standalone な `openapi` documentation UI service を使う。

`dev` container 内で issue-tracker、orchestrator、Web、`tq`、TUI、Codex CLI を同じ container namespace 上で動かす。次を使う。

- `TQ_HOME=/workspace/.tasq`
- `CODEX_HOME=/home/codex/.codex`
- 固定 non-root user
- Go cache、Web `node_modules`、Codex credential 用の named volume

Codex authentication は dev container 内で `codex login` を実行して行う。認証情報は `codex-home` named volume に保存する。Docker image には credential を含めず、default workflow では host の Codex credential を mount しない。

Process management は Makefile に残す。Makefile は issue-tracker、orchestrator、Web を dev container 内で起動し、background log を `.tmp/dev-logs/` に保存し、TUI は interactive command として扱う。

Default development Compose file から、issue-tracker、orchestrator、Web、Go tools の分割 service を削除する。

## Alternatives

### Split Compose services を維持する

Service ごとに container を分ける topology は分かりやすく、Compose logs も扱いやすい。しかし default dev workflow としては host/container address translation が残り、`TQ_HOME` state sharing が複雑になるため採用しない。

### すべて host で動かす

`localhost` と host path は自然になるが、Codex 周りの container isolation boundary がなくなり、local dependencies の再現性も落ちる。advanced manual workflow としては可能だが default にはしない。

### Process manager を使う

supervisord、overmind、foreman などで dev container 内の複数 process を管理できる。初回 migration では Makefile-based process management で十分なため、追加 runtime dependency は導入しない。

## Consequences

Local development は単一 network namespace を使うため、`state.json` が `127.0.0.1` を指しても、dev container 内の issue-tracker、orchestrator、`tq`、TUI から同じ意味で解決できる。

Dev container は Codex の isolation boundary でもある。`codex-home` volume は secret-bearing local state として扱う。この volume を削除すると再 login が必要になる。

Default の `make tq` は dev container 内で実行される。endpoint consistency は改善するが、project path を永続化する command では注意が必要である。container workspace path が見える可能性があるため、project record の durable model としては ADR-0001 の host-local path 方針を維持する。

Process lifecycle は Makefile が所有する。二重起動対策は狭い process pattern に依存し、background log は `.tmp/dev-logs/` に保存される。

旧 Compose service name に依存していた script や開発習慣は、dev-container target へ移行する必要がある。

## ADR-0001 との関係

ADR-0001 は project と workspace path の durable product model を記録している。つまり、それらは container runtime path ではなく host-local absolute path である。この ADR は default local development topology を変更するが、その durable model は置き換えない。

そのため、新しい dev-container workflow で project-path persistence command を default の `make tq` flow に含める前に、host-aware な path strategy が必要である。それまでは、project-path command は `/workspace` をそのまま永続化するのではなく、host-aware workflow として扱う。
