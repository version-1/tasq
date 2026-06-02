# tasq

実行可能な作業を管理し、coding agents に割り当て、Web UI または TUI から進捗を確認するための local-first issue tracker and task orchestrator です。

## Components

- Issue Tracker: SQLite backed の Go REST API です。issues、comments、projects、workspaces、UI summaries を所有します。
- Orchestrator: SQLite backed の Go service です。runtime inspection 用に run state と runner events を記録します。
- `tq`: agent と workflow tool が issue-tracker API 経由で issue を作成、取得、一覧表示、更新するための Go CLI です。
- Web UI: issue-tracker API 用の Next.js client です。
- TUI: 同じ issue-tracker API を使う Go terminal client です。

Architecture 全体は [docs/design.md](docs/design.md) を参照してください。
Local configuration は [docs/configuration.ja.md](docs/configuration.ja.md) を参照してください。

## Quick Start

Local development では `make` 経由で Docker Compose を使います。

```sh
make dev-up
```

このコマンドは `dev` container と OpenAPI UI を起動し、`dev` container 内で issue-tracker、orchestrator、Web UI を起動します。Docker Compose が host ports を自動割り当てし、割り当てられた URLs を表示します。

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

### Linux/WSL2 Sandbox Prerequisite

Codex は Linux sandboxing に Bubblewrap を使います。Dev image は `bubblewrap` を
install しますが、Codex の sandboxed command を安定して動かすには Linux / WSL2 host
側でも unprivileged user namespace creation が許可されている必要があります。Codex が
`bwrap: No permissions to create a new namespace` を報告する場合は、image に package が
ないだけではなく、host または Docker runtime の capability issue として扱います。

## Verification

標準の Compose-backed checks を実行します。

```sh
make dev-test
```

Go services と Web UI の両方に影響する変更を handoff する前は、broader build check を実行します。

```sh
make dev-build
```

## Documentation

- [docs/development.ja.md](docs/development.ja.md): repository workflow、task flow、documentation update rules、component workflow links。
- [WORKFLOW.md](WORKFLOW.md): orchestrator が使う Symphony runtime workflow contract。
- [docs/design.md](docs/design.md): system architecture と service boundaries。
- [docs/references/makefile.ja.md](docs/references/makefile.ja.md): Makefile targets、variables、local development command reference。
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

- Runtime state と SQLite files は repository の `.tasq/` 配下に作成され、git からは無視されます。
- Compose は Go module/build caches、`web/node_modules`、Codex login state、GitHub CLI login state を named Docker volumes に保存します。
- Orchestrator は Symphony-oriented runtime settings と issue ごとの agent prompt を `WORKFLOW.md` から読みます。
- Web UI を別 origin から配信する場合は、`NEXT_PUBLIC_ISSUE_TRACKER_URL` で issue-tracker API の origin を指定します。
- `make run-tq` 経由で実行した `tq` は `$TQ_HOME/system/state.json` から issue-tracker API URL を解決します。
- Codex を device auth で認証し、authentication を `codex-home` Docker volume に永続化するため、初回に `make dev-codex-login` を実行します。
- GitHub CLI を認証し、credential を `gh-config` Docker volume に永続化するため、初回に `make dev-gh-login` を実行します。Dev container から push する場合は HTTPS Git remote を使います。
- Codex または GitHub access が必要な agent workflow を実行する前に、`make dev-codex-status` と `make dev-gh-status` で dev container が認証済みであることを確認します。

## tq CLI

issue を一覧表示します。

```sh
make run-tq ARGS="issue list"
```

issue を取得します。

```sh
make run-tq ARGS="issue get 1"
```

issue を作成します。

```sh
make run-tq ARGS='issue create --title "Wire Symphony workflow" --description "Define the first workflow contract" --status ready --priority high'
```

よく使う status / text update には issue shortcut を使えます。

```sh
make run-tq ARGS="issue ready 1"
make run-tq ARGS="issue close 1"
make run-tq ARGS='issue rename 1 "Clarify workflow contract"'
make run-tq ARGS='issue edit 1 "Updated description"'
```

machine-readable output が必要な場合は `--output json` を使います。

```sh
make run-tq ARGS="--output json issue list"
```
