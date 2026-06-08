# Web UI Design

The Web UI is a Vite + React + TypeScript single-page app served by `cmd/web`. The Go server embeds the production build, serves SPA fallback routes, and proxies API calls.

## Routes

User-facing views are split by React Router routes:

- `/issues`
- `/projects/:projectKey/issues`
- `/issues/:id`
- `/dashboard`
- `/settings`

The root `/` route redirects to `/issues`.

## Rendering Model

The Web UI runs entirely in the browser. Keep browser state, API calls, routing, and translation concerns in React code under `cmd/web/frontend/src`.

The Go server owns:

- serving `cmd/web/frontend/dist` through `go:embed`
- returning `index.html` for non-API SPA routes
- proxying `/tracker/*` to the issue-tracker
- proxying `/orchestrator/*` to the orchestrator

Do not add server-rendered UI behavior to `cmd/web`. If runtime configuration is needed, prefer explicit Go flags or proxy paths over build-time browser environment variables.

## API Client

The issue-tracker API client is generated from `docs/openapi/issue-tracker.yml` with Orval.
The orchestrator API client is generated from `docs/openapi/orchestrator.yml` with Orval.

Run the generator from `cmd/web/frontend` whenever either OpenAPI definition changes:

```sh
npm run generate:api
```

Generated files live under `cmd/web/frontend/src/lib/generated` and must not be edited manually. Route-facing code should import API types through `cmd/web/frontend/src/lib/types.ts` unless it needs a generated operation directly.

## Component Structure

Shared shell concerns, such as navigation, summary loading, refresh handling, and page selection, live outside route-specific component directories.

Common components that are intentionally shared by multiple routes belong under `cmd/web/frontend/src/components`.

Page-specific components live under the owning route's `_components` directory:

- Issue page components belong in `cmd/web/frontend/src/app/issues/_components`.
- Workspace page components belong in `cmd/web/frontend/src/app/workspace/_components`.
- Settings page components belong in `cmd/web/frontend/src/app/settings/_components`.

Keep components close to the route that owns their behavior. Move code to shared components only when it is genuinely shared across routes.

Each component, whether shared or route-specific, must use a directory named after the component. Put the implementation, CSS Module, and component tests in that directory:

```text
<component-name>/
├── index.tsx
├── index.module.css
└── index.test.tsx
```

## Internationalization

The Web UI uses `react-i18next` for display strings.

Supported languages:

- Japanese (`ja`)
- English (`en`)

Keep user-facing UI text in `cmd/web/frontend/src/lib/i18n.ts`. Components should render translated text with `useTranslation()` instead of hard-coded display strings. User-provided issue content, API identifiers, and route path segments can remain untranslated.

## Styling

The frontend uses CSS Modules for component and page styling.

Keep `cmd/web/frontend/src/app/globals.css` limited to global tokens and base element resets. Feature-specific styles must live next to the component that owns the markup.

See [Web UI Color Palette](web-color-pallete.md) for global color tokens.

Examples:

- `cmd/web/frontend/src/components/layout/index.module.css`
- `cmd/web/frontend/src/app/issues/_components/issues-view/index.module.css`

Do not add feature-specific class selectors to `cmd/web/frontend/src/app/globals.css`.
