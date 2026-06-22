# Configuration

English counterpart: [configuration.md](configuration.md).

Tasq は、マシン単位の設定、実行時状態、サービスデータを置くローカルホームディレクトリとして `TQ_HOME` を使います。

既定では、`TQ_HOME` は `~/.tasq` に解決されます。開発時は、リポジトリ内のローカルディレクトリを指定できます。

```sh
TQ_HOME=./.tasq
```

既定の Compose 開発ワークフローでは、`dev` container 内でツールを実行し、
`TQ_HOME=/workspace/.tasq` を使います。`tq`、TUI、issue-tracker、orchestrator はすべて、その container 内の同じ実行時状態を読みます。

Codex の認証情報は `TQ_HOME` とは分離します。dev container では
`CODEX_HOME=/home/codex/.codex` を使い、`codex-home` named volume に保存します。
container 内で一度 `make dev-codex-login` を実行し、device auth で認証します。`codex-home`
volume を削除するとログイン状態も削除されます。device auth を使うことで、container 内にだけ存在する localhost callback へブラウザがリダイレクトして失敗する問題を避けます。

リポジトリ管理の Codex rules は `codex/rules/` に置き、dev container 内の
`/home/codex/.codex/rules` へ読み取り専用でマウントします。認証情報、個人用の上書き設定、
生成された承認判断、その他の秘密情報を含む Codex state は、リポジトリではなく
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

`config/` はユーザーが編集できます。`system/` は Tasq のプロセスが管理し、上書きされる可能性があります。
開発用サービスログは `system/log/` 配下に書き込まれます。`tq service start` は issue-tracker と orchestrator のログをこのディレクトリへ追記します。

## config.yaml

```yaml
author: "jiro"
max_concurrent_agents: 3
```

| Field | Default | Description |
|---|---:|---|
| `author` | `$USER` | `--author` と `TQ_AUTHOR` が未指定のときに `tq comment add` が使う既定の author です。 |
| `max_concurrent_agents` | `10` | orchestrator agent runs に対するマシン全体の同時実行数の上限です。 |

## state.json

実行中のサービスは、検出用メタデータを `system/state.json` に書き込みます。

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

API URL が指定されていない場合、`tq` と `tasq-tui` は `issue_tracker.addr` を読みます。

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

`WORKFLOW.md` は各プロジェクトリポジトリに残します。`$TQ_HOME/config/config.yaml` はマシン全体の設定と上限値を保存します。
