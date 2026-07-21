# tq コマンドリファレンス

`tq` は issue-tracker API 用のコマンドラインクライアントです。生の HTTP リクエストを直接扱わずに、エージェント、ワークフローツール、ローカル開発用コマンドから課題の作成、確認、更新、コメント追加を行うために使います。

## 実行方法

ローカルの dev container 開発環境では Makefile ターゲットを使います。

```sh
make run-tq ARGS="issue list"
```

このターゲットは、サービスプロセスを起動、停止、再起動せずに、起動済みの dev container 内でインストール済みの `tq` バイナリを実行します。既定のワークフローでは、`tq` は `$TQ_HOME/system/state.json` から issue-tracker API を解決します。

ホストだけで動かすワークフローでは、`tq` を直接実行することもできます。

```sh
TQ_HOME=./.tasq go run ./cmd/tq --api-url http://localhost:37651 issue list
```

## グローバルオプション

```text
tq [--api-url URL] [--output text|json] <resource|command> <action> [flags]
```

| Option | Default | Description |
|---|---|---|
| `--api-url URL` | `TQ_API_URL`、その後 `$TQ_HOME/system/state.json`、その後 `http://localhost:37651` | issue-tracker API のベース URL。 |
| `--output text\|json` | `text` | 出力形式。JSON 出力はスクリプトやエージェント向けです。 |

## リソース

| Resource | Actions |
|---|---|
| `issue` | `create`, `get`, `list`, `update` |
| `comment` | `add`, `list` |
| `project` | `add`, `remove`, `check`, `list` |
| `workflow` | `add`, `remove`, `show` |
| `migrate` | 保留中のマイグレーションを適用、`down`、`status` |
| `service` | `start`, `stop`, `status` |
| `update` | リリースをインストールしてサービスを再起動 |
| `version` | バージョン情報を表示 |

## バージョン

`tq` のバージョンとビルドコミットを表示します。

```sh
tq version
```

バージョン付きモジュールまたは GitHub Release からインストールしたリリースビルドは、Go のビルドメタデータに含まれるタグバージョンを表示します。ローカルビルドでは `dev` にフォールバックします。

## 更新

GitHub Release から `tq` と同じディレクトリに置くサービス実行ファイルをインストールし、ローカル DB のマイグレーションを適用して、ローカルサービスを再起動します。

```sh
tq update
```

このコマンドは、サービスを停止する前に現在のバージョンと更新先リリースを表示します。更新中はローカルサービスの停止と再起動が入るため、既定では確認を求めます。確認なしで進める場合は `-y` を渡します。

```sh
tq update -y
```

既定では最新の正式リリースをインストールします。特定の正式リリースまたは prerelease tag をインストールする場合は `--tag` を渡します。

```sh
tq update --tag v0.2.0-rc.1
```

更新フローは、サービス停止、固定のユーザーインストール先への release artifacts インストール、新しくインストールした `tq version` の確認、マイグレーション適用、サービス起動の順に進みます。いずれかの工程が失敗した場合、後続の工程は実行されません。

## 課題

### `issue list`

課題を一覧表示します。`--project KEY` を渡すと、結果を 1 つのプロジェクトに絞り込めます。

```sh
make run-tq ARGS="issue list"
```

```sh
make run-tq ARGS="issue list --project tasq"
```

JSON 出力:

```sh
make run-tq ARGS="--output json issue list"
```

### `issue get`

数値 ID で課題を 1 件取得します。

```sh
make run-tq ARGS="issue get 1"
```

### `issue create`

課題を作成します。`--project` は必須で、既存のプロジェクトキーを指定する必要があります。

```sh
make run-tq ARGS='issue create --project tasq --title "Write tq reference"'
```

フラグ:

| Flag | Required | Description |
|---|---:|---|
| `--project KEY` | yes | 課題を所有するプロジェクトキー。 |
| `--title TITLE` | yes | 課題のタイトル。 |
| `--description TEXT` | no | 課題の説明。 |
| `--status STATUS` | no | 課題のステータス。省略時は `backlog` です。 |
| `--priority PRIORITY` | no | 課題の優先度。省略時は `normal` です。 |
| `--assignee NAME` | no | 担当者名。 |
| `--dependency IDS` | no | カンマ区切りの課題 ID で依存関係を設定します。空の値は拒否されます。 |
| `--attach PATH` | no | PNG、JPEG、GIF、WebP 画像をアップロードし、説明に Markdown 画像参照を追記します。 |

