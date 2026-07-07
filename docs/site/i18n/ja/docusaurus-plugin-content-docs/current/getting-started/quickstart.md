---
id: quickstart
title: クイックスタート
sidebar_position: 2
---

# クイックスタート

この手順では Tasq をインストールし、local services を起動して、1 つの
project を登録し、最初の issue を作成します。

Codex permissions や繰り返し使う agent development environment を準備する
場合は、[Setup Guide](pathname:///getting-started/setup-guide) を使ってください。

## CLI をインストールする

最新の formal release をインストールし、`tq` が `PATH` に入っていることを
確認します。

```sh
curl -fsSLO https://raw.githubusercontent.com/version-1/tasq/main/scripts/install.sh
less install.sh
sh install.sh
export PATH="${HOME}/.local/bin:${PATH}"
tq version
```

実行前に installer の内容を確認してください。release archive には `tq` CLI
に加えて、local の `issue-tracker`、`orchestrator`、`web` service binaries が
含まれます。

## Local Services を起動する

Tasq は machine-local runtime data を `TQ_HOME` 配下に保存します。未設定の
場合、Tasq は `~/.tasq` を使います。

```sh
export TQ_HOME="${HOME}/.tasq"
tq migrate
tq service start
tq service status
```

`tq service start` は issue-tracker、orchestrator、Web server を固定の local
loopback ports で起動し、discovery state を `$TQ_HOME/system/state.json` に
書き込みます。

| Service | Port |
| --- | ---: |
| issue-tracker | `37651` |
| orchestrator | `37652` |
| web | `37653` |

## Project を登録する

issues を project に scope できるように、local repository を登録します。
追跡したい repository の中で次を実行します。

```sh
tq project add --key tasq-demo .
tq project check tasq-demo
```

## 最初の Issue を作成する

```sh
tq issue create \
  --project tasq-demo \
  --title "Write onboarding notes" \
  --description "Capture the first setup path for new contributors."
```

issue が作成されたことを確認し、queue に移動します。

```sh
tq issue list --project tasq-demo
tq issue ready 1
tq issue update 1 --status in_progress
tq comment add 1 --type progress --body "Started work."
tq issue update 1 --status review
```

## Web UI を開く

```sh
tq web
```

Web UI は local services がすでに動いていることを前提にします。開けない
場合は service state と logs を確認します。

```sh
tq service status
tq logs web
```

## Services を停止または再起動する

作業が終わったら `service stop` を使います。同じ local environment を再起動
する場合は、もう一度 `service start` を実行します。

```sh
tq service stop
tq service start
```
