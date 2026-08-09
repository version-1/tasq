# tq コマンドリファレンス

`tq` は Tasq のコマンドラインクライアントです。Issue Tracker の一般的な操作に対応する型付きコマンド、ローカルサービスとマイグレーションの管理コマンド、許可リストで制限された生の API コマンドを提供します。

## 実行方法

ローカルの開発コンテナでは Makefile ターゲットを使います。

```sh
make run-tq ARGS="issue list"
```

このターゲットは、サービスプロセスを起動、停止、再起動せずに、起動済みの開発コンテナ内でインストール済みの `tq` バイナリを実行します。既定のワークフローでは、`tq` は `$TQ_HOME/system/state.json` から Issue Tracker API の接続先を解決します。

ホストだけで動かすワークフローでは、`tq` を直接実行することもできます。

```sh
TQ_HOME=./.tasq go run ./cmd/tq --api-url http://localhost:37651 issue list
```

## グローバルオプション

```text
tq [--api-url URL] [--output text|json] <command> [args] [flags]
```

| オプション | 既定値 | 説明 |
|---|---|---|
| `--api-url URL` | `TQ_API_URL`、その後 `$TQ_HOME/system/state.json`、その後 `http://localhost:37651` | Issue Tracker API のベース URL。 |
| `--output text\|json` | `text` | 出力形式。JSON 出力はスクリプトやエージェント向けです。 |

## コマンド

| コマンド | 操作または用途 |
|---|---|
| `issue` | `create`、`get`、`list`、`watch`、`update`、`close`、`cancel`、`ready`、`draft`、`rename`、`edit` |
| `artifact` | `set`、`delete` |
| `comment` | `add`、`list` |
| `project` | `add`、`remove`、`check`、`list` |
| `workflow` | `add`、`remove`、`show` |
| `migrate` | 未適用マイグレーションの適用、`down` によるロールバック、`status` による状態確認 |
| `service` | `start`、`stop`、`status` |
| `logs` | サービスログの表示または追跡 |
| `web` | 実行中の Web UI を開く |
| `config` | ビルド、ホーム、解決済みの設定情報を表示 |
| `update` | リリースをインストールしてサービスを再起動 |
| `version` | バージョン情報を表示 |
| `api` | 許可リストにある Issue Tracker API へ生のリクエストを送信 |

## 生の API リクエスト

Issue Tracker の操作が型付きコマンドとして提供されていない場合に、`tq api` を使用します。このコマンドは、解決済みの Issue Tracker ベース URL に生のリクエストを送信します。汎用 HTTP クライアントではありません。API の意味については [Tasq API](../design/api.ja.md)、正式なエンドポイント契約については [OpenAPI 文書](../openapi/issue-tracker.yml)を参照してください。

```sh
tq api GET /api/v1/issues --query states=ready

tq api POST /api/v1/issues --header 'X-Request-ID: local-123' --data @request.json
```

構文は次のとおりです。

```text
tq api <method> <path> [--query key=value] [--header 'Name: value'] [--data value|@file|-]
```

HTTP メソッドは大文字と小文字を区別せず、内部で大文字に正規化します。パスには、エンコードされていない `/api/v1/...` 形式の絶対パスを指定します。完全 URL、フラグメント、ドットセグメント、空のセグメント、末尾のスラッシュは使用できません。パスに直接記述したクエリは保持し、繰り返し指定した `--query key=value` は指定順で追加します。クエリの名前と値は意味を検証せず、API に渡します。

HTTP メソッドとパスは、CLI が明示的に保持する現行 Issue Tracker ルートの許可リストに一致する必要があります。数値 ID は正の `int64` に限定します。許可リストは閉じた状態を既定とする設計であり、サーバーにルートを追加しても、CLI の許可リストを更新するまでは使用できません。Artifact の `PUT` / `DELETE` ルートは、型付きコマンドと合わせて許可リストに追加します。生の multipart をまだ扱えないため、`POST /api/v1/attachments` は一時的に除外しています。添付ファイルに対する `PATCH` も許可しません。

`--header` は繰り返し指定できます。ヘッダー名は大文字と小文字を区別せず、同名の場合は最後の値を使います。`Host`、`Content-Length`、`Transfer-Encoding`、`Connection`、`Trailer`、`Upgrade`、`Proxy-Connection` など、HTTP トランスポートが管理するヘッダーは指定できません。

