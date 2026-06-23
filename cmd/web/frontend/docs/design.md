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

Keep `src/app/` for route entry files only. Route files may load data from
layout hooks or route parameters, then render feature components, but do not
create `_components` directories under `src/app/`.

Page-level and domain-aware components live under the matching feature:

```text
features/
  dashboard/components/dashboard-view/
  issues/components/issues-view/
  issues/components/issue-detail-page/
  issues/components/conversation-page/
  projects/components/workflow-settings-view/
  settings/components/settings-view/
```

Feature page components should compose smaller components and keep their helper
types, tests, and component-specific styles next to the component directory.

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
  pane/
    index.tsx
    index.module.css
```

Keep `src/components/ui/` for domain-independent design-system primitives,
including shared modal and Markdown components, and `src/components/layout/` for
application shell and global layout components.

Project detail tab routes live under `src/app/projects/[projectKey]/`. Keep
tab-specific UI in the matching feature directory, such as
`src/features/projects/components/workflow-settings-view/`. The settings tab is
read-only and renders the synchronized `ProjectWorkflow` from
`GET /api/v1/projects/{id}/workflow`; do not add local `WORKFLOW.md` file reads
in frontend code.
