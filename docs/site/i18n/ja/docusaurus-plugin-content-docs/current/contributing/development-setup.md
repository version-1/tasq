---
id: development-setup
title: Development Setup
sidebar_position: 1
---

# Development Setup

Tasq development は Docker Compose でも host 直接でも実行できます。service integration、
Web UI behavior、Codex / GitHub CLI authentication、shared project workflow と合わせて
確認したい作業では Compose を優先してください。host-only mode は、短い CLI 確認、
release binary の smoke test、local installed `tq` の debugging に向いています。

repository 全体の workflow は
[docs/development.md](https://github.com/version-1/tasq/blob/main/docs/development.md) を
参照してください。

## Compose Environment

Compose workflow は 1 つの long-lived な `dev` container を使います。`make dev-up` は
container を build し、issue-tracker、orchestrator、Web processes をその中で起動し、
OpenAPI UI も起動して、割り当てられた local URLs を表示します。

Compose では既定で host ports が自動割り当てされるため、固定 port を前提にしないで
ください。現在の URLs が必要な場合は `make dev-ports` を実行します。

```sh
make dev-up
make dev-ports
```

よく使う commands は次のとおりです。

- `make dev-open` は Web UI と OpenAPI UI を browser で開きます。
- `make run-tq ARGS="issue list"` は dev container 内で `tq` を実行します。
- `make run-issue-tracker`
- `make run-orchestrator`
- `make run-web`
- `make run-migrate`
- `make run-logs` は service logs を follow します。
- `make dev-down` は Compose services を停止します。

## Host-Only Environment

host-only local operation では、`tq` を build し、local binary を `PATH` の先頭に置き、
repository-local な `TQ_HOME` を設定し、migrations を適用して services を起動します。

```sh
make build-tq
export PATH="$PWD/bin:$PATH"
export TQ_HOME="$PWD/.tasq"
tq migrate
tq service start
```

host services は固定の loopback ports を使い、discovery state を
`$TQ_HOME/system/state.json` に書き込みます。

| Service | Port |
| --- | ---: |
| issue-tracker | `37651` |
| orchestrator | `37652` |
| web | `37653` |

Compose services がすでに動いている場合は、host-only mode を慎重に使ってください。
ports や runtime state が衝突する場合は、片方を停止してからもう片方を起動します。

## Dev Container 内の Codex

dev container は `CODEX_HOME=/home/codex/.codex` を使い、`codex-home` named volume で
backed されます。container 内で device auth による認証を行うため、
`make dev-codex-login` を一度実行し、`make dev-codex-status` で確認してください。

GitHub CLI credentials は `gh-config` named volume に保存されます。dev container 内から
branch 作成、push、pull request 作成を行う workflow では、事前に `make dev-gh-login`
を一度実行してください。確認には `make dev-gh-status` を使います。

Linux と WSL2 hosts では、Codex の sandboxed commands に unprivileged user namespaces が
必要です。Codex が Bubblewrap namespace error を出す場合は、container 内の package
不足だけでなく、host または Docker runtime capability の問題として扱ってください。

## Documentation Work

docs site は `docs/site` 配下にあります。dependencies の install も任せたい場合は、
repository wrappers を使います。

```sh
make dev-docs
make dev-docs-ja
make dev-docs-build
```

documentation を編集する場合は、English と Japanese pages を同期してください。
development と design decisions では English docs が primary source of truth です。