`--data` には、リテラル値、`@file`、または標準入力を示す `-` を指定できます。使用できる HTTP メソッドは `POST`、`PUT`、`PATCH` です。本文が JSON として妥当かどうかは検証しません。本文があり、`Content-Type` ヘッダーを明示しなかった場合は `application/json` を使用します。

書き込みや削除操作でも確認は求めません。リダイレクトは追跡せず、HTTP のタイムアウトは 10 秒です。バイナリデータや HTTP エラーの本文も含め、レスポンスのバイト列を変更せずに標準出力へ書き出します。`--output` を指定しても変換しません。終了ステータスは、HTTP `2xx` が `0`、HTTP `3xx`～`5xx` または通信失敗が `1`、使用方法・入力・許可リストのエラーが `2` です。

## バージョン

`tq` のバージョンとビルドコミットを表示します。

```sh
tq version
```

バージョン付きモジュールまたは GitHub Release からインストールしたリリースビルドは、Go のビルドメタデータに含まれるタグバージョンを表示します。ローカルビルドでは `dev` にフォールバックします。

## 設定

`tq config` は、バージョン、ビルドプロファイル、`TQ_HOME` の上書き値、解決済みのホームディレクトリ、設定ファイルのパス、解決済みの設定値を表示します。機械可読な出力には `--output json` を指定します。設定ファイルの YAML はそのまま表示しません。

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

空でないビルドプロファイルを持つバイナリでは、汎用のリリース成果物がそのプロファイルを保持しないため、`tq update` は利用できません。

```sh
tq update --tag v0.2.0-rc.1
```

更新処理は、サービス停止、固定のユーザーインストール先へのリリース成果物のインストール、新しくインストールした `tq version` の確認、マイグレーション適用、サービス起動の順に進みます。いずれかの工程が失敗した場合、後続の工程は実行されません。

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

| フラグ | 必須 | 説明 |
|---|---:|---|
| `--project KEY` | はい | 課題を所有するプロジェクトキー。 |
| `--title TITLE` | はい | 課題のタイトル。 |
| `--description TEXT` | いいえ | 課題の説明。 |
| `--status STATUS` | いいえ | 課題のステータス。省略時は `backlog` です。 |
| `--priority PRIORITY` | いいえ | 課題の優先度。省略時は `normal` です。 |
| `--assignee NAME` | いいえ | 担当者名。 |
| `--dependency IDS` | いいえ | カンマ区切りの課題 ID で依存関係を設定します。空の値は拒否されます。 |
| `--attach PATH` | いいえ | PNG、JPEG、GIF、WebP 画像をアップロードし、説明に Markdown 画像参照を追記します。 |

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

| フラグ | 説明 |
|---|---|
| `--title TITLE` | 課題のタイトルを置き換えます。 |
| `--description TEXT` | 課題の説明を置き換えます。 |
| `--status STATUS` | 課題のステータスを置き換えます。 |
| `--priority PRIORITY` | 課題の優先度を置き換えます。 |
| `--assignee NAME` | 課題の担当者を置き換えます。 |
| `--dependency IDS` | カンマ区切りの課題 ID で依存関係全体を置き換えます。空の値は拒否されます。すべての依存関係を削除する場合は `--clear-dependencies` を使います。 |
| `--clear-dependencies` | すべての依存関係を削除します。`--dependency` と同時には指定できません。 |
| `--attach PATH` | PNG、JPEG、GIF、WebP 画像をアップロードし、説明に Markdown 画像参照を追記します。 |

添付ファイルは `![filename](attachment://<id>)` の形式で参照します。Issue Tracker は添付ファイル取得 API で画像を配信し、Web UI は Markdown 内の参照を画像として表示します。

### `issue watch`

実行準備ができた課題のキューを定期的に取得し、監視やエージェントへの割り当てに使う JSON オブジェクトを1行ずつ出力します。`event` レコードには新しく検出した課題が入り、一時的な API エラーはループを停止せず `error` レコードとして出力します。このコマンドは常に JSON Lines 形式を使い、グローバルの `--output` 設定を無視します。

