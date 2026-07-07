---
id: testing
title: Testing
sidebar_position: 3
---

# Testing

Run the narrowest useful verification first, then broaden checks when a change affects shared behavior, contracts, persistence, or user-facing flows.

## Recommended Matrix

| Change area | Start with | Broaden to |
| --- | --- | --- |
| Go package or service logic | `go test ./internal/<package>` | `go test ./...` |
| Issue-tracker or orchestrator API contract | targeted Go tests | API generation, Web typecheck, `make dev-test` |
| Web UI component or route | `npm run typecheck` in `cmd/web/frontend` | `npm run build`, `make dev-build` |
| Docs site | `npm run build` in `docs/site` | check English and Japanese pages in a browser |
| End-to-end service flow | local smoke flow | `make dev-build` |

## Go Tests

```sh
go test ./...
```

Use targeted packages while iterating, then run the full suite before handoff when shared behavior changed.

With Compose:

```sh
make dc-exec CMD="go test ./internal/config"
make dc-exec CMD="go test ./..."
```

## Web UI Checks

```sh
cd cmd/web/frontend
npm run typecheck
npm run build
```

Regenerate API clients with `npm run generate:api` when OpenAPI contracts
change. Commit the generated files under `cmd/web/frontend/src/lib/generated`
with the OpenAPI change, and update MSW handlers or fixtures when standalone
mock development uses the changed endpoint.

Use `npm run dev:mock` when you only need frontend behavior with in-memory mock
data. Use Compose when you need real issue-tracker or orchestrator behavior.

## Docs Site Checks

```sh
cd docs/site
npm run build
```

For repository-level workflow, `make dev-docs-build` wraps the docs-site build.

Docs changes should update English and Japanese pages together when both exist.
Check links, sidebar placement, and code block languages when adding new pages.

## Compose Verification

Use repository-level targets before handoff when a change touches multiple
runtime areas.

```sh
make dev-test
make dev-build
```

`make dev-test` runs Go tests and Web UI typecheck in the dev container.
`make dev-build` runs Go tests and the Web production build.

## Manual Verification

1. Start the dev environment with `make dev-up` or start host services with `tq service start`.
2. Create and update issues through `tq` or the Web UI.
3. Confirm issue summaries reflect status changes.
4. Confirm orchestrator runtime inspection is reachable when the orchestrator is enabled.

Record skipped checks in the pull request when a verification step is not
applicable or cannot be run locally.
