# Configuration

Japanese counterpart: [configuration.ja.md](configuration.ja.md).

For `TQ_HOME`, the directory layout, `config.yaml`, `state.json`, and resolution order, see the [Configuration reference](../site/docs/reference/configuration.md). This document covers only development-specific configuration: build profiles and the Compose dev container.

## Build Profiles

Development binaries can embed a build profile. An empty profile keeps the default home at `~/.tasq`; a profile such as `dev` resolves to `~/.tasq-dev`. Profiles contain only lowercase letters, digits, and hyphens. An explicitly set `TQ_HOME` always overrides the build profile.

The same profile must be embedded in `tq`, `issue-tracker`, `orchestrator`, and `web` so direct service invocation and managed startup use the same runtime state. Profile isolation does not provide concurrent service startup because service ports remain shared.

## Compose Dev Container

The default Compose development workflow runs tools inside the `dev` container with
`TQ_HOME=/workspace/.tasq`. `tq`, TUI, issue-tracker, and orchestrator all read the same runtime
state in that container.

Codex credentials are separate from `TQ_HOME`. The dev container uses
`CODEX_HOME=/home/codex/.codex`, backed by the `codex-home` named volume. Run `make dev-codex-login`
once to authenticate inside the container with device auth. Removing the `codex-home` volume removes
the login state. Device auth avoids browser redirects to a localhost callback that only exists inside
the container.

Repository-managed Codex rules live in `codex/rules/` and are mounted read-only into the dev
container at `/home/codex/.codex/rules`. Authentication, personal overrides, generated approval
decisions, and other secret-bearing Codex state stay in the `codex-home` volume instead of the
repository.
## Directory Layout

```text
$TQ_HOME/
├── config/
│   └── config.yaml
└── system/
    ├── state.json
    ├── log
    │   ├── issue-tracker.log
    │   ├── orchestrator.log
    │   └── web.log
    └── data/
        ├── issues.sqlite
        └── orchestrator.sqlite
```

`config/` is user-editable. `system/` is managed by Tasq processes and may be overwritten.
Development service logs are written under `system/log/`. `tq service start` appends issue-tracker and orchestrator logs to that directory.

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
    "addr": "127.0.0.1:37651",
    "db": "/Users/me/.tasq/system/data/issues.sqlite",
    "started_at": "2026-06-01T10:00:00Z",
    "process_started_at": "2026-06-01T10:00:00Z",
    "executable": "/Users/me/.tasq/system/bin/issue-tracker"
  },
  "orchestrator": {
    "pid": 12346,
    "addr": "http://127.0.0.1:37652",
    "db": "/Users/me/.tasq/system/data/orchestrator.sqlite",
    "started_at": "2026-06-01T10:00:01Z",
    "process_started_at": "2026-06-01T10:00:01Z",
    "executable": "/Users/me/.tasq/system/bin/orchestrator"
  }
}
```

All managed services record the process start time and executable path. `tq service status` verifies the process start time in addition to PID liveness, avoiding dependence on the caller's current working directory or an executable path containing spaces. Legacy entries without a process start time remain PID-validated until their service is restarted; `tq service stop` leaves such entries untouched because it cannot safely identify the target process. On shutdown, a service clears its entry only when the stored identity still matches its own process, so an older process cannot remove a replacement service's discovery state.

`tq` and `tasq-tui` read `issue_tracker.addr` when no API URL is provided.

## Resolution Order

Issue-tracker API URL:

```text
--api-url / -api flag > TQ_API_URL > state.json issue_tracker.addr > http://localhost:37651
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
