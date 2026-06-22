---
id: development-setup
title: Development Setup
sidebar_position: 1
---

# Development Setup

Tasq development は Docker Compose でも host 直接でも実行できます。Compose は services を long-lived な `dev` container にまとめ、host-only operation は `tq service start` を使います。

## Compose Environment

Compose workflow は issue-tracker、orchestrator、Web server を `dev` container 内で実行します。runtime state の default は container 内の `/workspace/.tasq` です。

```sh
make dev-up
make dev-ports
```

推奨される development commands は次のとおりです。

- `make run-issue-tracker`
- `make run-orchestrator`
- `make run-web`
- `make run-migrate`
- `make dev-codex-login`

## Host-Only Environment

host-only local operation では、`tq` を build し、`TQ_HOME` を設定し、migrations を適用して services を起動します。

```sh
make build-tq
export PATH="$PWD/bin:$PATH"
export TQ_HOME="$PWD/.tasq"
tq migrate
tq service start
```

services は loopback ports を使い、discovery state を `$TQ_HOME/system/state.json` 配下に書き込みます。

## Dev Container 内の Codex

dev container は `CODEX_HOME=/home/codex/.codex` を使い、`codex-home` named volume で backed されます。container 内で device auth による認証を行うため、`make dev-codex-login` を一度実行してください。
