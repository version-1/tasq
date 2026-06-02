# ADR-0005: App-server approval decision のために issue を blocked にする

## Context

ADR-0004 は Codex app-server approval request に対する conservative posture を記録した。command-execution と file-change の approval は、Tasq が requested action を完了できない場合に success として扱ってはならない。

しかし、その decision だけでは operator workflow が不十分である。Operator は requested approval details を確認し、yes/no を判断し、その明示的な approval decision を使って issue を retry する必要がある。Request を generic unsupported protocol failure として扱うと、context が不足し、次に operator が何をすべきか分かりにくい。

対象 Codex app-server schema では、approval response は `{"decision":"accept"}`、`{"decision":"decline"}`、`{"decision":"cancel"}` と related accepted variants として定義されている。Generated schema には `approved: false` response field はない。

## Decision

Tasq は `item/commandExecution/requestApproval` と `item/fileChange/requestApproval` を known approval request として扱う。

現在の unattended runner では、Tasq はこれらの app-server request に即時 `{"decision":"cancel"}` を返す。これは protocol-level denial であり、同時に Codex に current turn の interrupt を求める。runner はその後、run を `approval_required` error 付きの `failed` として terminal にする。

最新の issue state がまだ runnable（`ready` または `in_progress`）の場合、dispatcher は issue を `blocked` に更新し、approval method と raw request payload を含む blocker comment を作成する。Issue status は SPEC-compatible な `blocked` のままとし、Tasq は新しい `blocked_by_approval` issue status を導入しない。

Human approval は現時点では out of band に扱う。Operator は blocked issue と comment を確認し、その request を許可できるか判断し、retry のために issue を `ready` へ戻せる。将来の approval decision store では、次の run が matching request だけを auto-approve できるようにするべきである。

## Alternatives

### Unsupported JSON-RPC error を返し続ける

単純だが、known approval request が unknown protocol request のように見える。また blocker comment も operator 判断に必要な情報として弱い。

### `decline` を返す

`decline` は valid denial response である。ただし schema 上は agent が turn を継続する。Tasq は current run を terminal にし、operator が blocked issue を review できる状態にしたいので、`cancel` の方が desired behavior に合う。

### Human decision まで run を生かし続ける

True interactive approval を実現できるが、pending approval storage、operator UI、timeout、process supervision、crash recovery が必要になる。これは current unattended runner より大きい workflow である。

### 即時 auto-approve する

Operator が matching request を承認済みなら有用だが、unconditional auto-approval は広すぎる。Request type、command、file path、diff、reason、issue、場合によっては workspace による scoped matching が必要である。

## Consequences

Approval request は hidden protocol failure ではなく、visible blocker work になる。Blocked issue comment には operator が yes/no を判断するための context が残る。

Run は引き続きすばやく terminal になる。Tasq は human input を待つために app-server subprocess を生かし続けない。

Next-run auto-approval path は future work として残る。Comment を source of truth として parse するのではなく、structured approval decision が必要である。

## Notes

この ADR は response behavior について ADR-0004 を supersede する。ADR-0004 の core safety principle は維持する。つまり、declined または unapproved の required action を successful run として報告してはならない。
