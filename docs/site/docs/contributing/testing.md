---
id: testing
title: Testing
sidebar_position: 3
---

# Testing

Run the narrowest useful verification first, then broaden checks when a change affects shared behavior, contracts, persistence, or user-facing flows.

## Go Tests

```sh
go test ./...
```

Use targeted packages while iterating, then run the full suite before handoff when shared behavior changed.

## Web UI Checks

```sh
cd cmd/web/frontend
npm run typecheck
npm run build
```

Regenerate API clients with `npm run generate:api` when OpenAPI contracts change.

## Docs Site Checks

```sh
cd docs/site
npm run build
```

For repository-level workflow, `make dev-docs-build` wraps the docs-site build.

## Manual Verification

1. Start the dev environment with `make dev-up` or start host services with `tq service start`.
2. Create and update issues through `tq` or the Web UI.
3. Confirm issue summaries reflect status changes.
4. Confirm orchestrator runtime inspection is reachable when the orchestrator is enabled.
