---
id: cli-reference
title: CLI Reference
sidebar_position: 1
---

# CLI リファレンス

`tq` は、課題管理、プロジェクト設定、ワークフロー設定、API への直接アクセス、ローカルサービス、ログ、マイグレーション、Web UI、ターミナル UI を操作するコマンドラインインターフェースです。

## 全体の形式

```sh
tq [--api-url URL] [--output text|json] <resource> <action> [flags]
```

| フラグ | 説明 |
| --- | --- |
| `--api-url URL` | Issue Tracker API の URL。環境変数と状態ファイルによる検出を上書きします。 |
| `--output text\|json` | 出力形式。既定値は `text` です。 |

API URL は `--api-url`、`TQ_API_URL`、`$TQ_HOME/system/state.json`、`http://localhost:37651` の順で解決します。

## 課題コマンド

| コマンド | 用途 |
| --- | --- |
| `tq issue list [--project <key>]` | 課題を一覧表示します。プロジェクトを指定して絞り込めます。 |
| `tq issue get <id>` | 1 件の課題を表示します。 |
| `tq issue create --project <key> --title <title>` | 課題を作成します。 |
| `tq issue update <id> [flags]` | 課題のフィールドを更新します。 |
| `tq issue watch [--interval <seconds>] [--seen-ttl <seconds>] [--verbose]` | 実行可能な課題を定期取得し、JSON イベントを出力します。 |
| `tq issue close <id>` | 課題を `done` に移動します。 |
| `tq issue cancel <id>` | 課題を `cancelled` に移動します。 |
| `tq issue ready <id>` | 課題を `ready` に移動します。 |
| `tq issue draft <id>` | 課題を `backlog` に移動します。 |
| `tq issue rename <id> <title>` | タイトルを更新します。 |
| `tq issue edit <id> <description>` | 説明を更新します。 |

`issue create` では `--project` と `--title` が必須です。`--description`、
`--status`、`--priority`、`--assignee`、`--dependency <comma-separated-ids>`、
`--attach` を指定でき、省略時のステータスと優先度はそれぞれ `backlog` と
`normal` です。`issue update` では、少なくとも 1 つの更新フラグが必要です。
同じ更新可能フィールドに加えて、依存関係を置き換える `--dependency` と、
削除する `--clear-dependencies` を指定できます。これらは同時に指定できず、
空の依存関係値は拒否されます。

`--attach` では PNG、JPEG、GIF、WebP の画像を指定でき、`attachment://` の
Markdown 参照を追記します。アップロード後に参照の更新が失敗した場合は、
CLI がアップロード済みの添付ファイルを削除します。

`tq issue watch` はエージェントのループ処理向けです。NDJSON のイベントエンベロープを出力し、実行待ちキューを読み、設定された TTL の間は同じ課題を重複して出力しません。一時的な API エラーが起きても取得を継続します。`--interval` の既定値は 30 秒で、正の値が必要です。`--seen-ttl` の既定値は 900 秒で、`--interval` より大きくなければなりません。グローバルな `--output` は使用しません。

## Artifact コマンド

| コマンド | 用途 |
| --- | --- |
| `tq artifact set <issue-id> --type pull_request <url>` | 課題のプルリクエスト URL を作成または置き換えます。 |
| `tq artifact delete <issue-id> --type pull_request` | 課題のプルリクエスト URL を削除します。 |

どちらのコマンドも正の課題 ID と `--type` を必要とし、グローバルな text / JSON 出力モードに対応します。

現在サポートする種別は `pull_request` だけです。URL はホストを含み userinfo を持たない絶対 `http` または `https` URL で、UTF-8 で 4,096 バイト以下でなければなりません。同じ課題と種別に対して `artifact set` を繰り返すと URL を置き換えます。

## コメントコマンド

| コマンド | 用途 |
| --- | --- |
| `tq comment add <issue-id> --body <body>` | コメントを追加します。 |
| `tq comment list <issue-id>` | 課題のコメントを一覧表示します。 |

`comment add` には、`--type`（既定値は `general`。指定できる値は `progress`、`blocker`、`handoff`、`general`）、`--author`（`TQ_AUTHOR`、設定ファイルの `author`、`USER` の順に解決）、PNG、JPEG、GIF、WebP の画像を指定する `--attach` があります。

## プロジェクトとワークフローのコマンド

| コマンド | 用途 |
| --- | --- |
| `tq project add [path] [--key <key>]` | リポジトリを登録します。 |
| `tq project remove [-y] <key>` | キー入力による確認後にプロジェクトを削除します。`-y` で確認を省略できます。 |
| `tq project check [key]` | プロジェクト設定を検証します。 |
| `tq project list` | 登録済みプロジェクトを一覧表示します。 |
| `tq workflow add --project <key> (--file <path> \| --body <text>)` | ワークフローの上書きを保存します。 |
| `tq workflow remove --project <key>` | 保存済みの上書きを削除します。 |
| `tq workflow show --project <key> [--json]` | 解決済みワークフローを表示します。 |