例:

```sh
make run-tq ARGS='issue create --project tasq --title "Improve project list" --description "Render project list as a readable table." --status ready --priority high --assignee codex'
```

### `issue update`

課題のフィールドを 1 つ以上更新します。

```sh
make run-tq ARGS='issue update 1 --status in_progress'
```

少なくとも 1 つの更新フラグが必要です。

フラグ:

| Flag | Description |
|---|---|
| `--title TITLE` | 課題のタイトルを置き換えます。 |
| `--description TEXT` | 課題の説明を置き換えます。 |
| `--status STATUS` | 課題のステータスを置き換えます。 |
| `--changed-reason REASON` | 最新 status change reason を記録します。refine work では `refine_requested` のような安定した値を使います。 |
| `--changed-by NAME` | `--status` が issue status を変更したとき、status event に記録する actor。省略時は comment author fallback を使います。 |
| `--priority PRIORITY` | 課題の優先度を置き換えます。 |
| `--assignee NAME` | 課題の担当者を置き換えます。 |
| `--dependency IDS` | カンマ区切りの課題 ID で依存関係全体を置き換えます。空の値は拒否されます。すべての依存関係を削除する場合は `--clear-dependencies` を使います。 |
| `--clear-dependencies` | すべての依存関係を削除します。`--dependency` と同時には指定できません。 |
| `--attach PATH` | PNG、JPEG、GIF、WebP 画像をアップロードし、説明に Markdown 画像参照を追記します。 |

添付参照は `![filename](attachment://<id>)` の形式です。issue-tracker は attachment content API で画像を配信し、Web UI は Markdown から画像を表示します。

follow-up work のために issue を `ready` へ戻す場合は `--changed-reason` を使います。

```sh
make run-tq ARGS='issue ready 1 --changed-reason refine_requested --changed-by codex'
```

## コメント

### `comment add`

課題にコメントを追加します。

```sh
make run-tq ARGS='comment add 1 --body "Started implementation."'
```

フラグ:

| Flag | Required | Description |
|---|---:|---|
| `--body TEXT` | yes | コメント本文。 |
| `--author NAME` | no | コメントの作成者。省略時は `TQ_AUTHOR`、その後 `USER` を使います。 |
| `--type TYPE` | no | コメント種別。省略時は `general` です。 |
| `--attach PATH` | no | PNG、JPEG、GIF、WebP 画像をアップロードし、コメント本文に Markdown 画像参照を追記します。 |

### `comment list`

課題のコメントを一覧表示します。

```sh
make run-tq ARGS="comment list 1"
```

## サービス

### `service start`

issue-tracker、orchestrator、web をホストローカルのバックグラウンドプロセスとして起動します。サービスプロセスを起動する前に、ローカルの issue-tracker と orchestrator のデータベースを開き、保留中のマイグレーションがないか確認します。保留中のマイグレーションがある場合は、`tq migrate` の実行を促してすぐに終了します。保留中のものがなければ issue-tracker を先に起動し、health endpoint を待ってから orchestrator と web を起動します。ログは `$TQ_HOME/system/log/` 配下へ追記されます。

```sh
TQ_HOME=./.tasq go run ./cmd/tq service start
```

既定のサービスポート:

| Service | Port | Log |
|---|---:|---|
| issue-tracker | `37651` | `$TQ_HOME/system/log/issue-tracker.log` |
| orchestrator | `37652` | `$TQ_HOME/system/log/orchestrator.log` |
| web | `37653` | `$TQ_HOME/system/log/web.log` |

### `service status`

サービスの状態、PID、ポート、稼働時間を表示します。スクリプト向けに JSON 出力も使えます。

```sh
TQ_HOME=./.tasq go run ./cmd/tq service status
```

```sh
TQ_HOME=./.tasq go run ./cmd/tq --output json service status
```

### `service stop`

orchestrator を先に停止し、その後 issue-tracker を停止します。各プロセスに `SIGTERM` を送り、猶予期間内に終了しない場合は強制終了します。

```sh
TQ_HOME=./.tasq go run ./cmd/tq service stop
```

## プロジェクト

### `project add`

ローカルリポジトリをプロジェクトとして登録します。

```sh
make run-tq ARGS="project add ."
```

`project add` は既定で現在のディレクトリを使います。パスをホストローカルの絶対パスに解決し、ローカルに存在することを確認してから issue-tracker API に送信します。

