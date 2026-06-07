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

Tasq の `WORKFLOW.md` front matter は、Symphony front matter schema の小さな Tasq-specific subset
として意図的に定義されています。Tasq の canonical workflow contract は
[WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md) に記録されています。

Tasq がサポートする fields:

- `tasq.task_work_prompt` は Tasq extension です。orchestrator が rendered prompt の前に default `tq`
  issue-tracker synchronization instructions を付与するかを制御します。
- `polling.interval_ms`、`workspace.root`、`agent.max_concurrent_agents`、`agent.max_turns`、
  `agent.continuation_turns_enabled`、`agent.max_retry_attempts`、
  `agent.max_retry_backoff_ms`、`codex.command`、`codex.read_timeout_ms`、
  `codex.turn_timeout_ms`、`codex.stall_timeout_ms`、`server.port`、workspace lifecycle
  `hooks` をサポートします。
- `hooks.after_create`、`hooks.before_run`、`hooks.after_run`、`hooks.before_remove`、
  `hooks.timeout_ms` をサポートします。Hook scripts は issue workspace を working directory として
  `bash -lc` 経由で実行されます。
- `tracker.*` fields は partial Symphony compatibility のために parse されます。ただし現在の Tasq
  orchestrator は、Symphony が説明する Linear tracker client ではなく local issue-tracker API から work
  を読み取ります。
- Codex pass-through fields (`codex.approval_policy`、`codex.thread_sandbox`、
  `codex.turn_sandbox_policy`) は parse されますが、この repository の現在の root `WORKFLOW.md`
  template では使われていません。

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
- worktree は Git metadata を維持しつつ、`workspace.root` 配下の deterministic workspace path を保ちます。

`workspace.source` workflow field は意図的にサポートされません。Workspace manager が
`git rev-parse --show-toplevel` で repository を解決できるように、`workspace.root` は target Git
repository の内側でなければなりません。

Workspace branches は `agent/<workspace-key>` を使います。例: `agent/issue-42`。Cleanup は
`git worktree remove --force` を使い、対応する local branch を best-effort で削除し、orchestrator
startup 時に `git worktree prune` を実行します。

## Compatibility Notes

Symphony が scheduling に関して "tracker" と言う箇所は、将来の design で external tracker adapter
を明示的に追加しない限り、Tasq の orchestrator では local issue-tracker API と読み替えます。

Symphony が Linear-specific query semantics を説明している箇所は、local issue-tracker boundary が選択された
tracker adapter である間、orchestrator には適用されない requirement として扱います。
