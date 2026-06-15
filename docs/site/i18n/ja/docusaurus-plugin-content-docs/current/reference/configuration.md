---
id: configuration
title: Configuration
sidebar_position: 3
---

# Configuration

Tasq は machine-level configuration、runtime state、service logs、SQLite data の local home directory として `TQ_HOME` を使います。

default では `TQ_HOME` は `~/.tasq` に解決されます。repository-local development では workspace directory に設定してください。

```sh
export TQ_HOME="$PWD/.tasq"
```

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

`config/` は user-editable です。`system/` は Tasq processes が管理します。

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
  }
}
```

## 解決順序

| Value | Resolution |
| --- | --- |
| Issue-tracker API URL | `--api-url`, `TQ_API_URL`, `state.json`, `http://localhost:37651` |
| Comment author | `--author`, `TQ_AUTHOR`, `config.yaml author`, `$USER` |
| Orchestrator concurrency | workflow と `config.yaml max_concurrent_agents` の小さい方 |
