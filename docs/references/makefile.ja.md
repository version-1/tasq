# Makefile リファレンス

リポジトリの Makefile は、ローカル開発の主な入口です。Docker Compose をラップし、dev container を起動し、その container 内でサービスプロセスを起動して、割り当てられたローカル URL を表示します。

接頭辞ガイドとセクション分けされたターゲット一覧は `make help` で確認できます。ターゲット一覧は Makefile コメントから生成されます。

## 設定変数

| Variable | Default | Description |
|---|---|---|
| `COMPOSE` | `docker compose` | Docker Compose コマンド。ラッパーや別の Compose バイナリを使う場合に上書きします。 |
| `BROWSER_OPEN` | `open` | `dev-open` が使うブラウザー起動コマンド。ヘッドレス環境では別のコマンドや no-op コマンドに変更できます。 |
| `TQ_HOME` | `./.tasq` | ホスト上のリポジトリローカルな Tasq 実行時状態。dev container では `/workspace/.tasq` を使います。 |
| `ISSUE_TRACKER_PORT` | empty | issue-tracker のホストポート。empty の場合は Docker Compose が空きポートを割り当てます。 |
| `ORCHESTRATOR_PORT` | empty | orchestrator のホストポート。empty の場合は Docker Compose が空きポートを割り当てます。 |
| `OPENAPI_PORT` | empty | OpenAPI UI のホストポート。empty の場合は Docker Compose が空きポートを割り当てます。 |
| `WEB_PORT` | empty | Web UI のホストポート。empty の場合は Docker Compose が空きポートを割り当てます。 |
| `WEB_ISSUE_TRACKER_URL` | empty | 互換性のために予約されています。Go Web server は、ブラウザー向けのビルド時 API 接続先ではなくプロキシ設定を使います。 |
| `RELEASE_BRANCH` | `main` | 正式リリースターゲットが要求するブランチ。 |
| `RELEASE_REMOTE` | `origin` | リリースタグの push 先 remote。 |
| `RELEASE_REPO` | `version-1/tasq` | リリース資産から `tq` をインストールするときに使う GitHub リポジトリ。 |
| `TQ_INSTALL_DIR` | `$HOME/.local/bin` | リリース用インストールターゲットが `tq` バイナリを配置するディレクトリ。 |
| `TQ_INSTALL_NAME` | `tq` | リリース版 `tq` バイナリとしてインストールするコマンド名。管理対象サービスの実行ファイルは、固定名で同じ場所にインストールされます。 |
| `TQ_BUILD_COMMIT` | current short Git commit | ホスト用 `tq` ビルドに ldflags 経由で注入するコミット値。 |
| `TQ_BUILD_LDFLAGS` | `buildCommit` ldflags assignment | `make build-tq` が使う Go linker flags。 |
| `AIR_VERSION` | `v1.52.3` | Go サービスを watch mode で動かすために使う Air のバージョン。 |

固定ポートを使う例:

```sh
ISSUE_TRACKER_PORT=8080 ORCHESTRATOR_PORT=8081 OPENAPI_PORT=8082 WEB_PORT=3000 make dev-up
```

## 主な開発ターゲット

| Target | Purpose |
|---|---|
| `make build-tq` | ホスト用の `tq` バイナリを `./bin/tq` にビルドし、現在の短いコミットハッシュを `tq version` に注入します。 |
| `make dev-up` | `dev` container と OpenAPI UI を起動し、issue-tracker、orchestrator、Web をバックグラウンドで起動して URL を表示します。 |
| `make dev-restart` | Compose サービスを停止し、再度 `dev-up` を実行します。 |
| `make dev-down` | Compose サービスを停止します。 |
| `make dev-reset-db CONFIRM=1` | Compose を停止し、`.tasq/system/data/` 配下のローカル SQLite ファイルを削除して開発環境を再起動します。 |
| `make dev-openapi` | OpenAPI UI の Compose サービスだけを起動し、ポートを表示します。 |
| `make dev-open` | Web UI と OpenAPI UI をブラウザーで開きます。 |
| `make dev-ports` | 現在割り当てられている issue-tracker、orchestrator、OpenAPI、Web の URL を表示します。 |

`dev-up` は既定でホストポートを自動割り当てします。割り当てられた URL は表示しますが、ブラウザーは開きません。ブラウザーを開く場合は `make dev-open` を明示的に実行してください。

## Container ターゲット

`dc-*` ターゲットは、Docker Compose サービスの状態確認と dev container 自体に対する操作に使います。

| Target | Purpose |
|---|---|
| `make dc-ready` | 必要なツールと volume が `codex` user から書き込み可能になるまで待機します。 |
| `make dc-ps` | Docker Compose サービスの状態を表示します。 |
| `make dc-shell` | 実行中の dev container に `codex` user で shell を開きます。 |
| `make dc-exec CMD="..."` | dev container 内で任意のコマンドを `codex` user として実行します。 |

よく使う例:

