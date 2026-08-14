---
id: configuration
title: Configuration
sidebar_position: 3
---

# 設定リファレンス

Tasq は、マシン単位の設定、実行状態、サービスログ、SQLite データを保存するローカルホームディレクトリとして `TQ_HOME` を使います。

既定では `TQ_HOME` は `~/.tasq` です。リポジトリ内で開発する場合は、ワークスペース内のディレクトリを指定してください。

```sh
export TQ_HOME="$PWD/.tasq"
```

開発用バイナリには、`dev` のような小文字のビルドプロファイルを埋め込めます。この場合、既定のホームは `~/.tasq-dev` になります。`TQ_HOME` を明示した場合は常にそちらが優先されます。`tq`、Issue Tracker、orchestrator、Web UI が同じ状態を検出できるよう、すべてのバイナリで同じプロファイルを使用してください。

## Directory Layout

```text
$TQ_HOME/
├── config/
│   └── config.yaml
└── system/
    ├── state.json
    ├── log/
    │   ├── issue-tracker.log
    │   ├── orchestrator.log
    │   └── web.log
    └── data/
        ├── issues.sqlite
        ├── orchestrator.sqlite
        └── attachments/
```

`config/` は利用者が編集します。`system/` は Tasq のプロセスが管理します。

`tq config` を実行すると、バージョン、ビルドプロファイル、`TQ_HOME` の上書き、解決済みホーム、設定ファイルのパス、解決済みの値を確認できます。機械可読な出力には `tq --output json config` を使用します。

## config.yaml

```yaml
author: "jiro"
max_concurrent_agents: 3
```

| Field | Default | Description |
| --- | ---: | --- |
| `author` | `$USER` | `--author` と `TQ_AUTHOR` が未設定の場合の default comment author。 |
| `max_concurrent_agents` | `10` | orchestrator runs に対する machine-wide concurrency limit。 |

## state.json

running services は discovery metadata を `system/state.json` に書き込みます。

```json
{
  "issue_tracker": {
    "pid": 12345,
    "addr": "127.0.0.1:37651",
    "db": "/Users/me/.tasq/system/data/issues.sqlite",
    "started_at": "2026-06-01T10:00:00Z"
  },
  "orchestrator": {
    "pid": 12346,
    "addr": "http://127.0.0.1:37652",
    "db": "/Users/me/.tasq/system/data/orchestrator.sqlite",
    "started_at": "2026-06-01T10:00:01Z"
  },
  "web": {
    "pid": 12347,
    "addr": "127.0.0.1:37653",
    "started_at": "2026-06-01T10:00:02Z"
  }
}
```

## 解決順序

| Value | Resolution |
| --- | --- |
| Issue-tracker API URL | `--api-url`, `TQ_API_URL`, `state.json`, `http://localhost:37651` |
| Web UI URL | `state.json` の web entry。Web UI が起動していない場合は `tq service start` が必要です。 |
| Comment author | `--author`, `TQ_AUTHOR`, `config.yaml author`, `$USER` |
| Orchestrator concurrency | workflow と `config.yaml max_concurrent_agents` の小さい方 |
