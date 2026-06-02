# Makefile リファレンス

Repository の Makefile は local development の主な入口です。Docker Compose を wrap し、dev container を起動し、その container 内で service process を起動し、割り当てられた local URL を表示します。

Prefix guide と section 分けされた target 一覧は `make help` で確認できます。Target 一覧は Makefile comment から生成されます。

## Configuration Variables

| Variable | Default | Description |
|---|---|---|
| `COMPOSE` | `docker compose` | Docker Compose command。wrapper や別の Compose binary を使う場合に override します。 |
| `BROWSER_OPEN` | `open` | `dev-open` が使う browser opener。headless environment では別 command や no-op command にできます。 |
| `TQ_HOME` | `./.tasq` | Host 上の repository-local Tasq runtime state。dev container では `/workspace/.tasq` を使います。 |
| `ISSUE_TRACKER_PORT` | empty | issue-tracker の host port。empty の場合は Docker Compose が free port を割り当てます。 |
| `ORCHESTRATOR_PORT` | empty | orchestrator の host port。empty の場合は Docker Compose が free port を割り当てます。 |
| `OPENAPI_PORT` | empty | OpenAPI UI の host port。empty の場合は Docker Compose が free port を割り当てます。 |
| `WEB_PORT` | empty | Web UI の host port。empty の場合は Docker Compose が free port を割り当てます。 |
| `WEB_ISSUE_TRACKER_URL` | empty | Web UI に渡す issue-tracker URL。empty の場合は Makefile が割り当て済み issue-tracker port を解決します。 |
| `AIR_VERSION` | `v1.52.3` | Go service を watch mode で動かす Air version。 |

固定 port を使う例:

```sh
ISSUE_TRACKER_PORT=8080 ORCHESTRATOR_PORT=8081 OPENAPI_PORT=8082 WEB_PORT=3000 make dev-up
```

## Main Development Targets

| Target | Purpose |
|---|---|
| `make dev-up` | `dev` container と OpenAPI UI を起動し、issue-tracker、orchestrator、Web を background で起動して URL を表示します。 |
| `make dev-restart` | Compose services を停止し、再度 `dev-up` を実行します。 |
| `make dev-down` | Compose services を停止します。 |
| `make dev-reset-db CONFIRM=1` | Compose を停止し、`.tasq/system/data/` 配下の local SQLite files を削除して dev environment を再起動します。 |
| `make dev-openapi` | OpenAPI UI Compose service だけを起動し、ports を表示します。 |
| `make dev-open` | Web UI と OpenAPI UI を browser で開きます。 |
| `make dev-ports` | 現在割り当てられている issue-tracker、orchestrator、OpenAPI、Web の URL を表示します。 |

`dev-up` は default で host port を自動割り当てします。割り当てられた URL は表示しますが、browser は開きません。Browser を開く場合は `make dev-open` を明示的に実行します。

## Container Targets

`dc-*` targets は Docker Compose service status と dev container 自体に対する操作に使います。

| Target | Purpose |
|---|---|
| `make dc-ready` | 必要な tools と volumes が `codex` user から writable になるまで待機します。 |
| `make dc-ps` | Docker Compose service status を表示します。 |
| `make dc-shell` | 実行中の dev container に `codex` user で shell を開きます。 |
| `make dc-exec CMD="..."` | dev container 内で任意 command を `codex` user として実行します。 |

よく使う例:

```sh
make dc-ps
make dc-shell
make dc-exec CMD="go test ./internal/config"
```

## Runtime Targets

`run-*` targets は既に起動している dev container 内で動く process や command に使います。

| Target | Purpose |
|---|---|
| `make run-all` | 起動済み dev container 内で issue-tracker、orchestrator、Web を起動します。 |
| `make run-stop` | container を停止せず、dev container 内の Air と Next.js process だけを停止します。 |
| `make run-issue-tracker` | 起動済み dev container 内で issue-tracker process だけを起動します。 |
| `make run-is` | `run-issue-tracker` の alias です。 |
| `make run-orchestrator` | issue-tracker を起動してから orchestrator process を起動します。 |
| `make run-or` | `run-orchestrator` の alias です。 |
| `make run-web` | issue-tracker を起動してから Web process を起動します。 |
| `make run-w` | `run-web` の alias です。 |
| `make run-tui` | dev container 内で TUI を interactive に実行します。 |
| `make run-tq ARGS="..."` | service process を変更せず、起動済み dev container 内で installed `tq $(ARGS)` binary を実行します。 |
| `make run-ps` | dev container 内で動いている dev process を表示します。 |
| `make run-logs` | `.tmp/dev-logs/*.log` を follow します。 |

よく使う例:

```sh
make run-is
make run-or
make run-w
make run-tq ARGS="issue list"
make run-logs
```

## Verification Targets

| Target | Purpose |
|---|---|
| `make dev-test` | dev container 内で `go test ./...`、Web dependency install、Web typecheck を実行します。 |
| `make dev-build` | dev container 内で `go test ./...`、Web dependency install、Web production build を実行します。 |

## Authentication Targets

| Target | Purpose |
|---|---|
| `make dev-codex-login` | dev container 内で `codex login --device-auth` を実行し、credential を `codex-home` Docker volume に永続化します。 |
| `make dev-codex-check` | dev container 内で Codex CLI と `codex app-server` が利用できることを確認します。 |
| `make dev-gh-login` | dev container 内で `gh auth login` を実行し、credential を `gh-config` Docker volume に永続化します。 |
| `make dev-gh-status` | dev container 内で GitHub CLI authentication status を表示します。 |
| `make dev-tool-auth-check` | dev container 内で Codex login status と GitHub CLI authentication status をまとめて確認します。 |

Container login では、browser redirect が container 内の localhost callback に戻り host browser から
到達できない場合に device auth を使います。

例:

```sh
make dev-codex-login
make dev-codex-check
make dev-gh-login
make dev-gh-status
make dev-tool-auth-check
```

## Operational Notes

`dev` container は long-lived です。Service process は separate Compose service ではなく、`docker compose exec` で起動される通常の child process です。Process だけ止める場合は `make run-stop`、Compose services も止める場合は `make dev-down` を使います。

Makefile は container 内 command を `codex` user として実行します。Container startup 時に、Go module
cache、Go build cache、Web `node_modules`、Codex credential、GitHub CLI credential 用の named
volume が writable になるように準備します。

`NEXT_PUBLIC_ISSUE_TRACKER_URL` は Web process 起動時に解決されます。Web 起動後に issue-tracker の host port が変わった場合は、`make run-web` で Web process を再起動するか、`make dev-up` で全体を再起動してください。
