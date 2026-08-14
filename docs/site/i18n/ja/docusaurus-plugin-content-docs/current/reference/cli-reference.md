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
| `tq issue watch [--interval <duration>] [--seen-ttl <duration>] [--verbose]` | 実行可能な課題を定期取得し、JSON イベントを出力します。 |
| `tq issue close <id>` | 課題を `done` に移動します。 |
| `tq issue cancel <id>` | 課題を `cancelled` に移動します。 |
| `tq issue ready <id>` | 課題を `ready` に移動します。 |
| `tq issue draft <id>` | 課題を `backlog` に移動します。 |
| `tq issue rename <id> <title>` | タイトルを更新します。 |
| `tq issue edit <id> <description>` | 説明を更新します。 |

create と update は、該当する場合に `--title`、`--description`、`--status`、
`--priority`、`--assignee`、`--attach` を受け付けます。update では、依存関係を
置き換える `--dependency <comma-separated-ids>` と、依存関係を削除する
`--clear-dependencies` も指定できます。

`tq issue watch` はエージェントのループ処理向けです。実行待ちキューを読み、設定された TTL の間は同じ課題を重複して出力せず、`issue-ready` イベントを出力します。一時的な API エラーが起きても取得を継続します。

## Artifact コマンド

| コマンド | 用途 |
| --- | --- |
| `tq artifact set <issue-id> --type pull_request <url>` | 課題のプルリクエスト URL を作成または置き換えます。 |
| `tq artifact delete <issue-id> --type pull_request` | 課題のプルリクエスト URL を削除します。 |

どちらのコマンドも正の課題 ID と `--type` を必要とし、グローバルな text / JSON 出力モードに対応します。

## コメントコマンド

| コマンド | 用途 |
| --- | --- |
| `tq comment add <issue-id> --body <body>` | コメントを追加します。 |
| `tq comment list <issue-id>` | 課題のコメントを一覧表示します。 |

コメント種別には `progress`、`blocker`、`handoff`、`general` を指定できます。

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

## 実行環境のコマンド

| コマンド | 用途 |
| --- | --- |
| `tq service start` | issue-tracker、orchestrator、Web UI を起動します。 |
| `tq service stop` | local services を停止します。 |
| `tq service status` | service status を表示します。 |
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
