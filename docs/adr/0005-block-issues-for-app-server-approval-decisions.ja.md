# ADR-0005: App-server の承認判断のために課題を blocked にする

## Context

ADR-0004 は Codex app-server の承認要求に対する保守的な姿勢を記録した。command-execution と file-change の承認は、Tasq が要求された action を完了できない場合に成功として扱ってはならない。

しかし、その判断だけでは運用者ワークフローが不十分である。運用者は要求された承認の詳細を確認し、yes/no を判断し、その明示的な承認判断を使って課題を再試行する必要がある。要求を汎用的な unsupported protocol failure として扱うと、文脈が不足し、次に運用者が何をすべきか分かりにくい。

対象 Codex app-server schema では、approval response は `{"decision":"accept"}`、`{"decision":"decline"}`、`{"decision":"cancel"}` と関連する accepted variants として定義されている。生成された schema には `approved: false` response field はない。

## Decision

Tasq は `item/commandExecution/requestApproval` と `item/fileChange/requestApproval` を既知の承認要求として扱う。

現在の unattended runner では、Tasq はこれらの app-server request に即時 `{"decision":"cancel"}` を返す。これは protocol-level denial であり、同時に Codex に current turn の interrupt を求める。runner はその後、run を `approval_required` error 付きの `failed` として終端させる。

最新の課題状態がまだ runnable（`ready` または `in_progress`）の場合、dispatcher は課題を `blocked` に更新し、approval method と raw request payload を含む blocker comment を作成する。Issue status は SPEC-compatible な `blocked` のままとし、Tasq は新しい `blocked_by_approval` issue status を導入しない。

人による承認は現時点では out of band に扱う。運用者は `blocked` の課題と comment を確認し、その request を許可できるか判断し、再試行のために課題を `ready` へ戻せる。将来の承認判断ストアでは、次の run が一致する request だけを auto-approve できるようにするべきである。

## Alternatives

### Unsupported JSON-RPC error を返し続ける

単純だが、既知の承認要求が unknown protocol request のように見える。また blocker comment も運用者判断に必要な情報として弱い。

### `decline` を返す

`decline` は有効な denial response である。ただし schema 上はエージェントが turn を継続する。Tasq は current run を終端させ、運用者が `blocked` の課題を review できる状態にしたいので、`cancel` の方が望ましい挙動に合う。

### Human decision まで run を生かし続ける

真の対話型承認を実現できるが、pending approval storage、operator UI、timeout、process supervision、crash recovery が必要になる。これは現在の unattended runner より大きいワークフローである。

### 即時 auto-approve する

運用者が一致する request を承認済みなら有用だが、無条件の auto-approval は広すぎる。Request type、command、file path、diff、reason、issue、場合によっては workspace による scoped matching が必要である。

## Consequences

承認要求は hidden protocol failure ではなく、見える blocker work になる。Blocked issue comment には運用者が yes/no を判断するための文脈が残る。

Run は引き続きすばやく終端する。Tasq は human input を待つために app-server subprocess を生かし続けない。

Next-run auto-approval path は将来の作業として残る。Comment を source of truth として parse するのではなく、構造化された承認判断が必要である。

## Notes

この ADR は response behavior について ADR-0004 を supersede する。ADR-0004 の中核となる安全原則は維持する。つまり、declined または unapproved の required action を successful run として報告してはならない。
