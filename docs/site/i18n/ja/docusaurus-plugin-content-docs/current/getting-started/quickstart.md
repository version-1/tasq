---
id: quickstart
title: クイックスタート
sidebar_position: 2
---

# クイックスタート

この手順では fresh checkout から Tasq を起動し、最初の issue を作成します。

## CLI をビルドする

```sh
git clone https://github.com/version-1/tasq.git
cd tasq
go build -o ./bin/tq ./cmd/tq
export PATH="$PWD/bin:$PATH"
tq version
```

## ローカル状態を初期化する

Tasq は local state を `TQ_HOME` 配下に保存します。未設定の場合、Tasq は `~/.tasq` を使います。

```sh
export TQ_HOME="$PWD/.tasq"
tq migrate
tq service start
tq service status
```

`tq service start` は issue-tracker、orchestrator、Web server を local loopback ports で起動し、discovery state を `$TQ_HOME/system/state.json` に書き込みます。

## Project を登録する

現在の repository を登録し、issues を project に scope できるようにします。

```sh
tq project add --key tasq .
tq project check tasq
```

## Issue を作成して移動する

```sh
tq issue create \
  --project tasq \
  --title "Write onboarding notes" \
  --description "Capture the first setup path for new contributors."
```

```sh
tq issue list --project tasq
tq issue get 1
tq issue ready 1
tq issue update 1 --status in_progress
tq comment add 1 --type progress --body "Started work."
tq issue update 1 --status review
```

## Web UI を開く

```sh
tq web
```

Web UI は local service がすでに動いていることを前提にします。開けない場合は `tq service status` を確認し、`tq logs web` で logs を調べてください。
