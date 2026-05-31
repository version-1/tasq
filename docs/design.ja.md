# Tasq 設計

Tasq は、issue の管理、実行可能な作業の coding agent への割り当て、agent run 状態の Web UI と TUI からの観測を行う local-first のタスク実行システムです。

現在のアーキテクチャでは、issue 管理と orchestration を分離します。issue-tracker は issue state と user-facing API を所有します。orchestrator は agent run state と作業割り当て状態を所有します。UI client は issue-tracker のみにアクセスします。

## Goals

- issue state と run state を別概念として保ち、それぞれの owner を分ける。
- web-ui と tui が同じ user-facing API surface を使えるようにする。
- 複数 orchestrator instance が並列に動いても安全に work assignment できるようにする。
- issue-tracker が一時的に利用できない場合でも run state change を保持する。
- 実 Codex app-server runner を追加する前に検証できる小さな first implementation slice に保つ。

## Non-goals

- hosted multi-tenant operation。
- production authentication と authorization。
- first slice における完全な Codex app-server runner。
- first slice における Linear などの external tracker integration。
- SQLite-backed issue-tracker work item queue を超える distributed queue。

## Components

### web-ui

web-ui は issue operation のための Next.js client です。

Responsibilities:

- issue-tracker から issue summary を取得する。
- issue status、priority、assignee、latest run state を表示する。
- issue-tracker を呼び出して issue status 間の移動を行う。
- orchestrator へ直接アクセスしない。

Web UI の構造と styling convention は [../web/docs/design.md](../web/docs/design.md) を参照してください。

### tui

TUI は同じ issue-tracker API のための Go terminal client です。

Responsibilities:

- issue-tracker から issue summary を取得する。
- issue column と latest run state を描画する。
- one-shot rendering と watch-mode rendering をサポートする。
- orchestrator へ直接アクセスしない。

### issue-tracker

issue-tracker は issue management と display aggregation を所有します。

Responsibilities:

- issue を SQLite に保存する。
- issue を作成、編集、一覧表示する。
- issue が executable になるタイミングを決定する。
- issue が ready to run になったときに work item を作成する。
- orchestrator 向けの lease-backed work item claim API を提供する。
- orchestrator run event を idempotent に受け取る。
- run fact に基づいて issue status transition を適用する。
- UI/TUI summary API を提供する。

issue-tracker は issue status、priority、title、description、assignee、work item claim state、受信済み orchestrator event id の source of truth です。

### orchestrator

orchestrator は agent assignment と run state を所有します。

Responsibilities:

- issue-tracker work item queue を poll する。
- executable work item を lease 付きで claim する。
- 自身の SQLite database に run record を作成する。
- orchestration 設定に使う repository workflow contract を読み込む。
- configured workspace root 配下に sanitized per-issue workspace を作成する。
- durable outbox を通じて run state change を emit する。
- issue-tracker に受け入れられるまで outbox delivery を retry する。
- MVP では boundary を検証できる最小限の run lifecycle を simulate する。

orchestrator は run record、run attempt、run に紐づく claim token、outbox delivery state の source of truth です。

### agent

将来の agent は orchestrator に制御される Codex app-server process です。

Responsibilities:

- orchestrator から task を受け取る。
- workspace 内で task を実行する。
- JSON-RPC 経由で execution progress を orchestrator に報告する。

MVP では実 agent process は起動しません。issue-tracker/orchestrator contract を先に検証するため、最小限の run lifecycle のみを記録します。

### workspace

workspace manager は agent のための isolated execution environment を提供します。

Responsibilities:

- git workspace を作成、管理する。
- isolated workspace で parallel execution と verification をサポートする。
- debugging と recovery に必要な metadata を保持する。

現在の workspace manager は sanitized per-issue workspace directory を作成し、その path を run state に記録します。terminal cleanup と repository population は今後の作業です。

## Dependency Direction

User-facing client は issue-tracker API のみに依存します。

orchestrator は issue-tracker の work queue と event receiver API に依存します。issue-tracker は orchestrator を poll しません。代わりに、orchestrator push event から latest run snapshot を保存します。

```text
web-ui ─┐
        ├─ issue-tracker ── SQLite: issues, work_items, received_events, run_snapshots
tui ────┘       ▲
                │ claim work item / push run event
                │
        orchestrator ───── SQLite: runs, outbox_events
                │
                ├─ future: agent-runner ── Codex app-server over JSON-RPC
                └─ workspace manager ── git workspace / isolated runtime
```

## State Ownership

Issue status と run status は別物です。

Issue status は issue-tracker が所有します。

- `backlog`
- `ready`
- `in_progress`
- `review`
- `done`
- `blocked`
- `failed`

