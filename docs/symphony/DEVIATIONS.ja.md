# Tasq Symphony 差分

このドキュメントは、Tasq と上流の Symphony service specification
[SPEC.md](SPEC.md) の意図的な差分を記録します。

Tasq は orchestration、workspace、workflow、agent-runner、tracker、observability の方向性として
Symphony specification を使います。ただし Tasq には issue state を所有する local issue-tracker service
が既にあるため、一部の実装方針は異なります。

## Tracker Adapter

Tasq は Symphony specification の Section 11 で説明されている Linear tracker client を実装しません。

代わりに、Tasq は local issue-tracker API を tracker adapter boundary として扱います。

- issue-tracker は issue state と project data を所有します。
- orchestrator は workspace records、metadata、lifecycle behavior を所有します。
- issue-tracker は tracker adapter reads のために issue listing と issue-state query endpoints を公開します。
- orchestrator は historical run と runner-event data を自身の SQLite store に保持します。

これにより、external tracker integrations を orchestrator から外し、repository の既存 service boundary
を維持します。

## Workflow Front Matter Contract

Tasq の `WORKFLOW.md` front matter は、Symphony front matter schema の上に重ねた小さな
Tasq-specific contract として意図的に定義されています。Tasq の canonical workflow contract は
[WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md) に記録されています。次の表は、Symphony `SPEC.md`
の fields と Tasq front matter behavior の対応を示します。

| Top-level key | Child field | Symphony Support | Tasq Support | Tasq Extension | Tasq front matter / behavior |
| --- | --- | --- | --- | --- | --- |
| `tracker` | `kind` | ✓ | ✓ | × | 現在の orchestration は Linear client ではなく local issue-tracker API から読み取ります。 |
|  | `endpoint` | ✓ | ✓ | × | Symphony tracker shape との partial compatibility のために parse されます。 |
|  | `api_key` | ✓ | ✓ | × | `$VAR` resolution をサポートします。 |
|  | `project_slug` | ✓ | ✓ | × | `tracker.kind` が設定された場合のみ必須です。 |
|  | `active_states` | ✓ | ✓ | × | Tasq の dispatch eligibility は local issue-tracker states によって決まります。 |
|  | `terminal_states` | ✓ | ✓ | × | Tasq の cleanup/reconciliation は local issue-tracker states を使います。 |
| `polling` | `interval_ms` | ✓ | ✓ | × | Orchestrator polling interval です。 |
| `workspace` | `root` | ✓ | ✓ | × | 選択された `WORKFLOW.md` からの相対 path として解決され、worktree management のため target Git repository の内側である必要があります。 |
|  | `source` | × | × | × | Tasq は代わりに `workspace.root` 配下に Git worktree を作成します。 |
| `hooks` | `after_create` | ✓ | ✓ | × | 新規作成された workspace に対してのみ、issue workspace 内で `bash -lc` 経由で実行されます。 |
|  | `before_run` | ✓ | ✓ | × | 各 agent attempt の前に実行されます。 |
|  | `after_run` | ✓ | ✓ | × | 各 agent attempt 後に実行され、失敗は log されて無視されます。 |
|  | `before_remove` | ✓ | ✓ | × | workspace directory が存在する場合に cleanup 前に実行され、失敗は log され cleanup は継続します。 |
|  | `timeout_ms` | ✓ | ✓ | × | すべての workspace hooks に適用され、正の値である必要があります。 |
| `agent` | `max_concurrent_agents` | ✓ | ✓ | × | Global concurrent run limit です。 |
|  | `max_turns` | ✓ | ✓ | × | 1 run の最大 Codex turns です。Continuation turns には `agent.continuation_turns_enabled` も必要です。 |
|  | `max_retry_backoff_ms` | ✓ | ✓ | × | Retry backoff cap です。 |
|  | `max_concurrent_agents_by_state` | ✓ | ✓ | × | 正規化された map です。意味を持つかは local issue-tracker state model に依存します。 |
|  | `continuation_turns_enabled` | × | ✓ | ✓ | `agent.max_turns` が 1 より大きい場合でも continuation turns を gate します。 |
|  | `max_retry_attempts` | × | ✓ | ✓ | Run retry attempts の回数を設定します。 |
| `codex` | `command` | ✓ | ✓ | × | repository root の `WORKFLOW.md` は `codex --sandbox workspace-write app-server` を使います。 |
|  | `approval_policy` | ✓ | ✓ | × | repository root の `WORKFLOW.md` では設定されていません。 |
|  | `thread_sandbox` | ✓ | ✓ | × | repository root の `WORKFLOW.md` では設定されていません。 |
|  | `turn_sandbox_policy` | ✓ | ✓ | × | repository root の `WORKFLOW.md` では設定されていません。 |
|  | `turn_timeout_ms` | ✓ | ✓ | × | Codex turn timeout です。 |
|  | `read_timeout_ms` | ✓ | ✓ | × | Codex app-server read timeout です。 |
|  | `stall_timeout_ms` | ✓ | ✓ | × | Tasq は正の値として検証します。 |
| `server` | `port` | ✓ | ✓ | × | Optional HTTP extension port です。CLI `--port` で override できます。 |
| `tasq` | `task_work_prompt` | × | ✓ | ✓ | orchestrator が rendered prompt の前に default `tq` issue-tracker synchronization instructions を付与するかを制御します。 |

