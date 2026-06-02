# Tasq 運用

このドキュメントでは local development environment、verification command、open design decision を扱います。所有境界と component responsibility は [architecture.ja.md](architecture.ja.md) を参照してください。user-facing API surface は [api.ja.md](api.ja.md) を参照してください。

## Development Environment

Docker Compose は issue-tracker を container port `8080`、web-ui を container port `3000`、orchestrator service を optional runtime inspection 用に実行します。

Recommended commands:

- `make issue-tracker-up`
- `make orchestrator-up`
- `make dev-up`
- `make dev-up-forward`
- `make web-up`
- `make tui-up`
- `make dev-status`

Host commands:

- `go run ./cmd/tq --api-url http://localhost:8080 issue list`
- `TQ_API_URL=http://localhost:8080 go run ./cmd/tq issue get 1`

`make web-up` は issue-tracker、orchestrator、web-ui を起動します。Web UI は `/api/v1/...` を Compose network 内の issue-tracker に proxy します。

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

1. issue-tracker と Web UI を起動する。
2. UI または `tq` で issue を作成・更新する。
3. issue-tracker summary が issue status change を反映することを確認する。
4. runtime inspection が必要な場合は `--port` 付きで orchestrator を起動する。

## Open Decisions

- external tracker sync を issue-tracker 内に置くか provider interface の behind に置くか。
- Production authentication、authorization、network exposure。
- large full-fidelity Codex transcript を SQLite に残すか filesystem artifact に移すか。
