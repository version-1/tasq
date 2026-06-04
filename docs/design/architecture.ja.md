# Tasq アーキテクチャ

Tasq は、issue の管理と orchestrator run state の観測を行う local-first のタスクシステムです。

現在のアーキテクチャでは、issue 管理と orchestration を分離します。issue-tracker は issue state と user-facing API を所有します。orchestrator は historical run state と optional runtime inspection を所有します。UI client は主に issue-tracker にアクセスします。Web UI server は future run-state views のために orchestrator proxy path も公開します。

## Goals

- issue state と run state を別概念として保ち、それぞれの owner を分ける。
- web-ui、tui、agent-facing CLI tool が同じ user-facing API surface を使えるようにする。
- tracker API を issue、project、workspace、summary data に集中させる。
- orchestration runtime state を orchestrator local に保つ。

## Non-goals

- hosted multi-tenant operation。
- production authentication と authorization。
- first slice における完全な Codex app-server runner。
- first slice における Linear などの external tracker integration。
- dispatch や worker scheduling semantics。

## Components

### web-ui

web-ui は issue operation のための Go-served Vite + React single-page app です。

Responsibilities:

- issue-tracker から issue summary を取得する。
- issue status、priority、assignee を表示する。
- issue-tracker を呼び出して issue status 間の移動を行う。
- SPA fallback で browser routes を配信する。
- `/api/tracker/*` を issue-tracker に、`/api/orchestrator/*` を orchestrator に proxy する。

Web UI の構造と styling convention は [../../cmd/web/docs/design.md](../../cmd/web/docs/design.md) を参照してください。

### tui

TUI は同じ issue-tracker API のための Go terminal client です。

Responsibilities:

- issue-tracker から issue summary を取得する。
- issue column を描画する。
- one-shot rendering と watch-mode rendering をサポートする。
- orchestrator へ直接アクセスしない。

### tq

`tq` は、HTTP details を埋め込まずに issue state を変更する必要がある agent と workflow tool のための standalone Go CLI です。

Responsibilities:

- issue-tracker API 経由で issue を作成、取得、一覧表示、更新する。
- issue description と comment 用の image attachment を upload する。
- default では human-readable output を使い、tool use 向けに JSON output をサポートする。
- issue-tracker API URL を `--api-url`、`TQ_API_URL`、または `http://localhost:8080` から解決する。
- command が失敗した場合、stderr に machine-readable JSON error を出し、non-zero exit code を返す。
- orchestrator へ直接アクセスしない。

### issue-tracker

issue-tracker は issue management と display aggregation を所有します。

Responsibilities:

- issue を SQLite に保存する。
- 各 issue が必ず 1 つの project に属するようにする。
- issue を作成、編集、一覧表示する。
- attachment metadata を SQLite に保存し、attachment bytes を `$TQ_HOME` 配下に保存する。
- orchestrator や tool の reconciliation 向けに issue state を返す。
- UI/TUI summary API を提供する。

issue-tracker は issue status、priority、title、description、assignee、comment、attachment、project、workspace の source of truth です。
Linked issue が存在する project は削除できません。

### orchestrator

orchestrator は run state と runtime inspection を所有します。

Responsibilities:

- 自身の SQLite database に run record を作成する。
- orchestration 設定に使う repository workflow contract を読み込む。
- configured workspace root 配下に sanitized per-issue workspace を作成する。
- runner event と workspace metadata を記録する。
- runtime state と issue-specific run detail のための optional loopback HTTP API を公開する。

orchestrator は run record、run attempt、runner event、workspace metadata の source of truth です。

### agent

将来の agent は orchestrator に制御される Codex app-server process です。

Responsibilities:

- orchestrator から task を受け取る。
- workspace 内で task を実行する。
- JSON-RPC 経由で execution progress を orchestrator に報告する。

orchestrator は runner boundary を通じて Codex app-server を起動し、local runstore に run progress を記録します。

### workspace

workspace manager は agent のための isolated execution environment を提供します。

Responsibilities:

- git workspace を作成、管理する。
- isolated workspace で parallel execution と verification をサポートする。
- debugging と recovery に必要な metadata を保持する。

現在の workspace manager は sanitized per-issue workspace directory を作成し、新規 workspace を configured repository source から populate し、recovery と debugging のために cleanup/population metadata を記録します。

## Dependency Direction

User-facing client と agent-facing workflow tool は issue-tracker API のみに依存します。

orchestrator は issue-tracker の work queue や event receiver endpoint を使いません。historical run と runner-event data は orchestrator SQLite store に残り、optional orchestrator HTTP API から参照します。

```text
web-ui ─┐
tui ────┼─ issue-tracker ── SQLite: issues, comments, attachments, projects, workspaces
tq ─────┘
                 │
                 └─ $TQ_HOME/system/data/attachments

        orchestrator ───── SQLite: runs, runner_events, workspace metadata
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

orchestrator は issue status を直接変更しません。issue status の変更は issue-tracker の issue API 経由で行います。

## Current MVP Behavior

現在の implementation slice には次が含まれます。

- `cmd/issue-tracker`
- `cmd/tq`
- `cmd/orchestrator`
- issue、project、workspace のための issue-tracker SQLite table。
- issue-tracker attachment metadata in SQLite and image bytes under `$TQ_HOME`。
- run、runner event、workspace metadata、workspace setup failure のための orchestrator SQLite table。
- web-ui と tui が利用する issue-tracker summary API。
- `tq` が利用する issue CRUD API。
- `attachment://<id>` image reference を含む Markdown issue description と comment body。
- Codex runner lifecycle: app-server startup、live-thread turn、enabled 時の continuation turn、terminal run status reporting。

simulated runner は narrow test 用に残しますが、production wiring は Codex app-server runner を使います。
