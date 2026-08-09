# Web UI Design

The Web UI is a Vite + React + TypeScript single-page app served by `cmd/web`. The Go server embeds the production build, serves SPA fallback routes, and proxies API calls.

## Routes

User-facing views are split by React Router routes:

- `/dashboard`
- `/dashboard/table`
- `/dashboard/stats`
- `/projects/:projectKey`
- `/projects/:projectKey/issues`
- `/projects/:projectKey/table`
- `/projects/:projectKey/settings`
- `/issues/:id`
- `/settings`

The root `/` route redirects to `/dashboard`.
The project detail root `/projects/:projectKey` redirects to `/projects/:projectKey/issues`.
Dashboard pages have exactly three fixed tabs:

- `board`: shows the issue board across all projects.
- `table`: shows a paginated issue table across all projects.
- `stats`: shows dashboard metrics and running agent summaries.

Project detail pages have exactly three fixed tabs:

- `issues`: shows the existing issue board scoped to the selected project.
- `table`: shows a paginated issue table scoped to the selected project.
- `settings`: shows the synchronized `WORKFLOW.md` for the selected project as read-only data.

The project settings tab reads from the existing issue-tracker endpoint `GET /api/v1/projects/{id}/workflow`. It renders `ProjectWorkflow.body` as sanitized Markdown, renders `ProjectWorkflow.frontmatter` as a recursive key-value tree, and displays `ProjectWorkflow.updatedAt` as the synchronization timestamp. It must not read local `WORKFLOW.md` files directly and must not expose editing controls.

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
- Project detail tab components belong in `cmd/web/frontend/src/app/projects/[projectKey]/**/_components`.
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

## Artifact links

Issue cards show an `Open pull request` context-menu item only when the issue has a `pull_request` artifact. The issue details sidebar similarly shows an `Artifacts` section with a `Pull request` link only when that artifact exists. Neither surface renders an empty section, placeholder, or editing control when artifacts are absent.

Both links open the external URL in a new tab without an opener. Artifact creation, updating, and deletion remain CLI and API operations; the Web UI is display-only for artifacts.

## Theme

The Web UI supports light and dark themes. `cmd/web/frontend/index.html` resolves
the initial theme before React mounts and sets `data-theme` on `html`, preventing
the first paint from using the wrong token set.

The resolution order is:

1. A valid persisted value in `localStorage` under `tasq.theme` (`light` or
   `dark`).
2. The operating system preference from `prefers-color-scheme` when no valid
   persisted value is available.

`src/components/layout/use-theme.ts` owns the runtime behavior used by the
layout sidebar switch. Changing the switch updates `html[data-theme]` and
persists the explicit choice under `tasq.theme`. While no explicit choice is
stored, the hook follows operating system preference changes; once a choice is
stored, that choice takes precedence.

## Styling

The frontend uses CSS Modules for component and page styling.

Keep `cmd/web/frontend/src/app/globals.css` limited to global tokens and base element resets. Feature-specific styles must live next to the component that owns the markup.

Light tokens are defined on `:root`; `[data-theme="dark"]` overrides the color
and shadow token values. Components must continue to consume semantic CSS
variables rather than branching on the active theme or introducing theme-local
color literals.

See [Web UI Color Palette](web-color-pallete.md) for global color tokens.

Examples:

- `cmd/web/frontend/src/components/layout/index.module.css`
- `cmd/web/frontend/src/app/issues/_components/issues-view/index.module.css`

Do not add feature-specific class selectors to `cmd/web/frontend/src/app/globals.css`.
