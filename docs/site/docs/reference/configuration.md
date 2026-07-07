---
id: configuration
title: Configuration
sidebar_position: 3
---

# Configuration

Tasq uses `TQ_HOME` as its local home directory for machine-level configuration, runtime state, service logs, and SQLite data.

By default, `TQ_HOME` resolves to `~/.tasq`. For repository-local development, set it to a workspace directory.

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

`config/` is user-editable. `system/` is managed by Tasq processes.

## config.yaml

```yaml
author: "jiro"
max_concurrent_agents: 3
```

| Field | Default | Description |
| --- | ---: | --- |
| `author` | `$USER` | Default comment author when `--author` and `TQ_AUTHOR` are not set. |
| `max_concurrent_agents` | `10` | Machine-wide concurrency limit for orchestrator runs. |

## state.json

Running services write discovery metadata to `system/state.json`.

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

## Resolution Order

| Value | Resolution |
| --- | --- |
| Issue-tracker API URL | `--api-url`, `TQ_API_URL`, `state.json`, `http://localhost:37651` |
| Web UI URL | `state.json` web entry, then `tq service start` if the Web UI is not running |
| Comment author | `--author`, `TQ_AUTHOR`, `config.yaml author`, `$USER` |
| Orchestrator concurrency | minimum of workflow and `config.yaml max_concurrent_agents` |
