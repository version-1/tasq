# ADR-0004: Unsupported app-server approval では実行を failed にする

## Context

Tasq は Codex を app-server JSON-RPC protocol 経由で実行する。turn の途中で、Codex は `item/commandExecution/requestApproval` や `item/fileChange/requestApproval` のような server-to-client 承認要求を送ることがある。

Tasq は現時点で対話型の承認フローを実装していない。orchestrator は無人で動くため、運用者の代わりにコマンドやファイル変更を承認する判断を安全に行えない。また、必要な承認を拒否したあとに成功を返すのも危険である。要求されたコマンドやファイル変更が実行されていないのに、課題が完了したように見えるためである。

Codex app-server は承認要求に対して Tasq が unsupported JSON-RPC error を返したあとでも、`turn/completed` を発行することがある。そのため、`turn/completed` だけを成功として扱うと、必要な action が拒否された事実を隠してしまう。

## Decision

Tasq はコマンド実行とファイル変更の承認要求を unsupported approval failure として扱う。

runner が `item/commandExecution/requestApproval` または `item/fileChange/requestApproval` を受け取った場合、app-server には unsupported JSON-RPC error を返し、アクティブな turn が unsupported approval を必要としたことを記録する。同じ turn が後から `turn/completed` を発行しても、runner は成功ではなく failed run result を返す。

dispatcher は既存の failed-run path を使う。最新の課題状態がまだ `ready` の場合、dispatcher は課題を `blocked` に更新し、runner の失敗理由を含む blocker comment を作成する。

## Alternatives

### `turn/completed` を成功として扱う

Protocol handling は単純になるが、必要なコマンドやファイル変更が拒否されたあとでも run を `succeeded` にできてしまう。未完了の作業を隠し、課題状態を誤解させる。

### Operator approval を待って stall する

人による承認の可能性は残せるが、Tasq にはまだ approval UI、operator routing、永続的な承認状態がない。無期限に待つと無人オーケストレーションの信頼性が落ちる。

### App-server approval request を auto-approve する

信頼済み環境では task completion が改善する可能性がある。しかし信頼境界が広がるため、コマンド範囲、ファイル範囲、認証情報、サンドボックス、監査可能性を含む別のセキュリティ判断が必要である。この slice ではその判断を行わない。

## Consequences

Unsupported required approval は明確に失敗する。後続の `turn/completed` notification は、Tasq が必須 action を拒否した事実を上書きしない。

Ready issue は既存の dispatcher の挙動により `blocked` へ移り、blocker comment が追加される。これにより失敗がユーザーに見え、poller が同じ課題を ready のまま即座に再 queue することを防ぐ。

人による承認があれば完了できた task も、Tasq が承認ポリシーを実装するまでは失敗する。これは意図的な挙動であり、false success よりも明示的な blocked state を優先する。

App-server approval posture は保守的なまま維持する。将来、対話型承認や scoped auto-approval を追加する場合は、別 ADR として記録する。

## Notes

この ADR は `docs/symphony/SPEC.md` が求める承認姿勢を記録する。approval と user-input behavior は implementation-defined だが、承認要求によって run を無期限に stalled にしてはならない。

関連する実装メモは `docs/symphony/CODEX_APP_SERVER.md` にある。