| フラグ | 既定値 | 説明 |
|---|---:|---|
| `--interval SECONDS` | `30` | 取得間隔。正の値を指定します。 |
| `--seen-ttl SECONDS` | `900` | 同じ課題の再出力を抑止する期間。`--interval` より大きい値を指定します。 |
| `--verbose` | 無効 | 取得状況を `info` レコードとして追加出力します。 |

### 課題操作の短縮コマンド

次のコマンドは、`issue update` のフラグを指定せずに1つの項目を更新します。

```text
tq issue close <id>
tq issue cancel <id>
tq issue ready <id>
tq issue draft <id>
tq issue rename <id> <title>
tq issue edit <id> <description>
```

状態変更の短縮コマンドは、順に `done`、`cancelled`、`ready`、`backlog` を設定します。

## Artifact

Artifact は、課題に外部成果物を関連付けます。初期対応の type は `pull_request` で、1 つの課題と type の組み合わせにつき 1 件の URL を保存します。

### `artifact set`

Artifact を作成または更新します。課題 ID は正の値でなければならず、`--type` は必須です。

```sh
tq artifact set 14 --type pull_request https://github.com/example/tasq/pull/42
```

URL は前後の空白を除去した後、host を持つ絶対 `http` / `https` URL で、userinfo を含まないことが必要です。UTF-8 で 4,096 bytes を超える値は拒否されます。同じ課題と type に対してこのコマンドを繰り返すと、Artifact の元の作成時刻を維持したまま URL を置き換えます。

### `artifact delete`

課題の Artifact を削除します。課題 ID は正の値でなければならず、`--type` は必須です。

```sh
tq artifact delete 14 --type pull_request
```

両コマンドはグローバルな text / JSON 出力の規約に従います。課題が不明な場合や Artifact が存在しない場合は対応する `404` エラーを返し、未対応の type や不正な URL は成功リクエストの前に拒否します。

## コメント

### `comment add`

課題にコメントを追加します。

```sh
make run-tq ARGS='comment add 1 --body "Started implementation."'
```

フラグ:

| フラグ | 必須 | 説明 |
|---|---:|---|
| `--body TEXT` | はい | コメント本文。 |
| `--author NAME` | いいえ | コメントの作成者。省略時は `TQ_AUTHOR`、その後 `USER` を使います。 |
| `--type TYPE` | いいえ | コメント種別。省略時は `general` です。 |
| `--attach PATH` | いいえ | PNG、JPEG、GIF、WebP 画像をアップロードし、コメント本文に Markdown 画像参照を追記します。 |

### `comment list`

課題のコメントを一覧表示します。

```sh
make run-tq ARGS="comment list 1"
```

## サービス

### `service start`

Issue Tracker、オーケストレーター、Web UI をホスト上のバックグラウンドプロセスとして起動します。サービスプロセスを起動する前に、ローカルの Issue Tracker とオーケストレーターのデータベースを開き、未適用のマイグレーションがないか確認します。未適用のマイグレーションがある場合は、`tq migrate` の実行を促してすぐに終了します。なければ Issue Tracker を先に起動し、ヘルスチェックの成功を待ってからオーケストレーターと Web UI を起動します。ログは `$TQ_HOME/system/log/` 以下へ追記します。

```sh
TQ_HOME=./.tasq go run ./cmd/tq service start
```

既定のサービスポート:

| サービス | ポート | ログ |
|---|---:|---|
| issue-tracker | `37651` | `$TQ_HOME/system/log/issue-tracker.log` |
| orchestrator | `37652` | `$TQ_HOME/system/log/orchestrator.log` |
| web | `37653` | `$TQ_HOME/system/log/web.log` |

既定ポートのいずれかが使用中の場合、`tq service start` は全サービスに異なる OS 選択の loopback ポートを提案し、確認を求めます。`y` または `yes` を入力すると続行し、拒否または非対話入力ではサービスを起動しません。確認を省略して提案ポートを受け入れるには `tq service start -y` を使います。Tasq は確認後に提案ポートを再確認し、ポートが取得されていた場合は別の組を選ばず失敗します。

### `service status`

サービスの状態、PID、ポート、稼働時間を表示します。スクリプト向けに JSON 出力も使えます。

```sh
TQ_HOME=./.tasq go run ./cmd/tq service status
```

