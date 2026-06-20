# Stream ready issues for the orchestrator

`tq issue watch` is an experimental, NDJSON-only command intended for the orchestrator skill. It polls the issue tracker, emits one JSON envelope per ready issue, and runs until killed.

For the full orchestration loop (hand-off rules, branch / worktree setup, status updates after dispatch), use [[tq-orchestrator]]. This file only covers the `tq` invocation.

## Invocation

```bash
tq issue watch --interval 30 --seen-ttl 900
```

Flag reference (`--interval`, `--seen-ttl`, `--verbose`) and the full NDJSON envelope shapes live in [resources/issue.md](../resources/issue.md). Keep that file as the single source of truth — do not re-state the table here.

Key constraints to remember when wiring the dispatcher:

- `--seen-ttl` must be strictly greater than `--interval`, otherwise the command refuses to start.
- `event` and `error` envelopes are always emitted; `info` only with `--verbose`.
- Transient `error` envelopes do not stop the loop — the watcher keeps polling.

## How to run it

Run under the Monitor tool so each stdout line becomes one notification. Do not invoke it with a blocking Bash call — the command runs until SIGINT / kill / Monitor timeout.

## After receiving an event

The expected agent response (defined by [[tq-orchestrator]]):

```bash
ID=42
tq issue update "$ID" --status in_progress
tq comment add "$ID" --author claude-code --type handoff \
  --body "Dispatched to subagent <name>"
```

## See also

- [resources/issue.md](../resources/issue.md) — `watch` flag reference
- [[tq-orchestrator]] — full orchestration policy
