# Issue Tracker Workflow

The issue-tracker is the user-facing API and the source of truth for issues, projects, workspaces, and issue summaries.

Use this workflow when changing `cmd/issue-tracker`, `internal/issue`, `db/schema/issue_tracker.sql`, or the issue-tracker OpenAPI contract.

## Scope

- Keep issue state owned by the issue-tracker.
- Keep orchestrator run state out of the issue-tracker persistence model.
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
TQ_HOME=./.tasq go run ./cmd/issue-tracker -addr :8080
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

- Issue status changes must go through issue APIs.
- Summary responses should reflect issue data only, not orchestrator run snapshots.
- Removed endpoint families such as work item claims and orchestrator event receivers must not be reintroduced without an explicit contract change.