`project add` はパスをホスト上の絶対パスへ解決し、存在を確認してから登録します。`project remove` はプロジェクトと、その配下の課題、コメント、添付ファイル、ワークフロー上書き、実行データの削除を警告し、`-y` を指定しない限り正確なプロジェクトキーの入力を求めます。実行中の run があるプロジェクトでは失敗します。

`workflow show` は、登録済みプロジェクトの `WORKFLOW.md`、保存済みプロジェクト上書き、`$TQ_HOME/WORKFLOW.md` の順に解決します。

## 実行環境のコマンド

| コマンド | 用途 |
| --- | --- |
| `tq service start` | issue-tracker、orchestrator、Web UI を起動します。 |
| `tq service stop` | local services を停止します。 |
| `tq service status` | service status を表示します。 |
| `tq orchestrator start` | ローカル orchestrator だけを起動します。起動中のローカル Issue Tracker が必要です。 |
| `tq orchestrator stop` | ローカル orchestrator だけを graceful に停止します。 |
| `tq orchestrator status` | ローカル orchestrator の status を表示します。 |
| `tq logs <service> [-n <lines>] [-f]` | service logs を読みます。 |
| `tq migrate` | migrations を適用します。 |
| `tq migrate down` | migrations を rollback します。 |
| `tq migrate status` | migration status を表示します。 |
| `tq web` | 実行中の Web UI を開きます。 |
| `tq tui` | 実験的な読み取り専用ターミナル UI を開きます。別名は `tq console` と `tq c` です。 |
| `tq config` | ビルド、ホームディレクトリ、解決済み設定を表示します。 |
| `tq version` | バージョン情報を出力します。 |
| `tq update [-y] [--tag <tag>]` | リリースをインストールし、データベースを移行してサービスを再起動します。 |

ログ対象には `tracker` または `issue-tracker`、`orchestrator`、`web` を指定できます。

`service start` はプロセスを起動する前に Issue Tracker と orchestrator の未適用マイグレーションを確認し、必要な場合は `tq migrate` を案内します。既定ではポート 37651、37652、37653 を使います。いずれかが使用中の場合は loopback ポートを提案し、`-y` がない限り確認を求めます。`service stop` は Web、orchestrator、Issue Tracker の順に停止します。`service status` はサービスの状態、PID、ポート、稼働時間を表示し、JSON 出力にも対応します。

`logs` は `$TQ_HOME/system/log/` 以下のファイルを読み、`-n` と `-f` を使えますが JSON 出力には対応しません。`web` はサービス状態から URL を開き、Web UI が起動していない場合は失敗します。

`tq tui [--orchestrator-url URL]` はターミナルを必要とし、テキスト出力だけに対応します。課題、コメント、Artifact、実行状態を読み取りますが、データを変更するリクエストは送信しません。`--orchestrator-url` を指定すると、`state.json` による orchestrator の検出を上書きできます。

`tq config` は、バージョン、ビルドプロファイル、`TQ_HOME` の上書き、解決済みホームディレクトリ、設定ファイルのパス、解決済みの値を表示します。YAML の生データは表示しません。スクリプトではグローバルオプションの `--output json` を使用します。

`tq update` は現在のバージョンと更新先のバージョンを表示し、ローカルサービスの停止と再起動が入ることを確認してから、既定では最新の正式リリースをインストールします。その後、新しくインストールされた `tq version` を確認し、マイグレーションを実行してサービスを起動します。`-y` は確認プロンプトを省略します。`--tag` は特定のリリースまたはプレリリースのタグをインストールします。

`dev` など空でないビルドプロファイルを持つバイナリでは、汎用のリリース成果物がそのプロファイルを保持しないため、`tq update` を使用できません。

具体的な手順とサービス停止時の注意事項は、[Tasq を更新する](pathname:///guides/update-tasq)を参照してください。

## API 直接実行コマンド

型付きコマンドがない Issue Tracker 操作には `tq api` を使用します。

```sh
tq api GET /api/v1/issues --query states=ready
tq api POST /api/v1/issues --header 'X-Request-ID: local-123' --data @request.json
```

```text
tq api <method> <path> [--query key=value] [--header 'Name: value'] [--data value|@file|-]
```

パスには許可リストに含まれる、エンコードされていない絶対 `/api/v1/...` パスを指定します。完全な URL、フラグメント、ドットセグメント、空セグメント、末尾のスラッシュは拒否されます。`--query` と `--header` は繰り返し指定できます。`--data` にはリテラル値、`@file`、または標準入力を表す `-` を指定でき、`POST`、`PUT`、`PATCH` でだけ使用できます。

書き込みや削除の前に確認プロンプトは表示されず、リダイレクトにも追従しません。タイムアウトは 10 秒です。レスポンスのバイト列をそのまま出力するため、グローバルな `--output` は変換に使われません。終了コードは HTTP `2xx` で `0`、HTTP または通信エラーで `1`、使い方、入力、許可リストの検証エラーで `2` です。