```sh
TQ_HOME=./.tasq go run ./cmd/tq --output json service status
```

### `service stop`

オーケストレーターを先に停止し、その後 Issue Tracker を停止します。各プロセスに `SIGTERM` を送り、猶予期間内に終了しない場合は強制終了します。

```sh
TQ_HOME=./.tasq go run ./cmd/tq service stop
```

## ログと Web UI

サービスログの末尾1,000行を表示します。

```sh
tq logs issue-tracker
tq logs orchestrator
tq logs web
```

表示行数は `-n LINES`、追記内容の継続表示は `-f` で指定します。`tracker` は `issue-tracker` の別名です。このコマンドは `$TQ_HOME/system/log/` 以下のファイルを読み取り、`--output json` には対応しません。

実行中の Web UI を既定のブラウザーで開きます。

```sh
tq web
```

Web UI の URL はローカルサービスの状態から取得します。Web UI が起動していない場合、コマンドは失敗します。

## 実験的な端末コンソール

次のどのコマンドでも、正式な閲覧専用端末 UI を起動できます。

```sh
tq tui
tq console
tq c
```

コマンドの後に `--orchestrator-url URL` を指定すると、サービス状態に保存されたアドレスより優先されます。出力は `--output text` だけに対応し、入出力の両方に TTY が必要です。すべてのプロジェクトの課題を一覧表示し、プロジェクト、複数状態、キーワードをサーバー側で絞り込みます。一度に50件ずつ取得し、一覧は15秒ごと、選択中の課題は5秒ごとに更新します。

横幅が広い場合は課題一覧と詳細を同時に表示します。狭い端末では Enter で詳細を開き、Esc で一覧へ戻ります。詳細タブは Overview、Comments、Artifacts、Run です。`q` または Ctrl+C で終了し、矢印または `j`/`k` で移動します。Tab でタブを切り替え、`/` で検索し、`f` で絞り込みを開き、`r` で更新し、`?` でヘルプを開きます。`o` では HTTP(S) の Artifact URL を開き、PgUp では古いコメントを読み込みます。フッターには常に実験的な機能であることを表示します。

コンソールが送信する HTTP 要求は GET だけです。Issue Tracker の障害時は再試行画面を表示します。オーケストレーターが未設定、利用不能、または不正な応答を返した場合は Run タブだけを縮退表示し、404 は実行データがない状態として区別します。

## プロジェクト

### `project add`

ローカルリポジトリをプロジェクトとして登録します。

```sh
make run-tq ARGS="project add ."
```

`project add` は既定で現在のディレクトリを使います。パスをホストローカルの絶対パスに解決し、ローカルに存在することを確認してから issue-tracker API に送信します。

フラグ:

| フラグ | 説明 |
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

プロジェクトキーを指定してプロジェクトを削除します。既定では、`project remove` は取り消せない操作であることと、削除対象になるプロジェクトおよび子孫データを表示し、削除を開始する前に正確なプロジェクトキーの入力を求めます。削除対象には、プロジェクト、課題、コメント、添付ファイル、ワークフロー上書き、実行データなどが含まれます。

```sh
make run-tq ARGS="project remove tasq"
```

エージェントやスクリプトから使う場合は、`-y` で確認を省略できます。

```sh
make run-tq ARGS="project remove -y tasq"
```

プロジェクトに実行中の処理がある場合、削除前にコマンドは失敗し、API が返した理由を表示します。

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
2. Issue Tracker API に保存されたプロジェクトワークフロー。
3. グローバルな `$TQ_HOME/WORKFLOW.md`。

テキスト出力では、`# Source: ...` ヘッダーに続けて解決済みの `WORKFLOW.md` の内容を出力します。構造化された出力には `--json` またはグローバルの `--output json` を使います。

```sh
make run-tq ARGS="workflow show --project tasq --json"
```

## マイグレーション

### `migrate`

`$TQ_HOME` 以下にあるローカルの Issue Tracker とオーケストレーターのデータベースに、未適用の SQLite マイグレーションをすべて適用します。

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
cancelled
duplicate
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

Issue Tracker API は、プロジェクトパスが絶対パスであることを検証しますが、API サーバーのファイルシステム上に存在するかどうかは検証しません。`tq project add` が、プロジェクトレコードを作成する前にローカルで存在確認を行います。
