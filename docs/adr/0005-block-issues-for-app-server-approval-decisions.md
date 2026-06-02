# ADR-0005: Block Issues for App-Server Approval Decisions

## Context

ADR-0004 recorded a conservative posture for Codex app-server approval requests: command-execution
and file-change approvals must not be treated as success when Tasq cannot complete the requested
action.

That decision still leaves an operator workflow question. Operators need to see the requested
approval details, decide yes or no, and then retry the issue with an explicit approval decision.
Treating the request as a generic unsupported protocol failure hides too much context and makes the
next operator action unclear.

The targeted Codex app-server schema defines approval responses as `{"decision":"accept"}`,
`{"decision":"decline"}`, `{"decision":"cancel"}`, and related accepted variants. There is no
`approved: false` response field in the generated schema.

## Decision

Tasq treats `item/commandExecution/requestApproval` and `item/fileChange/requestApproval` as known
approval requests.

For the current unattended runner, Tasq immediately responds to those app-server requests with
`{"decision":"cancel"}`. This is a protocol-level denial that also asks Codex to interrupt the
current turn. The runner then terminates the run as `failed` with an `approval_required` error.

If the latest issue state is still `ready`, the dispatcher marks the issue `blocked` and creates a
blocker comment containing the approval method and raw request payload. The issue status remains the
SPEC-compatible `blocked`; Tasq does not introduce a new `blocked_by_approval` issue status.

Human approval is handled out of band for now. An operator reviews the blocked issue and comment,
decides whether the request is acceptable, and can move the issue back to `ready` for retry. A future
approval decision store should allow the next run to auto-approve only the matching request.

## Alternatives

### Keep Returning Unsupported JSON-RPC Errors

This is simple, but it makes known approval requests look like unknown protocol requests. It also
produces less useful blocker comments.

### Respond with `decline`

`decline` is a valid denial response, but the schema says the agent will continue the turn. Tasq wants
the current run to become terminal so the operator can review the blocked issue, so `cancel` matches
the desired behavior better.

### Keep the Run Alive Until a Human Decides

This would support true interactive approval, but it requires pending approval storage, operator UI,
timeouts, process supervision, and crash recovery. That is a larger workflow than the current
unattended runner.

### Auto-Approve Immediately

This can be useful after an operator has approved a matching request, but unconditional auto-approval
is too broad. It needs scoped matching by request type, command, file path, diff, reason, issue, and
possibly workspace.

## Consequences

Approval requests become visible blocker work instead of hidden protocol failures. The blocked issue
comment contains enough context for an operator to make a yes/no decision.

Runs still terminate quickly. Tasq does not keep app-server subprocesses alive while waiting for
human input.

The next-run auto-approval path remains future work. It needs structured approval decisions rather
than parsing comments as the source of truth.

## Notes

This ADR supersedes ADR-0004 for the response behavior. ADR-0004's core safety principle remains:
declined or unapproved required actions must not be reported as successful runs.
