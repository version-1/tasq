# Tasq 運用

このドキュメントでは、ローカル開発環境、検証コマンド、未決定の設計事項を扱います。所有境界とコンポーネントの責務は [architecture.ja.md](architecture.ja.md) を参照してください。ユーザー向け API サーフェスは [api.ja.md](api.ja.md) を参照してください。

## Development Environment

Docker Compose は、ローカル開発環境を長時間起動する `dev` container と standalone OpenAPI UI container に集約します。`dev` container 内では issue-tracker が container port `8080`、orchestrator が container port `8081`、Go Web サーバーが container port `3000` で待ち受けます。

個人マシンでホストのみの運用を行う場合、`tq service start` は issue-tracker、orchestrator、web をバックグラウンドプロセスとして起動します。固定ローカルポート `37651`、`37652`、`37653` を優先し、いずれかが使用中なら全サービスに OS 選択の loopback ポートを 1 つずつ提案して、対話確認（または `-y`）後に起動します。確認後に提案ポートを再確認し、取得されていた場合は別の組を選ばず失敗します。選択したアドレスは検出用状態として `$TQ_HOME/system/state.json` に書き込み、ログを `$TQ_HOME/system/log/` 配下へ追記します。

データベースが新規の場合、または未適用のマイグレーションがある場合は、サービス起動前に `tq migrate` を実行します。`tq service start` はサービスプロセスを起動する前に issue-tracker と orchestrator のデータベースを確認し、未適用のマイグレーションがあれば `tq migrate` の実行を案内して終了します。サービス側もスキーマ変更を自動適用せず、同じ案内を出して fail fast します。

state に記録されたローカル orchestrator だけを管理するには、`tq orchestrator start`、`tq orchestrator stop`、`tq orchestrator status` を使います。起動時には state に記録された Issue Tracker が起動中である必要があり、既定 orchestrator port を使います。Issue Tracker や Web service の起動、停止、再設定は行いません。停止は冪等で、`tq service stop` と同じ graceful shutdown 経路を使います。

推奨コマンド:

- `make run-issue-tracker`
- `make run-orchestrator`
- `make dev-up`
- `make run-web`
- `make run-tui`
- `make dev-ports`
- `make dev-codex-login`

CLI コマンド:

- `make build-tq`
- `make run-migrate`
- `TQ_HOME=./.tasq go run ./cmd/tq migrate`
- `make run-tq ARGS="issue list"`
- `make run-tq ARGS="issue get 1"`
- `TQ_HOME=./.tasq go run ./cmd/tq service status`
- `tq tui`（別名: `tq console`、`tq c`）

`tq tui` には対話型端末と text 出力が必要です。Issue Tracker の URL は通常の CLI 解決順に従います。`--orchestrator-url` はサービス状態に保存されたオーケストレーターのアドレスより優先されます。オーケストレーターが未設定または利用できない場合は Run タブだけを縮退表示し、Issue Tracker の障害時は再試行画面を表示します。

`make build-tq` はホスト用の `tq` binary を `./bin/tq` に build し、release build と同じ共有 build-info ldflags variable を通じて現在の short commit hash を `tq version` に注入します。

`make dev-up` は OpenAPI UI を起動し、`dev` container 内で issue-tracker、orchestrator、web-ui を起動します。実行時状態は `$TQ_HOME` 配下に保存され、container 内の既定値は `/workspace/.tasq` です。`run-all` step はサービス起動前に migration を明示的に適用します。`make dev-codex-login` は device auth を使い、Codex の認証情報を `codex-home` Docker volume に永続化します。

## Verification

現在の検証コマンド:

```sh
go test ./...
```

```sh
cd cmd/web/frontend
npm run typecheck
npm run build
```

手動検証:

1. `make dev-up` で開発環境を起動する。
2. UI または `tq` で課題を作成・更新する。
3. issue-tracker のサマリーに課題の状態変更が反映されることを確認する。
4. 表示された orchestrator URL で実行時調査機能を確認する。

Web サーバーは `/tracker/*` を issue-tracker に、`/orchestrator/*` を orchestrator にプロキシします。Compose では、`make run-web` が dev container 内の `127.0.0.1:8080` と `127.0.0.1:8081` を backend URL として Web サーバーを起動します。

## Open Decisions

- 外部 tracker 同期を issue-tracker 内に置くか、provider interface の背後に置くか。
- 本番向けの認証、認可、ネットワーク公開。
- 完全な Codex transcript を SQLite に残すか、filesystem artifact に移すか。
