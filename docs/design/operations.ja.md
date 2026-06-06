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

## Autonomous Approval Boundary

Tasq は、通常の repository work は human-in-the-loop なしで進められる一方、文書化された safety boundary を越える request は Codex `requestApproval` event として残る、自律的な環境で agent を実行するべきです。

目標は approval をなくすことではありません。Approval に意味を持たせることです。基本的な inspection、issue workspace 内の local edit、通常の verification command は configured sandbox 内で実行できるべきです。一方で、より広い command、workspace 外の filesystem access、credential access、network access、その他 host に影響する action は autonomous boundary の外に残し、approval-required work として表面化させます。

Local development では、baseline environment を次のようにします。

- Codex app-server を Tasq の `dev` container 内で実行する。
- [ADR-0006](../adr/0006-run-app-server-with-workspace-write-sandbox.ja.md) の documented workspace-write sandbox posture で Codex を起動する。
- Codex authentication と personal state は repository ではなく `codex-home` Docker volume に保持する。
- Repository-managed Codex rules は `codex/rules/` から read-only mount する。
- Tasq runtime state は `$TQ_HOME` 配下に置き、Codex credential と分離する。
- command-execution と file-change の `requestApproval` event は [ADR-0005](../adr/0005-block-issues-for-app-server-approval-decisions.ja.md) に従って扱う。つまり request を cancel し、run を `approval_required` で failed にし、request details 付きで issue を blocked にする。

これにより、Tasq は安全な unattended default を持てます。Agent は container と workspace boundary の内側では進められますが、Tasq はより広い trust decision が必要な action を暗黙に approve しません。

Operator は blocked issue を out of band に確認します。Request を許可できる場合、operator は environment を変更する、より狭い rule を追加する、または明示的な approval decision を行ったうえで issue を retry できます。Tasq は、requested action が文書化された policy のもとで実際に実行されていない限り、blocked approval を success に変換してはいけません。

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

Web server は `/tracker/*` を issue-tracker に、`/orchestrator/*` を orchestrator に proxy します。Compose では `make run-web` が dev container 内の `127.0.0.1:8080` と `127.0.0.1:8081` を backend URL として Web server を起動します。

## Open Decisions

- external tracker sync を issue-tracker 内に置くか provider interface の behind に置くか。
- Production authentication、authorization、network exposure。
- large full-fidelity Codex transcript を SQLite に残すか filesystem artifact に移すか。
