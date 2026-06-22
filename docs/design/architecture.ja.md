# Tasq アーキテクチャ

Tasq は、課題管理と orchestrator の実行状態の観測をローカル優先で行うタスクシステムです。

現在のアーキテクチャでは、課題管理とオーケストレーションを分離します。issue-tracker は課題の状態とユーザー向け API を所有します。orchestrator は過去の実行状態と任意の実行時調査機能を所有します。UI クライアントは主に issue-tracker にアクセスします。Web UI サーバーは、将来の実行状態ビューに備えて orchestrator へのプロキシパスも公開します。

## Goals

- 課題の状態と実行状態を別々の概念として扱い、それぞれの所有者を分ける。
- web-ui、tui、エージェント向け CLI ツールが同じユーザー向け API サーフェスを使えるようにする。
- tracker API の対象を課題、プロジェクト、ワークスペース、サマリーデータに集中させる。
- オーケストレーションの実行時状態を orchestrator のローカルに閉じ込める。

## Non-goals

- ホスト型のマルチテナント運用。
- 本番向けの認証と認可。
- 最初の実装範囲における完全な Codex app-server runner。
- 最初の実装範囲における Linear などの外部 tracker 連携。
- 割り当てや worker スケジューリングの意味論。

## Components

### web-ui

web-ui は、課題操作のために Go から配信される Vite + React の single-page app です。

Responsibilities:

- issue-tracker から課題サマリーを取得する。
- 課題の状態、優先度、担当者を表示する。
- issue-tracker を呼び出して課題を状態間で移動する。
- SPA fallback でブラウザルートを配信する。
- `/tracker/*` を issue-tracker に、`/orchestrator/*` を orchestrator にプロキシする。

Web UI の構造とスタイル規約は [web.ja.md](web.ja.md) を参照してください。

### tui

TUI は、同じ issue-tracker API を使う Go 製のターミナルクライアントです。

Responsibilities:

- issue-tracker から課題サマリーを取得する。
- 課題カラムを描画する。
- 1 回限りの描画と watch-mode rendering をサポートする。
- orchestrator を直接呼び出さない。

### tq

`tq` は、HTTP の詳細を埋め込まずに課題の状態を変更する必要があるエージェントやワークフローツール向けの単体 Go CLI です。

Responsibilities:

- issue-tracker API 経由で課題を作成、取得、一覧表示、更新する。
- 課題の説明とコメント用の画像添付ファイルをアップロードする。
- 既定では人が読みやすい出力を使い、ツール利用向けに JSON 出力をサポートする。
- issue-tracker API URL を `--api-url`、`TQ_API_URL`、`$TQ_HOME/system/state.json`、または `http://localhost:37651` から解決する。
- `tq service` でホストローカルの issue-tracker と orchestrator のプロセスを管理する。
- コマンドが失敗した場合は、stderr に機械判読可能な JSON エラーを出力し、ゼロ以外の終了コードを返す。
- orchestrator を直接呼び出さない。

### issue-tracker

issue-tracker は課題管理と表示用集約を所有します。

Responsibilities:

- 課題を SQLite に保存する。
- 各課題が必ず 1 つのプロジェクトに属するようにする。
- 課題を作成、編集、一覧表示する。
- 添付ファイルのメタデータを SQLite に保存し、添付ファイルのバイト列を `$TQ_HOME` 配下に保存する。
- orchestrator やツールの突き合わせに使う課題の状態を返す。
- UI/TUI 向けのサマリー API を提供する。

issue-tracker は課題の状態、優先度、タイトル、説明、担当者、コメント、添付ファイル、プロジェクトの source of truth です。
関連する課題が存在するプロジェクトは削除できません。

### orchestrator

orchestrator は実行状態と実行時調査機能を所有します。

Responsibilities:

- 自身の SQLite データベースに実行レコードを作成する。
- オーケストレーションの設定に使うリポジトリワークフロー契約を読み込む。
- 設定済みのワークスペースルート配下に、課題ごとにサニタイズ済みのワークスペースを作成する。
- runner event とワークスペースメタデータを記録する。
- 実行時状態と課題ごとの実行詳細を扱う任意の loopback HTTP API を公開する。

orchestrator は run record、run attempt、runner event、workspace metadata の source of truth です。

### agent

将来のエージェントは、orchestrator に制御される Codex app-server プロセスです。

Responsibilities:

- orchestrator からタスクを受け取る。
- ワークスペース内でタスクを実行する。
- JSON-RPC 経由で実行の進捗を orchestrator に報告する。

orchestrator は runner boundary を通じて Codex app-server を起動し、ローカルの runstore に実行の進捗を記録します。

### workspace

workspace manager は、エージェント向けに隔離された実行環境を提供します。

Responsibilities:

- git ワークスペースを作成、管理する。
- 隔離されたワークスペースで並列実行と検証をサポートする。
- デバッグと復旧に必要なメタデータを保持する。

現在の workspace manager は、課題ごとにサニタイズ済みのワークスペースディレクトリを作成します。新しく作成したワークスペースは、設定済みのリポジトリソースから構築し、復旧とデバッグに使うクリーンアップおよび構築時のメタデータを記録します。

## Dependency Direction

ユーザー向けクライアントとエージェント向けワークフローツールは、issue-tracker API のみに依存します。

orchestrator は issue-tracker の作業キューやイベント受信エンドポイントを使いません。過去の実行データと runner-event データは orchestrator の SQLite store に残り、任意の orchestrator HTTP API から参照できます。

```text
web-ui ─┐
tui ────┼─ issue-tracker ── SQLite: issues, comments, attachments, projects
tq ─────┘
                 │
                 └─ $TQ_HOME/system/data/attachments

        orchestrator ───── SQLite: runs, runner_events, workspace metadata
                │
                ├─ future: agent-runner ── Codex app-server over JSON-RPC
                └─ workspace manager ── git workspace / isolated runtime
```

## State Ownership

課題の状態と実行状態は別々のものです。

課題の状態は issue-tracker が所有します。

- `backlog`
- `ready`
- `in_progress`
- `review`
- `done`
- `blocked`
- `failed`

実行状態は orchestrator が所有します。

- `queued`
- `starting`
- `running`
- `waiting_for_input`
- `succeeded`
- `failed`
- `cancelled`

orchestrator は課題の状態を直接変更しません。課題の状態変更は issue-tracker の課題 API 経由で行います。

## Current MVP Behavior

現在の実装範囲には次が含まれます。

- `cmd/issue-tracker`
- `cmd/tq`
- `cmd/orchestrator`
- 課題とプロジェクト用の issue-tracker SQLite テーブル。
- issue-tracker の添付ファイルメタデータを SQLite に保存し、画像バイト列を `$TQ_HOME` 配下に保存する仕組み。
- 実行、runner event、ワークスペースメタデータ、ワークスペースセットアップ失敗用の orchestrator SQLite テーブル。
- web-ui と tui が利用する issue-tracker サマリー API。
- `tq` が利用する課題 CRUD API。
- `attachment://<id>` 画像参照を含む Markdown 形式の課題説明とコメント本文。
- Codex runner lifecycle: app-server startup、live-thread turn、有効化時の continuation turn、終端の run status reporting。

simulated runner は限定的なテスト用に残しますが、本番配線では Codex app-server runner を使います。