Run status は orchestrator が所有します。

- `queued`
- `starting`
- `running`
- `waiting_for_input`
- `succeeded`
- `failed`
- `cancelled`

orchestrator は issue status を直接変更しません。orchestrator は run fact を emit します。issue-tracker はその fact を受け取り、issue status rule を適用します。

## Work Item Queue

issue status が `ready` に変更されると、issue は executable になります。

issue が `ready` になると、issue-tracker は pending work item を作成します。orchestrator は全 issue を scan せず、issue が executable かどうかも判断しません。work item queue のみを poll します。

同じ issue を再実行すると新しい work item が作成されます。これにより claim token、run attempt、result が単一の execution request に紐づきます。

## Claim And Lease

Work item claim は lease-backed です。

orchestrator が work item を claim すると、issue-tracker は次の値を記録します。

- `claimed_by`
- `claim_token`
- `lease_until`
- incremented attempt count

orchestrator が停止した場合、または将来の実装で renewal が止まった場合、work item は `lease_until` 後に再び claimable になります。

claim token は 1 つの work item claim に対する generation marker です。orchestrator run event は、その claim token が work item の current claim token と一致する場合のみ適用されます。expired claim からの late event は idempotency のため記録されますが、issue state を更新することはできません。

## Run Events And Outbox

orchestrator は run event を issue-tracker に送信する前に、自身の SQLite outbox に書き込みます。

issue-tracker は run event を idempotent に受け入れます。

- 各 event は一意な `eventId` を持つ。
- 処理済み event id は SQLite に保存される。
- 重複 event id は already accepted として扱われる。

これにより orchestrator は state transition を二重適用せずに delivery を retry できます。

## Current MVP Behavior

現在の implementation slice には次が含まれます。

- `cmd/issue-tracker`
- `cmd/orchestrator`
- issue、work item、received orchestrator event、run snapshot のための issue-tracker SQLite table。
- run と outbox event のための orchestrator SQLite table。
- web-ui と tui が利用する issue-tracker summary API。
- lease-backed work item claim API。
- idempotent run event receiver。
- orchestrator の polling、claim、run creation、outbox delivery。
- 最小限の simulated run lifecycle: `queued -> running -> succeeded`。

simulated lifecycle は意図的に一時的なものです。目的は Codex app-server runner、terminal workspace cleanup、repository population を追加する前に service boundary を検証することです。

## API Surface

issue-tracker は user-facing API です。

現在の issue-tracker endpoint:

- `GET /api/v1/health`
- `GET /api/v1/summary`
- `GET /api/v1/projects`
- `POST /api/v1/projects`
- `GET /api/v1/projects/{id}`
- `PATCH /api/v1/projects/{id}`
- `DELETE /api/v1/projects/{id}`
- `GET /api/v1/workspaces`
- `POST /api/v1/workspaces`
- `GET /api/v1/workspaces/{id}`
- `PATCH /api/v1/workspaces/{id}`
- `DELETE /api/v1/workspaces/{id}`
- `GET /api/v1/issues`
- `POST /api/v1/issues`
- `GET /api/v1/issues/{id}`
- `PATCH /api/v1/issues/{id}`
- `POST /api/v1/work-items/claim`
- `POST /api/v1/orchestrator-events`

JSON success response は `{ "data": ..., "meta": {} }` を使います。JSON error response は `{ "error": { "code": "...", "message": "..." }, "meta": {} }` を使います。

orchestrator は現在 user-facing HTTP API を持ちません。外部依存は issue-tracker API です。

## Development Environment

Docker Compose は issue-tracker を container port `8080`、web-ui を container port `3000`、orchestrator を background worker として実行します。

Recommended commands:

- `make issue-tracker-up`
- `make orchestrator-up`
- `make dev-up`
- `make dev-up-forward`
- `make web-up`
- `make tui-up`
- `make dev-status`

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

Manual MVP verification:

1. issue-tracker と orchestrator を起動する。
2. status `ready` の issue を作成する。
3. issue-tracker が work item を作成することを確認する。
4. orchestrator が claim し、run event を emit することを確認する。
5. issue-tracker summary が latest run status `succeeded` の issue を `review` として表示することを確認する。

## Open Decisions

- 正確な Codex app-server JSON-RPC contract。
- Agent runner process lifecycle と cancellation behavior。
- long-running real agent の lease renewal cadence。
- Workspace cleanup policy と repository population strategy。
- Retry limit と manual intervention threshold。
- external tracker sync を issue-tracker 内に置くか provider interface の behind に置くか。
- Production authentication、authorization、network exposure。
- run log を SQLite に直接保存するか filesystem artifact として参照するか。