Table notes:

- `Symphony Support` は、その field が `SPEC.md` の core field または documented extension field として定義されていることを表します。
- `Tasq Support` は、Tasq がその field を `WORKFLOW.md` front matter contract で parse または実装していることを表します。
- `Tasq Extension` は、その field が Tasq-specific であり、Symphony core schema の一部ではないことを表します。
- `tracker.kind` は Symphony dispatch では必須です。Tasq は parse しますが、local dispatch では使いません。
- Symphony では `codex.stall_timeout_ms <= 0` によって stall detection を無効化できます。Tasq はこの field を正の値として検証します。

Symphony との差分:

- `WORKFLOW.md` は orchestrator process startup 時に一度だけ load されます。Dynamic watch/reload と、
  変更された settings の runtime re-application は延期されています。
- Unknown front matter fields は forward compatibility のために無視されます。
- `workspace.source` はサポートされません。Tasq は `workspace.root` 配下に Git worktree で issue
  workspaces を作成します。
- Large transcript artifact paths と observability sinks は workflow front matter では設定できません。
  Tasq は runner progress、workspace metadata、workspace setup failures、cleanup state を orchestrator
  SQLite database に記録します。
- local `tq project check` command は full Symphony schema ではなく、Tasq の default project template
  が要求する front matter fields を検証します。

## Workspace Key

Symphony は `MT-649` のような `issue.identifier` から workspace keys を導出します。
Tasq は canonical issue identifier として `issue-<ID>` を意図的に使い、workspace naming にも同じ値を
使います。例: `issue-42`。

理由:

- Tasq の local issue-tracker は stable numeric IDs を持つ issues を所有します。
- Tasq は現在、human-readable external tracker identifier を別 field として model していません。
- `issue-<ID>` を使うことで、Linear-specific fields を local issue contract に追加せずに workspace paths
  を deterministic に保てます。

同じ workspace key convention は、workspace metadata、startup terminal cleanup、active-run
reconciliation でも使われます。

## Current Implementation Gap

現在の orchestrator は Symphony conformance に向けて段階的に移行しています。まだ完全な Symphony
implementation ではありません。

実装済み、または進行中:

- Symphony front matter の小さな supported subset を使った workflow file loading。
- `WORKFLOW.md` は process startup 時にのみ load されます。runtime reload は意図的に延期されています。
- Workspace root resolution と sanitized per-issue workspace directories。
- `hooks.timeout_ms` を含む workspace lifecycle hooks。
- simulated implementation と Codex app-server subprocess implementation を持つ runner interface。
- SQLite runner event logging と workspace metadata records。
- live Codex app-server thread に対する config-gated continuation turns。
- capped exponential backoff を使う in-process retry scheduling。
- terminal/non-active issue states と stall handling のための active-run reconciliation。
- 初回 workspace creation 時の Git worktree workspace creation。
- cleanup metadata を伴う terminal および failed/cancelled workspace cleanup。
- workspace setup failures の operator-facing logs。

未実装:

- Dynamic `WORKFLOW.md` reload。
- full variable と filter checking を伴う strict prompt rendering。
- Token/rate-limit accounting。
- Full optional Symphony HTTP status/API surface。

Tasq は [WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md) に記録されている workflow front matter fields
をサポートします。Unknown fields は forward compatibility のために無視されます。

## Workspace Creation Strategy

Tasq は repository source directory から files を copy するのではなく、`git worktree add` で per-issue
orchestrator workspaces を作成します。

理由:

- coding agent には per-issue workspace が Git repository root である必要があります。
- `.git` なしで files を copy すると、repository inspection commands が別の Git root を観測し、edits
  が issue workspace ではなく parent repository を対象にする可能性があります。
- worktree は Git metadata を維持しつつ、issue project の `Project.Location` と configured relative
  `workspace.root` 配下の deterministic workspace path を保ちます。

`workspace.source` workflow field は意図的にサポートされません。Workspace manager が
relative workspace-root suffix を解決できるように、`workspace.root` は orchestrator project の Git
repository の内側でなければなりません。Tasq は dispatched issue ごとに `Issue.ProjectID` を local
issue-tracker API で解決し、同じ relative suffix を使って referenced project location 配下に worktree
を作成します。例: `<Project.Location>/.worktrees/agents/issue-42`。

Workspace branches は `agent/<workspace-key>` を使います。例: `agent/issue-42`。Cleanup は
`git worktree remove --force` を使い、対応する local branch を best-effort で削除し、orchestrator
startup 時に `git worktree prune` を実行します。

## Compatibility Notes

Symphony が scheduling に関して "tracker" と言う箇所は、将来の design で external tracker adapter
を明示的に追加しない限り、Tasq の orchestrator では local issue-tracker API と読み替えます。

Symphony が Linear-specific query semantics を説明している箇所は、local issue-tracker boundary が選択された
tracker adapter である間、orchestrator には適用されない requirement として扱います。
