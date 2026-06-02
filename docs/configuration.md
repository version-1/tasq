# Configuration

Japanese counterpart: [configuration.ja.md](configuration.ja.md).

Tasq uses `TQ_HOME` as its local home directory for machine-level configuration, runtime state, and service data.

By default, `TQ_HOME` resolves to `~/.tasq`. For development, set it to a repository-local directory:

```sh
TQ_HOME=./.tasq
```

The default Compose development workflow runs tools inside the `dev` container with
`TQ_HOME=/workspace/.tasq`. `tq`, TUI, issue-tracker, and orchestrator all read the same runtime
state in that container.

Codex credentials are separate from `TQ_HOME`. The dev container uses
`CODEX_HOME=/home/codex/.codex`, backed by the `codex-home` named volume. Run `make codex-login`
once to authenticate inside the container with device auth. Removing the `codex-home` volume removes
the login state. Device auth avoids browser redirects to a localhost callback that only exists inside
the container.

## Directory Layout

```text
$TQ_HOME/
├── config/
│   └── config.yaml
└── system/
    ├── state.json
    └── data/
        ├── issues.sqlite
        └── orchestrator.sqlite
```

`config/` is user-editable. `system/` is managed by Tasq processes and may be overwritten.

## config.yaml

```yaml
author: "jiro"
max_concurrent_agents: 3
```

| Field | Default | Description |
|---|---:|---|
| `author` | `$USER` | Default author used by `tq comment add` when `--author` and `TQ_AUTHOR` are not set. |
| `max_concurrent_agents` | `10` | Machine-wide concurrency limit for orchestrator agent runs. |

## state.json

Running services write discovery metadata to `system/state.json`.

```json
{
  "issue_tracker": {
    "pid": 12345,
    "addr": "127.0.0.1:51234",
    "db": "/Users/me/.tasq/system/data/issues.sqlite",
    "started_at": "2026-06-01T10:00:00Z"
  },
  "orchestrator": {
    "pid": 12346,
    "addr": "http://127.0.0.1:51235",
    "db": "/Users/me/.tasq/system/data/orchestrator.sqlite",
    "started_at": "2026-06-01T10:00:01Z"
  }
}
```

`tq` and `tasq-tui` read `issue_tracker.addr` when no API URL is provided.

## Resolution Order

Issue-tracker API URL:

```text
--api-url / -api flag > TQ_API_URL > state.json issue_tracker.addr > http://localhost:8080
```

Comment author:

```text
--author flag > TQ_AUTHOR > config.yaml author > $USER
```

Orchestrator concurrency:

```text
effective max = min(WORKFLOW.md agent.max_concurrent_agents, config.yaml max_concurrent_agents)
```

`WORKFLOW.md` stays in each project repository. `$TQ_HOME/config/config.yaml` stores machine-wide preferences and limits.
