# Issue Tracker Workflow

The issue-tracker is the user-facing API and the source of truth for issues, work items, received orchestrator event ids, and latest run snapshots.

Use this workflow when changing `cmd/issue-tracker`, `internal/issue`, `db/schema/issue_tracker.sql`, or the issue-tracker OpenAPI contract.

## Scope

- Keep issue state and work item claim state owned by the issue-tracker.
- Keep orchestrator run state as received facts, not as directly managed issue-tracker state.
- Keep UI and TUI clients talking to the issue-tracker API only.
- Keep contract changes reflected in `docs/openapi/issue-tracker.yml`.

See [../../docs/design.md](../../docs/design.md) for the architecture boundary.

## Local Run

Prefer the repository-level Compose flow when testing service interactions:

```sh
make issue-tracker-up
make dev-ports
```

For host-only development:

```sh
go run ./cmd/issue-tracker -addr :8080 -db tasq-issues.sqlite
```

## Change Flow

1. Update the domain, store, API handler, and OpenAPI contract together when a behavior changes the public API.
2. Keep SQLite schema changes in `db/schema/issue_tracker.sql`.
3. Regenerate the Web UI API client when `docs/openapi/issue-tracker.yml` changes:

   ```sh
   cd web
   npm run generate:api
   ```

4. Do not manually edit generated files under `web/lib/generated`.

## Verification

Run focused Go tests while developing:

```sh
go test ./internal/issue/...
```

Run the broader repository checks before handing off a contract or persistence change:

```sh
go test ./...
cd web
npm run typecheck
```

Use `make dev-test` when verifying through the Compose toolchain.

## Operational Notes

- Claim tokens are generation markers for work item claims.
- Duplicate orchestrator event ids must be accepted idempotently.
- Late events from expired claims may be recorded, but must not update current issue state.
