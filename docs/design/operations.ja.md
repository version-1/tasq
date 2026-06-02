# Tasq 運用

このドキュメントでは local development environment、verification command、open design decision を扱います。所有境界と component responsibility は [architecture.ja.md](architecture.ja.md) を参照してください。user-facing API surface は [api.ja.md](api.ja.md) を参照してください。

## Development Environment

Docker Compose は local development を長時間起動する `dev` container と standalone OpenAPI UI container に集約します。`dev` container 内では issue-tracker が container port `8080`、orchestrator が container port `8081`、web-ui が container port `3000` で待ち受けます。

Recommended commands:

- `make issue-tracker-up`
- `make orchestrator-up`
- `make dev-up`
- `make dev-up-forward`
- `make web-up`
- `make tui-up`
- `make dev-ports`
- `make codex-login`

CLI commands:

- `make tq ARGS="issue list"`
- `make tq ARGS="issue get 1"`

`make dev-up` は OpenAPI UI を起動し、`dev` container 内で issue-tracker、orchestrator、web-ui を起動します。Runtime state は `$TQ_HOME` 配下に保存され、container 内の default は `/workspace/.tasq` です。`make codex-login` は device auth を使い、Codex authentication を `codex-home` Docker volume に永続化します。

## Verification

現在の verification command:

```sh
go test ./...
```

```sh
cd web
npm run typecheck
npm run build
```

Manual verification:

1. `make dev-up` で dev environment を起動する。
2. UI または `tq` で issue を作成・更新する。
3. issue-tracker summary が issue status change を反映することを確認する。
4. 表示された orchestrator URL で runtime inspection を確認する。

## Open Decisions

- external tracker sync を issue-tracker 内に置くか provider interface の behind に置くか。
- Production authentication、authorization、network exposure。
- large full-fidelity Codex transcript を SQLite に残すか filesystem artifact に移すか。
