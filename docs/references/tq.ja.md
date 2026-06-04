# tq コマンドリファレンス

`tq` は issue-tracker API の command-line client です。raw HTTP request を直接扱わずに、agent、workflow tool、local development command から issue の作成、参照、更新、コメント追加を行うために使います。

## 実行方法

ローカルの dev container 開発環境では Makefile target を使います。

```sh
make run-tq ARGS="issue list"
```

この target は service process を起動、停止、再起動せず、起動済み dev container 内で installed `tq` binary を実行します。Default workflow では、`tq` は `$TQ_HOME/system/state.json` から issue-tracker API を解決します。

Host-only workflow では直接 `tq` を実行することもできます。

```sh
TQ_HOME=./.tasq go run ./cmd/tq --api-url http://localhost:37651 issue list
```

## Global Options

```text
tq [--api-url URL] [--output text|json] <resource|command> <action> [flags]
```

| Option | Default | Description |
|---|---|---|
| `--api-url URL` | `TQ_API_URL`、その後 `$TQ_HOME/system/state.json`、その後 `http://localhost:37651` | issue-tracker API の base URL。 |
| `--output text\|json` | `text` | Output format。JSON output は script や agent 向けです。 |

## Resources

| Resource | Actions |
|---|---|
| `issue` | `create`, `get`, `list`, `update` |
| `comment` | `add`, `list` |
| `project` | `add`, `remove`, `check`, `list` |
| `service` | `start`, `stop`, `status` |
| `version` | version information を表示 |

## Version

`tq` の version と build commit を表示します。

```sh
tq version
```

Versioned module または GitHub Release から install した release build は、Go build metadata に含まれる tag version を表示します。Local build は `dev` に fallback します。

## Issues

### `issue list`

Issue を一覧表示します。`--project KEY` を渡すと 1 つの project に絞り込めます。

```sh
make run-tq ARGS="issue list"
```

```sh
make run-tq ARGS="issue list --project tasq"
```

JSON output:

```sh
make run-tq ARGS="--output json issue list"
```

### `issue get`

Numeric ID で issue を 1 件取得します。

```sh
make run-tq ARGS="issue get 1"
```

### `issue create`

Issue を作成します。`--project` は必須で、既存の project key を指定する必要があります。

```sh
make run-tq ARGS='issue create --project tasq --title "Write tq reference"'
```

Flags:

| Flag | Required | Description |
|---|---:|---|
| `--project KEY` | yes | Issue を所有する project key。 |
| `--title TITLE` | yes | Issue title。 |
| `--description TEXT` | no | Issue description。 |
| `--status STATUS` | no | Issue status。省略時は `backlog` です。 |
| `--priority PRIORITY` | no | Issue priority。省略時は `normal` です。 |
| `--assignee NAME` | no | Assignee name。 |
| `--attach PATH` | no | PNG、JPEG、GIF、WebP image を upload し、description に Markdown image reference を追記します。 |

Example:

```sh
make run-tq ARGS='issue create --project tasq --title "Improve project list" --description "Render project list as a readable table." --status ready --priority high --assignee codex'
```

### `issue update`

Issue の field を 1 つ以上更新します。

```sh
make run-tq ARGS='issue update 1 --status in_progress'
```

少なくとも 1 つの update flag が必要です。

Flags:

| Flag | Description |
|---|---|
| `--title TITLE` | Issue title を置き換えます。 |
| `--description TEXT` | Issue description を置き換えます。 |
| `--status STATUS` | Issue status を置き換えます。 |
| `--priority PRIORITY` | Issue priority を置き換えます。 |
| `--assignee NAME` | Issue assignee を置き換えます。 |
| `--attach PATH` | PNG、JPEG、GIF、WebP image を upload し、description に Markdown image reference を追記します。 |

Attachment reference は `![filename](attachment://<id>)` を使います。issue-tracker は attachment content API で画像を配信し、Web UI は Markdown から画像を表示します。

## Comments

### `comment add`

Issue に comment を追加します。

```sh
make run-tq ARGS='comment add 1 --body "Started implementation."'
```

Flags:

| Flag | Required | Description |
|---|---:|---|
| `--body TEXT` | yes | Comment body。 |
| `--author NAME` | no | Comment author。省略時は `TQ_AUTHOR`、その後 `USER` を使います。 |
| `--type TYPE` | no | Comment type。省略時は `general` です。 |
| `--attach PATH` | no | PNG、JPEG、GIF、WebP image を upload し、comment body に Markdown image reference を追記します。 |

### `comment list`

Issue の comment を一覧表示します。

```sh
make run-tq ARGS="comment list 1"
```

## Services

### `service start`

issue-tracker と orchestrator を host-local background process として起動します。この command は issue-tracker を先に起動し、health endpoint を待ってから orchestrator を起動します。Log は `$TQ_HOME/system/log/` 配下へ追記されます。

```sh
TQ_HOME=./.tasq go run ./cmd/tq service start
```

Default service ports:

| Service | Port | Log |
|---|---:|---|
| issue-tracker | `37651` | `$TQ_HOME/system/log/issue-tracker.log` |
| orchestrator | `37652` | `$TQ_HOME/system/log/orchestrator.log` |

### `service status`

service state、PID、port、uptime を表示します。Script 向けに JSON output も使えます。

```sh
TQ_HOME=./.tasq go run ./cmd/tq service status
```

```sh
TQ_HOME=./.tasq go run ./cmd/tq --output json service status
```

### `service stop`

orchestrator を先に停止し、その後 issue-tracker を停止します。各 process へ `SIGTERM` を送り、grace period 内に終了しない場合は kill します。

```sh
TQ_HOME=./.tasq go run ./cmd/tq service stop
```

## Projects

### `project add`

Local repository を project として登録します。

```sh
make run-tq ARGS="project add ."
```

`project add` はデフォルトで current directory を使います。path を host-local absolute path に解決し、local に存在することを確認してから issue-tracker API に送信します。

Flags:

| Flag | Description |
|---|---|
| `--key KEY` | Project key。省略時は project directory name から kebab-case key を生成します。 |

Examples:

```sh
make run-tq ARGS='project add --key tasq .'
make run-tq ARGS='project add ../another-project'
```

### `project list`

登録済み project を一覧表示します。

```sh
make run-tq ARGS="project list"
```

Machine-readable output には JSON を使います。

```sh
make run-tq ARGS="--output json project list"
```

### `project check`

Local project の workflow files を check します。

```sh
make run-tq ARGS="project check"
make run-tq ARGS="project check tasq"
```

Project key を指定しない場合、`project check` は current directory に登録された project を探します。

### `project remove`

Project key で project を削除します。

```sh
make run-tq ARGS="project remove tasq"
```

## Valid Values

### Issue Status

```text
backlog
ready
in_progress
review
blocked
failed
done
```

### Issue Priority

```text
low
normal
high
urgent
```

### Comment Type

```text
progress
blocker
handoff
general
```

## Path Handling

Project path は host-local absolute path として保存されます。つまり `project add .` は、`/workspace` のような container-only runtime path ではなく、host machine 上でユーザーから見える path を記録します。

Issue-tracker API は project path が absolute path であることを検証しますが、API server filesystem 上に存在するかどうかは検証しません。`tq project add` client が、project record を作成する前に local existence check を行います。
