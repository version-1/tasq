# tasq

実行可能な作業を管理し、coding agents に割り当て、Web UI または TUI から進捗を確認するための local-first issue tracker and task orchestrator です。

## Components

- Issue Tracker: SQLite backed の Go REST API です。issues、work items、UI summaries を所有します。
- Orchestrator: SQLite backed の Go worker です。実行可能な work を claim し、run state を記録します。
- Web UI: issue-tracker API 用の Next.js client です。
- TUI: 同じ issue-tracker API を使う Go terminal client です。

Architecture 全体は [docs/design.md](docs/design.md) を参照してください。

## Quick Start

Local development では `make` 経由で Docker Compose を使います。

```sh
make web-up
```

このコマンドは issue-tracker、orchestrator、OpenAPI UI、Web UI を起動します。Docker Compose が host ports を自動割り当てし、割り当てられた URLs を表示します。

URLs を再表示します。

```sh
make dev-ports
```

環境を停止します。

```sh
make dev-down
```

利用可能な development commands を一覧します。

```sh
make help
```

## Verification

標準の Compose-backed checks を実行します。

```sh
make dev-test
```

Go services と Web UI の両方に影響する変更を handoff する前は、broader build check を実行します。

```sh
make dev-build-app
```

## Documentation

- [WORKFLOW.md](WORKFLOW.md): repository workflow、task flow、documentation update rules、component workflow links。
- [docs/design.md](docs/design.md): system architecture と service boundaries。
- [cmd/issue-tracker/WORKFLOW.md](cmd/issue-tracker/WORKFLOW.md): issue-tracker development workflow。
- [cmd/orchestrator/WORKFLOW.md](cmd/orchestrator/WORKFLOW.md): orchestrator development workflow。
- [web/WORKFLOW.md](web/WORKFLOW.md): Web UI development workflow。
- [web/docs/design.md](web/docs/design.md): Web UI structure と styling conventions。
- [docs/openapi/issue-tracker.yml](docs/openapi/issue-tracker.yml): issue-tracker OpenAPI contract。
- [docs/symphony/README.md](docs/symphony/README.md): Symphony documentation index。
- [docs/symphony/SPEC.md](docs/symphony/SPEC.md): Symphony orchestration and runner specification。
- [docs/symphony/DEVIATIONS.md](docs/symphony/DEVIATIONS.md): Symphony specification からの intentional deviations。

English counterpart: [README.md](README.md).

## Notes

- SQLite files は repository の `.data/` 配下に作成され、git からは無視されます。
- Compose は Go module/build caches と `web/node_modules` を named Docker volumes に保存します。
- Orchestrator は Symphony-oriented runtime settings を `WORKFLOW.md` から読みます。
- Web UI を別 origin から配信する場合は、`NEXT_PUBLIC_ISSUE_TRACKER_URL` で issue-tracker API の origin を指定します。
