# Orchestrator Workflow

The orchestrator owns run state and optional runtime inspection. It records runs in its own SQLite database, prepares per-issue workspaces, and exposes loopback APIs when HTTP serving is enabled.

Use this workflow when changing `cmd/orchestrator`, `internal/orchestrator`, or `db/schema/orchestrator.sql`.

## Scope

- Keep run records, run attempts, runner events, and workspace metadata owned by the orchestrator.
- Do not make the orchestrator mutate issue state directly.
- Use issue-tracker issue APIs only for reading or reconciling issue-facing state.
- Keep runtime inspection local to the orchestrator HTTP API.

See [../../docs/design.md](../../docs/design.md) for the service boundary.

## Local Run

Start issue-tracker and orchestrator together through Compose when checking service boundaries:

```sh
make orchestrator-up
make dev-ports
```

For host-only development, start the issue-tracker first, then run:

```sh
go run ./cmd/orchestrator \
  -db tasq-orchestrator.sqlite \
  -issue-tracker http://localhost:8080 \
  -port 8081
```

## Change Flow

1. Keep runstore, workflow, runner, workspace, and HTTP server changes separated in the implementation.
2. Keep SQLite schema changes in `db/schema/orchestrator.sql`.
3. Keep issue-tracker contract changes intentional and reflected in its OpenAPI document.
4. When adding runner behavior, keep run progress recorded through the orchestrator runstore.

## Verification

Run focused orchestrator tests while developing:

```sh
go test ./internal/orchestrator/...
```

Run broader checks before handing off behavior that changes runs, workspaces, or runtime inspection:

```sh
go test ./...
```

Use `make dev-test` when verifying through the Compose toolchain.

## Operational Notes

- The orchestrator is the source of truth for its own run records and runner events.
- Issue status changes must go through the issue-tracker issue APIs.
- `/api/v1/refresh` returns `503` when no refresher is configured.
- Workspace setup metadata should be retained enough for debugging and recovery.
