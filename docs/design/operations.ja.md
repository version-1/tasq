# Tasq 運用

このドキュメントでは local development environment、verification command、open design decision を扱います。所有境界と component responsibility は [architecture.ja.md](architecture.ja.md) を参照してください。user-facing API surface は [api.ja.md](api.ja.md) を参照してください。

## Development Environment

Docker Compose は local development を長時間起動する `dev` container と standalone OpenAPI UI container に集約します。`dev` container 内では issue-tracker が container port `8080`、orchestrator が container port `8081`、Go Web server が container port `3000` で待ち受けます。

Personal machine 上の host-only operation では、`tq service start` が issue-tracker と orchestrator を background process として起動します。固定 local port `37651` と `37652` を使い、discovery state を `$TQ_HOME/system/state.json` に書き込み、log を `$TQ_HOME/system/log/` 配下へ追記します。

Recommended commands:

- `make run-issue-tracker`
- `make run-orchestrator`
- `make dev-up`
- `make run-web`
- `make run-tui`
- `make dev-ports`
- `make dev-codex-login`

CLI commands:

- `make run-tq ARGS="issue list"`
- `make run-tq ARGS="issue get 1"`
- `TQ_HOME=./.tasq go run ./cmd/tq service status`

`make dev-up` は OpenAPI UI を起動し、`dev` container 内で issue-tracker、orchestrator、web-ui を起動します。Runtime state は `$TQ_HOME` 配下に保存され、container 内の default は `/workspace/.tasq` です。`make dev-codex-login` は device auth を使い、Codex authentication を `codex-home` Docker volume に永続化します。

## Verification

現在の verification command:

```sh
go test ./...
```

```sh
cd cmd/web/frontend
npm run typecheck
npm run build
```

Manual verification:

1. `make dev-up` で dev environment を起動する。
2. UI または `tq` で issue を作成・更新する。
3. issue-tracker summary が issue status change を反映することを確認する。
4. 表示された orchestrator URL で runtime inspection を確認する。

Web server は `/api/tracker/*` を issue-tracker に、`/api/orchestrator/*` を orchestrator に proxy します。Compose では `make run-web` が dev container 内の `127.0.0.1:8080` と `127.0.0.1:8081` を backend URL として Web server を起動します。

## Open Decisions

- external tracker sync を issue-tracker 内に置くか provider interface の behind に置くか。
- Production authentication、authorization、network exposure。
- large full-fidelity Codex transcript を SQLite に残すか filesystem artifact に移すか。
