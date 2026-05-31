# Web UI Workflow

The Web UI is a statically exportable Next.js App Router client. It talks to the issue-tracker API only and does not call the orchestrator directly.

Use this workflow when changing files under `web`.

## Scope

- Keep runtime behavior in Client Components.
- Keep API calls routed through the issue-tracker API client.
- Keep generated API files under `web/lib/generated` regenerable.
- Keep user-facing strings in `web/lib/i18n.ts`.
- Keep feature-specific styles next to the owning component or route.

See [docs/design.md](docs/design.md) for Web UI structure and styling conventions.

## Local Run

Prefer the repository-level Compose flow when checking the Web UI with the issue-tracker:

```sh
make web-up
make dev-ports
```

For host-only development:

```sh
cd web
npm install
NEXT_PUBLIC_ISSUE_TRACKER_URL=http://localhost:8080 npm run dev
```

## API Generation

The issue-tracker API client is generated from `../docs/openapi/issue-tracker.yml`.

Run this command from `web` whenever the OpenAPI document changes:

```sh
npm run generate:api
```

Do not manually edit `web/lib/generated`.

## Component Flow

1. Put route-owned UI under the route's `_components` directory.
2. Move components to `web/components` only when they are genuinely shared.
3. Use the component directory shape defined in `web/docs/design.md`.
4. Keep translated display strings in `web/lib/i18n.ts`.
5. Keep API-facing types imported through `web/lib/types.ts` unless a generated operation is required directly.

## Verification

Run type checking before handing off Web UI changes:

```sh
cd web
npm run typecheck
```

Run a production build for changes that affect routing, static export behavior, generated API usage, or global styling:

```sh
cd web
npm run build
```

Use `make dev-build-app` when verifying Go and Web UI changes together through Compose.

## Static Export Notes

- Do not add server actions, route handlers, server redirects, cookies, headers, or Metadata API behavior for Web UI runtime needs.
- Configure the API origin with `NEXT_PUBLIC_ISSUE_TRACKER_URL` when the static UI is served from a different origin.
- Do not rely on Next.js rewrites, redirects, or headers at runtime.
