---
id: development-setup
title: 開発環境セットアップ
sidebar_position: 1
---

# 開発環境セットアップ

Tasq development は Docker Compose 内でも host 上でも実行できます。Compose は service を長寿命の `dev` container にまとめ、host-only operation は `tq service start` を使います。

## Compose environment

Compose workflow は issue-tracker、orchestrator、Web server を `dev` container 内で実行します。runtime state の default は container 内の `/workspace/.tasq` です。

```sh
make dev-up
make dev-ports
```

推奨される development command は次のとおりです。

- `make run-issue-tracker`
- `make run-orchestrator`
- `make run-web`
- `make run-migrate`
- `make dev-codex-login`

## Host-only environment

host-only local operation では、`tq` を build し、`TQ_HOME` を設定し、migration を適用して service を起動します。

```sh
go build -o ./bin/tq ./cmd/tq
export PATH="$PWD/bin:$PATH"
export TQ_HOME="$PWD/.tasq"
tq migrate
tq service start
```

service は loopback port を使い、discovery state を `$TQ_HOME/system/state.json` 配下に書き込みます。

## Dev Container 内の Codex

dev container は `CODEX_HOME=/home/codex/.codex` を使い、`codex-home` named volume に保存します。container 内で device auth するため、`make dev-codex-login` を一度実行してください。
