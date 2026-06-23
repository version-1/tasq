# Web Frontend Design

This document defines local structure rules for the React/Vite frontend under `cmd/web/frontend`.

## Routing

Routes are declared manually in `src/App.tsx` with `react-router-dom`.

Use path parameters for scoped resource pages and detail pages. Issue lists use `/issues` for all projects and `/projects/:projectKey/issues` for a single project. Project detail pages use `/projects/:projectKey` and redirect to `/projects/:projectKey/issues`; their fixed tabs are `/projects/:projectKey/issues` and `/projects/:projectKey/settings`. Issue details use `/issues/:id` and the source files live under `src/app/issues/[id]/`.

## Component Placement

Use one component per file. Components that own styles must be placed in a directory with this shape:

```text
component-name/
  index.tsx
  index.module.css
```

Keep component-specific styles in that component's `index.module.css`. Shared non-component helpers may live next to the component directories, such as `types.ts`, but avoid putting multiple React components in the same file.

Route-local components live under each route's `_components` directory. They
compose page state, hooks, and feature components, but they should not become
the shared home for reusable domain UI.

For the issue list route, keep only the route composition under
`src/app/issues/_components/issues-view/`:

```text
issues-view/
  index.tsx
  index.module.css
```

The top-level `issues-view/index.tsx` should hold only the `IssuesView`
component and import reusable issue-domain UI from `src/features/issues/`.

Issue-domain UI components shared across routes live under
`src/features/issues/components/`:

```text
features/issues/components/
  board/
    index.tsx
    index.module.css
  card/
    index.tsx
    index.module.css
  markdown/
    index.tsx
    index.module.css
  pane/
    index.tsx
    index.module.css
```

Keep `src/components/ui/` for domain-independent design-system primitives,
including shared modal components, and `src/components/layout/` for application
shell and global layout components.

Project detail tab pages live under `src/app/projects/[projectKey]/`. Keep tab-specific UI in each route's `_components` directory. The settings tab is read-only and renders the synchronized `ProjectWorkflow` from `GET /api/v1/projects/{id}/workflow`; do not add local `WORKFLOW.md` file reads in frontend code.