```sh
make dc-ps
make dc-shell
make dc-exec CMD="go test ./internal/config"
```

## Runtime ターゲット

`run-*` ターゲットは、既に起動している dev container 内で動くプロセスやコマンドに使います。

| Target | Purpose |
|---|---|
| `make run-all` | マイグレーションを適用してから、起動済み dev container 内で issue-tracker、orchestrator、Web を起動します。 |
| `make run-migrate` | 起動済み dev container 内でローカル SQLite マイグレーションを適用します。 |
| `make run-stop` | container を停止せず、dev container 内の Air 管理下のサービスプロセスだけを停止します。 |
| `make run-issue-tracker` | 起動済み dev container 内で issue-tracker プロセスだけを起動します。 |
| `make run-is` | `run-issue-tracker` の alias です。 |
| `make run-orchestrator` | issue-tracker を起動してから orchestrator プロセスを起動します。 |
| `make run-or` | `run-orchestrator` の alias です。 |
| `make run-web` | issue-tracker を起動してから Web プロセスを起動します。 |
| `make run-w` | `run-web` の alias です。 |
| `make run-tui` | dev container 内で TUI を対話的に実行します。 |
| `make run-tq ARGS="..."` | サービスプロセスを変更せず、起動済み dev container 内でインストール済みの `tq $(ARGS)` バイナリを実行します。 |
| `make run-ps` | dev container 内で動いている開発用プロセスを表示します。 |
| `make run-logs` | `$TQ_HOME/system/log/*.log` を追跡表示します。 |

よく使う例:

```sh
make run-is
make run-migrate
make run-or
make run-w
make run-tq ARGS="issue list"
make run-logs
```

## 検証ターゲット

| Target | Purpose |
|---|---|
| `make dev-test` | dev container 内で `go test ./...`、Web 依存関係のインストール、Web typecheck を実行します。 |
| `make dev-build` | dev container 内で `go test ./...`、Web 依存関係のインストール、Web 本番ビルドを実行します。 |

## リリースターゲット

| Target | Purpose |
|---|---|
| `make prerelease` | `scripts/release.sh` 経由で prerelease タグを作成して push します。 |
| `make prerelease version=v0.3.0` | 特定の正式バージョンを基にした prerelease タグを作成して push します。 |
| `make release version=v0.1.1` | `scripts/release.sh` 経由で正式リリースタグを作成して push します。 |
| `make install-tq` | 最新の正式リリースから `tq` と管理対象サービスの実行ファイルを `$HOME/.local/bin` にインストールします。 |
| `make install-tq version=v0.1.0` | 特定のリリースタグから `tq` と管理対象サービスの実行ファイルをインストールします。 |
| `make install-tq-prerelease` | 最新の prerelease から `tq` と管理対象サービスの実行ファイルをインストールします。バージョンを指定しない場合は `gh` が必要です。 |
| `make install-tq-prerelease version=v0.1.0-pre.1` | 特定の prerelease タグから `tq` と管理対象サービスの実行ファイルをインストールします。 |

タグ、GitHub Actions、GoReleaser を含む全体の流れは [Deployment Flow](../design/deployment.ja.md) を参照してください。

## 認証ターゲット

| Target | Purpose |
|---|---|
| `make dev-codex-login` | dev container 内で `codex login --device-auth` を実行し、認証情報を `codex-home` Docker volume に永続化します。 |
| `make dev-codex-status` | dev container 内で Codex authentication status を表示します。 |
| `make dev-gh-login` | dev container 内で `gh auth login` と `gh auth setup-git` を実行し、認証情報を `gh-config` Docker volume に永続化します。 |
| `make dev-gh-status` | dev container 内で GitHub CLI authentication status を表示します。 |

container login では、ブラウザーのリダイレクト先が container 内の localhost callback になり、ホスト側ブラウザーから到達できない場合に device auth を使います。

認証ターゲットは、既存の `dev` container 内でコマンドを実行するだけです。container の build や起動は行わないため、dev container がない場合は先に `make dev-up` を実行します。

例:

```sh
make dev-codex-login
make dev-codex-status
make dev-gh-login
make dev-gh-status
```

## 運用メモ

`dev` container は長時間使う前提です。サービスプロセスは独立した Compose サービスではなく、`docker compose exec` で起動される通常の子プロセスです。プロセスだけを止める場合は `make run-stop`、Compose サービスも止める場合は `make dev-down` を使います。

Makefile は container 内のコマンドを `codex` user として実行します。container の起動時には、Go module cache、Go build cache、`cmd/web/frontend` 配下の Web `node_modules`、Codex 認証情報、GitHub CLI 認証情報用の named volume が書き込み可能になるよう準備します。

Go Web server は、ブラウザーからの API 呼び出しを container-local URL 経由で issue-tracker と orchestrator にプロキシします。フロントエンドコード、プロキシ設定、バックエンドポートを変更した場合は、`make run-web` で Web プロセスを再起動してください。