フラグ:

| Flag | Description |
|---|---|
| `--key KEY` | プロジェクトキー。省略時はプロジェクトディレクトリ名から kebab-case のキーを生成します。 |

例:

```sh
make run-tq ARGS='project add --key tasq .'
make run-tq ARGS='project add ../another-project'
```

### `project list`

登録済みプロジェクトを一覧表示します。

```sh
make run-tq ARGS="project list"
```

機械判読可能な出力には JSON を使います。

```sh
make run-tq ARGS="--output json project list"
```

### `project check`

ローカルプロジェクトのワークフローファイルを確認します。

```sh
make run-tq ARGS="project check"
make run-tq ARGS="project check tasq"
```

プロジェクトキーを指定しない場合、`project check` は現在のディレクトリに登録されたプロジェクトを探します。

### `project remove`

プロジェクトキーを指定してプロジェクトを削除します。既定では、`project remove` は取り消し不可の操作である警告と、削除対象になるプロジェクトおよび子孫データを表示し、削除を開始する前に正確なプロジェクトキーの入力を求めます。削除では、プロジェクトと、課題、コメント、添付ファイル、ワークフロー上書き、run data などの子孫データが削除されます。

```sh
make run-tq ARGS="project remove tasq"
```

agent や script から使う場合は、`-y` でプロンプトをスキップできます。

```sh
make run-tq ARGS="project remove -y tasq"
```

プロジェクトに実行中の run がある場合、削除前にコマンドは失敗し、API が返した理由を表示します。

## ワークフロー

### `workflow add`

プロジェクトのデータベース側ワークフロー上書きを追加または置き換えます。

```sh
make run-tq ARGS="workflow add --project tasq --file WORKFLOW.md"
```

### `workflow remove`

プロジェクトのデータベース側ワークフロー上書きを削除します。削除後のワークフロー解決は、プロジェクトの `WORKFLOW.md` ファイルまたはグローバルワークフローのフォールバックに戻ります。

```sh
make run-tq ARGS="workflow remove --project tasq"
```

### `workflow show`

プロジェクトで解決された `WORKFLOW.md` の内容を表示します。

```sh
make run-tq ARGS="workflow show --project tasq"
```

このコマンドは、ワークフロー解決と同じ参照順序を使います。

1. 登録済みプロジェクトの場所にある `WORKFLOW.md`。
2. issue-tracker API に保存されたプロジェクトワークフロー。
3. グローバルな `$TQ_HOME/WORKFLOW.md`。

テキスト出力では、`# Source: ...` ヘッダーに続けて解決済みの `WORKFLOW.md` の内容を出力します。構造化された出力には `--json` またはグローバルの `--output json` を使います。

```sh
make run-tq ARGS="workflow show --project tasq --json"
```

## マイグレーション

### `migrate`

`$TQ_HOME` 配下にあるローカルの issue-tracker と orchestrator のデータベースに、保留中の SQLite マイグレーションをすべて適用します。

```sh
make run-tq ARGS="migrate"
```

このコマンドはサービスを起動せずに実行でき、各データベースの `schema_migrations` テーブルにマイグレーション状態を記録します。

### `migrate down`

ローカルデータベースごとに、適用済みマイグレーションを 1 つずつロールバックします。

```sh
make run-tq ARGS="migrate down"
```

### `migrate status`

ローカルデータベースごとに、適用済みおよび保留中のマイグレーションを一覧表示します。

```sh
make run-tq ARGS="migrate status"
```

スクリプトでは JSON 出力を使えます。

```sh
make run-tq ARGS="--output json migrate status"
```

## 有効な値

### 課題ステータス

```text
backlog
ready
in_progress
review
blocked
failed
done
```

### 課題優先度

```text
low
normal
high
urgent
```

### コメント種別

```text
progress
blocker
handoff
general
```

## パスの扱い

プロジェクトパスは、ホストローカルの絶対パスとして保存されます。つまり `project add .` は、`/workspace` のようなコンテナ内だけの実行時パスではなく、ホストマシン上でユーザーから見えるパスを記録します。

issue-tracker API は、プロジェクトパスが絶対パスであることを検証しますが、API サーバーのファイルシステム上に存在するかどうかは検証しません。`tq project add` クライアントが、プロジェクトレコードを作成する前にローカルで存在確認を行います。
