---
id: configuration
title: 設定
sidebar_position: 3
---

# 設定

Tasq は `TQ_HOME` を machine-level configuration、runtime state、service log、SQLite data のための local home directory として使います。

default では、`TQ_HOME` は `~/.tasq` に解決されます。repository-local development では workspace directory に設定してください。

```sh
export TQ_HOME="$PWD/.tasq"
```

## Directory layout

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

`config/` は user-editable です。`system/` は Tasq process が管理します。

## config.yaml

```yaml
author: "jiro"
max_concurrent_agents: 3
```

| Field | Default | Description |
| --- | ---: | --- |
| `author` | `$USER` | `--author` と `TQ_AUTHOR` が未設定の場合の default comment author。 |
| `max_concurrent_agents` | `10` | orchestrator run の machine-wide concurrency limit。 |

## state.json

実行中の service は discovery metadata を `system/state.json` に書き込みます。

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
  }
}
```

## 解決順序

| Value | Resolution |
| --- | --- |
| Issue-tracker API URL | `--api-url`、`TQ_API_URL`、`state.json`、`http://localhost:37651` |
| Comment author | `--author`、`TQ_AUTHOR`、`config.yaml author`、`$USER` |
| Orchestrator concurrency | workflow と `config.yaml max_concurrent_agents` の小さい方 |
