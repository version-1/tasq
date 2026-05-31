# Orchestrator Workflow

The orchestrator owns agent assignment and run state. It claims executable work items from the issue-tracker, records runs in its own SQLite database, and delivers run events through a durable outbox.

Use this workflow when changing `cmd/orchestrator`, `internal/orchestrator`, or `db/schema/orchestrator.sql`.

## Scope

- Keep run records, run attempts, claim tokens attached to runs, and outbox delivery state owned by the orchestrator.
- Do not make the orchestrator mutate issue state directly.
- Communicate issue-facing changes by pushing run events to the issue-tracker.
- Preserve retry-safe outbox behavior before adding real agent or workspace execution.

See [../../docs/design.md](../../docs/design.md) for the service boundary.

## Local Run

Start issue-tracker and orchestrator together through Compose when checking the queue and event boundary:

```sh
make orchestrator-up
make dev-ports
```

For host-only development, start the issue-tracker first, then run:

```sh
go run ./cmd/orchestrator \
  -db tasq-orchestrator.sqlite \
  -issue-tracker http://localhost:8080
```

## Change Flow

1. Keep polling, claim, run creation, and outbox delivery changes separated in the implementation.
2. Keep SQLite schema changes in `db/schema/orchestrator.sql`.
3. Treat the current simulated lifecycle as temporary verification behavior.
4. When adding real runner behavior later, keep the issue-tracker contract stable unless an API change is intentional.

## Verification

Run focused orchestrator tests while developing:

```sh
go test ./internal/orchestrator/...
```

Run broader checks before handing off behavior that changes claims, runs, or outbox delivery:

```sh
go test ./...
```

Use `make dev-test` when verifying through the Compose toolchain.

## Operational Notes

- The orchestrator must write run events to the outbox before delivery.
- Delivery retries must not double-apply issue-tracker state transitions.
- Claim tokens tie run events to one work item claim generation.
- Lease behavior must remain safe for parallel orchestrator instances.
