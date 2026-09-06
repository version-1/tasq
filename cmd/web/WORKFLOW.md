# Web UI Workflow

The Web UI is a Go HTTP server that embeds a Vite + React + TypeScript single-page app. The server serves frontend assets and reverse-proxies API traffic to local backend services.

Use this workflow when changing files under `cmd/web`.

## Scope

- Keep the Go server responsible for static asset serving, SPA fallback, and API reverse proxying.
- Keep browser UI code under `cmd/web/frontend/src`.
- Keep generated issue-tracker and orchestrator API files under `cmd/web/frontend/src/lib/generated` regenerable.
- Keep user-facing strings in `cmd/web/frontend/src/lib/i18n.ts`.
- Keep feature-specific styles next to the owning component or route.

See [Frontend Design](frontend/docs/design.md) for frontend ownership and
component placement rules, and [UI Design System](frontend/docs/ui-design-system.md)
for styling conventions.

## Local Run

Prefer the repository-level Compose flow when checking the Web UI with the issue-tracker:

```sh
make run-web
make dev-ports
```

For host-only development:

```sh
cd cmd/web/frontend
npm install
npm run build
cd ../../..
go run ./cmd/web
```

Frontend standalone development without issue-tracker or orchestrator:

```sh
cd cmd/web/frontend
npm run dev:mock
```

`dev:mock` sets `VITE_MSW=true` and starts the browser MSW worker from
`cmd/web/frontend/public/mockServiceWorker.js`. Mock data is in-memory and resets
when the page reloads.

## API Generation

See [../../docs/development.md](../../docs/development.md#api-generation) for the OpenAPI regeneration command and generated-file ownership.

Update MSW handlers and fixtures under `cmd/web/frontend/src/mocks` when standalone frontend development uses the changed endpoint.

## Component Flow

See [frontend/docs/design.md](frontend/docs/design.md) for frontend routing, feature/component placement, and directory ownership rules.

- Keep translated display strings in `cmd/web/frontend/src/lib/i18n.ts`.
- Keep API-facing types imported through `cmd/web/frontend/src/lib/types.ts` unless a generated operation is required directly.

## Verification

Run type checking before handing off Web UI changes:

```sh
cd cmd/web/frontend
npm run typecheck
```

Run a production build for changes that affect routing, generated API usage, global styling, or embedded assets:

```sh
cd cmd/web/frontend
npm run build
cd ../../..
go build -o .tmp/tasq-web ./cmd/web
```

See [../../docs/development.md](../../docs/development.md#verification) for the Compose-backed `make dev-build` check.
