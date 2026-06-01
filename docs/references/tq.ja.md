# tq コマンドリファレンス

`tq` は issue-tracker API の command-line client です。raw HTTP request を直接扱わずに、agent、workflow tool、local development command から issue の作成、参照、更新、コメント追加を行うために使います。

## 実行方法

ローカルの Compose 開発環境では Makefile target を使います。

```sh
make tq ARGS="issue list"
```

この target は必要に応じて Compose の issue-tracker service を起動し、割り当てられた localhost port を解決して、host 上で `go run ./cmd/tq` を実行します。

直接 `tq` を実行することもできます。

```sh
go run ./cmd/tq --api-url http://localhost:8080 issue list
```

## Global Options

```text
tq [--api-url URL] [--output text|json] <resource> <action> [flags]
```

| Option | Default | Description |
|---|---|---|
| `--api-url URL` | `TQ_API_URL`、その後 `http://localhost:8080` | issue-tracker API の base URL。 |
| `--output text\|json` | `text` | Output format。JSON output は script や agent 向けです。 |

## Resources

| Resource | Actions |
|---|---|
| `issue` | `create`, `get`, `list`, `update` |
| `comment` | `add`, `list` |
| `project` | `add`, `remove`, `check`, `list` |

## Issues

### `issue list`

Issue を一覧表示します。

```sh
make tq ARGS="issue list"
```

JSON output:

```sh
make tq ARGS="--output json issue list"
```

### `issue get`

Numeric ID で issue を 1 件取得します。

```sh
make tq ARGS="issue get 1"
```

### `issue create`

Issue を作成します。

```sh
make tq ARGS='issue create --title "Write tq reference"'
```

Flags:

| Flag | Required | Description |
|---|---:|---|
| `--title TITLE` | yes | Issue title。 |
| `--description TEXT` | no | Issue description。 |
| `--status STATUS` | no | Issue status。省略時は `backlog` です。 |
| `--priority PRIORITY` | no | Issue priority。省略時は `normal` です。 |
| `--assignee NAME` | no | Assignee name。 |

Example:

```sh
make tq ARGS='issue create --title "Improve project list" --description "Render project list as a readable table." --status ready --priority high --assignee codex'
```

### `issue update`

Issue の field を 1 つ以上更新します。

```sh
make tq ARGS='issue update 1 --status in_progress'
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

## Comments

### `comment add`

Issue に comment を追加します。

```sh
make tq ARGS='comment add 1 --body "Started implementation."'
```

Flags:

| Flag | Required | Description |
|---|---:|---|
| `--body TEXT` | yes | Comment body。 |
| `--author NAME` | no | Comment author。省略時は `TQ_AUTHOR`、その後 `USER` を使います。 |
| `--type TYPE` | no | Comment type。省略時は `general` です。 |

### `comment list`

Issue の comment を一覧表示します。

```sh
make tq ARGS="comment list 1"
```

## Projects

### `project add`

Local repository を project として登録し、workspace record を作成します。

```sh
make tq ARGS="project add ."
```

`project add` はデフォルトで current directory を使います。path を host-local absolute path に解決し、local に存在することを確認してから issue-tracker API に送信します。

Flags:

| Flag | Description |
|---|---|
| `--key KEY` | Project key。省略時は project directory name から kebab-case key を生成します。 |
| `--workspace-name NAME` | Workspace name。省略時は project directory name を使います。 |

Examples:

```sh
make tq ARGS='project add --key tasq --workspace-name tasq .'
make tq ARGS='project add ../another-project'
```

### `project list`

登録済み project を一覧表示します。

```sh
make tq ARGS="project list"
```

Machine-readable output には JSON を使います。

```sh
make tq ARGS="--output json project list"
```

### `project check`

Local project の workflow files を check します。

```sh
make tq ARGS="project check"
make tq ARGS="project check tasq"
```

Project key を指定しない場合、`project check` は current directory に登録された project を探します。

### `project remove`

Project key で project を削除します。

```sh
make tq ARGS="project remove tasq"
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

Issue-tracker API は project path と workspace path が absolute path であることを検証しますが、API server filesystem 上に存在するかどうかは検証しません。`tq project add` client が、project record を作成する前に local existence check を行います。
