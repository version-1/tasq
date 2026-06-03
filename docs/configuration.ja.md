# Configuration

English counterpart: [configuration.md](configuration.md).

Tasq は machine-level configuration、runtime state、service data の置き場所として `TQ_HOME` を使います。

default では `TQ_HOME` は `~/.tasq` に解決されます。development では repository local の directory を指定できます。

```sh
TQ_HOME=./.tasq
```

Default の Compose development workflow は、tool を `dev` container 内で実行し、
`TQ_HOME=/workspace/.tasq` を使います。`tq`、TUI、issue-tracker、orchestrator は同じ
container 内の runtime state を読みます。

Codex credential は `TQ_HOME` とは分離します。dev container では
`CODEX_HOME=/home/codex/.codex` を使い、`codex-home` named volume に保存します。
Container 内で一度 `make dev-codex-login` を実行し、device auth で認証します。`codex-home`
volume を削除すると login state も削除されます。Device auth を使うことで、container 内にだけ
存在する localhost callback へ browser が redirect して失敗する問題を避けます。

Repository-managed な Codex rules は `codex/rules/` に置き、dev container 内の
`/home/codex/.codex/rules` へ read-only mount します。Authentication、personal override、
generated approval decision、その他 secret-bearing な Codex state は repository ではなく
`codex-home` volume に保持します。

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

`config/` は user-editable です。`system/` は Tasq processes が管理し、上書きされる可能性があります。
Development service logs は `system/log/` 配下に書き込まれます。`tq service start` は issue-tracker と orchestrator の log をこの directory へ追記します。

## config.yaml

```yaml
author: "jiro"
max_concurrent_agents: 3
```

| Field | Default | Description |
|---|---:|---|
| `author` | `$USER` | `--author` と `TQ_AUTHOR` が未指定のときに `tq comment add` が使う default author です。 |
| `max_concurrent_agents` | `10` | orchestrator agent runs の machine-wide concurrency limit です。 |

## state.json

running service は discovery metadata を `system/state.json` に書き込みます。

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

`tq` と `tasq-tui` は API URL が指定されていない場合に `issue_tracker.addr` を読みます。

## Resolution Order

issue-tracker API URL:

```text
--api-url / -api flag > TQ_API_URL > state.json issue_tracker.addr > http://localhost:37651
```

comment author:

```text
--author flag > TQ_AUTHOR > config.yaml author > $USER
```

orchestrator concurrency:

```text
effective max = min(WORKFLOW.md agent.max_concurrent_agents, config.yaml max_concurrent_agents)
```

`WORKFLOW.md` は各 project repository に残します。`$TQ_HOME/config/config.yaml` は machine-wide preferences と limits を保存します。
