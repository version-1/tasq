# ADR-0004: Fail Runs on Unsupported App-Server Approvals

## Context

Tasq runs Codex through the app-server JSON-RPC protocol. During a turn, Codex may send server-to-client approval requests such as `item/commandExecution/requestApproval` or `item/fileChange/requestApproval`.

Tasq does not currently implement an interactive approval flow. The orchestrator runs unattended and cannot safely decide to approve commands or file changes on behalf of an operator. Returning success after declining a required approval is also unsafe, because the issue may look complete even though the requested command or file change never happened.

The Codex app-server can still emit `turn/completed` after Tasq returns an unsupported JSON-RPC error for an approval request. Treating only `turn/completed` as success can therefore mask a declined required action.

## Decision

Tasq treats command-execution and file-change approval requests as unsupported approval failures.

When the runner receives `item/commandExecution/requestApproval` or `item/fileChange/requestApproval`, it returns an unsupported JSON-RPC error to the app-server and remembers that the active turn required an unsupported approval. If the same turn later emits `turn/completed`, the runner still returns a failed run result instead of success.

The dispatcher then uses its existing failed-run path. If the latest issue state is still `ready`, the dispatcher marks the issue `blocked` and creates a blocker comment containing the runner failure reason.

## Alternatives

### Treat `turn/completed` as Success

This keeps the protocol handling simple, but it can mark a run succeeded after a required command or file change was declined. That hides incomplete work and can leave the issue in a misleading state.

### Stall and Wait for Operator Approval

This would preserve the possibility of human approval, but Tasq does not yet have an approval UI, operator routing, or durable approval state. Waiting indefinitely would make unattended orchestration unreliable.

### Auto-Approve App-Server Approval Requests

Auto-approval could improve task completion in a trusted environment, but it expands the trust boundary and needs a separate security decision covering command scope, file scope, credentials, sandboxing, and auditability. Tasq does not make that decision in this slice.

## Consequences

Unsupported required approvals fail clearly. A later `turn/completed` notification no longer overrides the fact that Tasq declined a required action.

Ready issues move to `blocked` with a blocker comment through the existing dispatcher behavior. This makes the failure visible to users and prevents the poller from immediately re-queueing the same issue as if it were still ready.

Some tasks that could have completed with human approval now fail until Tasq implements an approval policy. This is intentional: it favors explicit blocked state over false success.

The app-server approval posture remains conservative. Future work can add interactive approvals or scoped auto-approval as a separate ADR.

## Notes

This ADR documents the approval posture required by `docs/symphony/SPEC.md`: approval and user-input behavior is implementation-defined, but approval requests must not leave runs indefinitely stalled.

Related implementation notes are in `docs/symphony/CODEX_APP_SERVER.md`.
