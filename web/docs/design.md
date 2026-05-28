# Web UI Design

The Web UI is a Next.js App Router client for issue operations. It talks to the issue-tracker API only and does not call the orchestrator directly.

## Routes

User-facing views are split by App Router pages:

- `/issues`
- `/agents`
- `/settings`

The root `/` route redirects to `/issues`.

## Component Structure

Shared shell concerns, such as navigation, summary loading, refresh handling, and page selection, live outside route-specific component directories.

Common components that are intentionally shared by multiple routes belong under `web/components`.

Page-specific components live under the owning route's `_components` directory:

- Issue page components belong in `web/app/issues/_components`.
- Agent page components belong in `web/app/agents/_components`.
- Settings page components belong in `web/app/settings/_components`.

Keep components close to the route that owns their behavior. Move code to `web/components` only when it is genuinely shared across routes.

Each component, whether shared or route-specific, must use a directory named after the component. Put the implementation, CSS Module, and component tests in that directory:

```text
<component-name>/
├── index.tsx
├── index.module.css
└── index.test.css
```

## Internationalization

The Web UI uses `react-i18next` for display strings.

Supported languages:

- Japanese (`ja`)
- English (`en`)

Keep user-facing UI text in `web/lib/i18n.ts`. Components should render translated text with `useTranslation()` instead of hard-coded display strings. User-provided issue content, API identifiers, and route path segments can remain untranslated.

## Styling

The Next.js app uses CSS Modules for component and page styling.

Keep `web/app/globals.css` limited to global tokens and base element resets. Feature-specific styles must live next to the component that owns the markup.

Examples:

- `web/components/layout/index.module.css`
- `web/app/issues/_components/issues-view/index.module.css`
- `web/app/agents/_components/agents-view/index.module.css`

Do not add feature-specific class selectors to `web/app/globals.css`.
