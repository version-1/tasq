# ADR-0004: Unsupported app-server approval では run を failed にする

## Context

Tasq は Codex を app-server JSON-RPC protocol 経由で実行する。turn の途中で、Codex は `item/commandExecution/requestApproval` や `item/fileChange/requestApproval` のような server-to-client approval request を送ることがある。

Tasq は現時点で interactive approval flow を実装していない。orchestrator は unattended に動くため、operator の代わりに command や file change を approve する判断を安全に行えない。また、必要な approval を decline したあとに success を返すのも危険である。requested command や file change が実行されていないのに、issue が完了したように見えるためである。

Codex app-server は approval request に対して Tasq が unsupported JSON-RPC error を返したあとでも、`turn/completed` を emit することがある。そのため、`turn/completed` だけを success として扱うと、必要な action が declined された事実を隠してしまう。

## Decision

Tasq は command-execution と file-change の approval request を unsupported approval failure として扱う。

runner が `item/commandExecution/requestApproval` または `item/fileChange/requestApproval` を受け取った場合、app-server には unsupported JSON-RPC error を返し、active turn が unsupported approval を必要としたことを記録する。同じ turn が後から `turn/completed` を emit しても、runner は success ではなく failed run result を返す。

dispatcher は既存の failed-run path を使う。最新の issue state がまだ `ready` の場合、dispatcher は issue を `blocked` に更新し、runner failure reason を含む blocker comment を作成する。

## Alternatives

### `turn/completed` を success として扱う

Protocol handling は単純になるが、必要な command や file change が declined されたあとでも run を succeeded にできてしまう。未完了の作業を隠し、issue state を誤解させる。

### Operator approval を待って stall する

Human approval の可能性は残せるが、Tasq にはまだ approval UI、operator routing、durable approval state がない。無期限に待つと unattended orchestration の信頼性が落ちる。

### App-server approval request を auto-approve する

Trusted environment では task completion が改善する可能性がある。しかし trust boundary が広がるため、command scope、file scope、credential、sandbox、auditability を含む別の security decision が必要である。この slice ではその判断を行わない。

## Consequences

Unsupported required approval は明確に失敗する。後続の `turn/completed` notification は、Tasq が required action を declined した事実を上書きしない。

Ready issue は既存の dispatcher behavior により `blocked` へ移り、blocker comment が追加される。これにより failure が user に見え、poller が同じ issue を ready のまま即座に再 queue することを防ぐ。

Human approval があれば完了できた task も、Tasq が approval policy を実装するまでは失敗する。これは意図的な挙動であり、false success よりも explicit blocked state を優先する。

App-server approval posture は conservative なまま維持する。将来、interactive approval や scoped auto-approval を追加する場合は、別 ADR として記録する。

## Notes

この ADR は `docs/symphony/SPEC.md` が求める approval posture を記録する。approval と user-input behavior は implementation-defined だが、approval request によって run を indefinitely stalled にしてはならない。

関連する implementation note は `docs/symphony/CODEX_APP_SERVER.md` にある。
