# ADR-0006: App-server を workspace-write sandbox で起動する

## Context

Tasq は local dev container 内から Codex app-server を起動する。この topology では、container が agent execution、credential、mounted file、network access に対する最初の isolation boundary になる。

Codex も generated shell command を sandbox 実行できる。しかし現在の dev-container environment では、`pwd` や `git status` のような command が Bubblewrap の namespace error で失敗することがある。

```text
bwrap: No permissions to create a new namespace
```

この failure により、Codex は低リスクな repository inspection command でも command approval を request する。Tasq は approval workflow により issue を blocked にする。Unknown または broad request で blocked にするのは正しいが、基本的な workspace inspection で繰り返し blocked になると local orchestration が使いにくくなる。

## Decision

Repository workflow は Codex app-server を次の command で起動する。

```sh
codex app-server --sandbox workspace-write
```

これにより、Tasq run では app-server が default で workspace-write sandbox posture を使う。

Tasq は ADR-0005 の approval workflow を維持する。Codex が command-execution または file-change approval を request した場合、Tasq は request を cancel し、run を `approval_required` で failed にし、最新 state が still-ready の issue を request details 付きで blocked にする。

Dev container は local development の primary isolation boundary として扱う。Project は Bubblewrap namespace creation を動かすためだけに dev container を default で privileged にしない。

## Alternatives

### App-server command に explicit sandbox を指定しない

Sandbox behavior が Codex default と local config に委ねられる。明示性が低く、host や container の違いによって local Tasq behavior を再現しづらくなる。

### Dev container を privileged にする

Bubblewrap が namespace を作れる可能性は上がるが、container boundary が弱くなる。Container を primary local isolation layer として扱う以上、routine development の default としては tradeoff が悪い。

### Codex sandbox を完全に無効化する

Externally sandboxed environment では approval prompt を減らせる可能性がある。しかし default posture としては広すぎる。必要な場合のみ explicit opt-in profile として導入するべきである。

### Codex rules または exec policy だけに依存する

Codex rules と exec policy は低リスク command には有用である。しかし runtime sandbox posture を明確にする代替にはならない。また namespace creation failure 後に発生する sandbox-escape approval を完全に covered できるとは限らない。

## Consequences

Local Tasq run は documented かつ reproducible な Codex sandbox mode を持つ。

Codex は per-issue workspace 内で basic workspace write access を持つ。一方で、より broad な approval request は引き続き blocked issue work になる。

Project は default で privileged container を避け続ける。Developer が別の sandbox posture を必要とする場合は、explicit opt-in workflow または profile として導入する。

## Notes

この ADR は ADR-0002 の single dev-container topology と ADR-0005 の blocked approval workflow を補完する。
