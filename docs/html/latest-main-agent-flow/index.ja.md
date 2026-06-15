# Latest Main Agent Dispatch Flow

このメモは `docs/html/latest-main-agent-flow/index.html` の日本語版要約です。
対象は `origin/main` commit `36437c9` 時点の Tasq Symphony orchestrator です。
`docs/html/index.ja.html` から、他の HTML ビジュアルガイドと合わせて参照できます。

## 目的

最新の `main` を元に agent を起動し、issue tracker の ready issue を run として割り振り、Codex app-server の turn が完了して run が terminal status になるまでの流れを説明します。

## 全体像

- `cmd/orchestrator/main.go` が workflow、run store、workspace manager、tracker client、dispatcher、poller を組み立てます。
- `Poller.Poll` が issue-tracker API から `ready` 状態の issue を取得します。
- `workspace.Manager` が issue ごとの workspace を作成または再利用します。
- `runstore.Store` が `runs` に `queued` run を作り、`runner_events` に queue event を記録します。
- `Dispatcher.Dispatch` が queued run を claim し、concurrency limit の範囲で goroutine を起動します。
- `CodexRunner` が workspace を cwd として `codex app-server` を起動し、JSON-RPC over stdio で thread と turn を開始します。
- 対象 turn の `turn/completed` notification を受け取ると runner は `succeeded` を返し、dispatcher が `runs.status` を terminal status に更新します。

## 重要な補足

Agent 自身が `runs.status` を直接更新するわけではありません。
Agent / Codex app-server から届くのは protocol event です。
`turn/completed` が届けば orchestrator 側の `CodexRunner` が `succeeded` を返し、error notification、timeout、process exit、startup failure などは `failed` に変換されます。
最終的な `UpdateRunStatus(succeeded|failed|cancelled)` を実行するのは `Dispatcher.startRun` です。

## 最新 main から起動する条件

現在の実装では、orchestrator が各 workspace で `git fetch` や `git checkout main` を実行するわけではありません。
`workspace.source` に指定されたディレクトリを per-issue workspace へコピーします。
そのため「最新 main から agent を起動する」には、`workspace.source` が最新 `main` の checkout を指している必要があります。

コピー時には次のディレクトリは除外されます。

- `.git`
- `.worktrees`
- `node_modules`

## 図に含めた内容

- Architecture: issue tracker、orchestrator、workspace manager、run store、Codex app-server の境界
- End-to-End Flow: startup、poll、queue、dispatch、agent turn、terminal update の時系列
- Interactive Timeline: responsibility が startup、queueing、dispatch、agent のどこにあるかを step 表示
- Run State: `queued`、`running`、`succeeded`、`failed`、`cancelled` の遷移
- Contracts: workflow config、runner task、store、workspace、Codex app-server の入出力
- Gaps and Notes: origin/main の実装と `docs/symphony/SPEC.md` の差分

## 参照した主なファイル

- `origin/main:cmd/orchestrator/main.go`
- `origin/main:internal/orchestrator/coordinator/poller.go`
- `origin/main:internal/orchestrator/coordinator/dispatcher.go`
- `origin/main:internal/orchestrator/runner/runner.go`
- `origin/main:internal/orchestrator/workspace/workspace.go`
- `origin/main:internal/orchestrator/workflow/workflow.go`
- `origin/main:db/schema/orchestrator.sql`
- `docs/symphony/SPEC.md`

## 外部依存

HTML 図は CDN を使っていません。
inline SVG と inline JavaScript だけで構成しているため、ローカルのブラウザで直接開いて読めます。
