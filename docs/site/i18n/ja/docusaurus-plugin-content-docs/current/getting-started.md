---
id: getting-started
title: はじめに
sidebar_position: 1
---

# はじめに

この guide では、fresh checkout から local Tasq service の起動、最初の issue 作成までを扱います。

## Install

repository を clone し、`tq` command を build します。

```sh
git clone https://github.com/version-1/tasq.git
cd tasq
go build -o ./bin/tq ./cmd/tq
```

binary を `PATH` に追加するか、`./bin/tq` として実行します。

```sh
export PATH="$PWD/bin:$PATH"
tq version
```

## Initialize Local State

Tasq は local service data を `$TQ_HOME` 配下に保存します。`TQ_HOME` が未設定の場合は current user の default home directory を使います。

service 起動前に database migration を適用します。

```sh
tq migrate
```

issue-tracker、orchestrator、Web UI service を起動します。

```sh
tq service start
tq service status
```

current repository を Tasq project として登録します。`--key` を渡さない場合、project key は directory name から推測されます。

```sh
tq project add --key tasq .
tq project check tasq
```

## Create Your First Issue

project に issue を作成します。

```sh
tq issue create \
  --project tasq \
  --title "Write onboarding notes" \
  --description "Capture the first setup path for new contributors."
```

issue を一覧します。

```sh
tq issue list --project tasq
```

issue の詳細を確認します。

```sh
tq issue get 1
```

workflow に沿って issue status を進めます。

```sh
tq issue ready 1
tq issue update 1 --status in_progress
tq comment add 1 --type progress --body "Started work."
tq issue update 1 --status review
```

local service が起動している場合は Web UI を開けます。

```sh
tq web
```

## Useful Defaults

- `tq` は issue-tracker URL を `--api-url`、`TQ_API_URL`、`$TQ_HOME/system/state.json`、`http://localhost:37651` の順に解決します。
- script や agent が structured output を必要とする場合は `--output json` を使います。
- service log は `$TQ_HOME/system/log/` に書き込まれ、`tq logs` で確認できます。
