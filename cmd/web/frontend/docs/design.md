# Web Frontend Design

This document defines local structure rules for the React/Vite frontend under `cmd/web/frontend`.

## Routing

Routes are declared manually in `src/App.tsx` with `react-router-dom`.

Use path parameters for resource detail pages. For example, issue details use `/issues/:id` and the source files live under `src/app/issues/[id]/`.

## Component Placement

Use one component per file. Components that own styles must be placed in a directory with this shape:

```text
component-name/
  index.tsx
  index.module.css
```

Keep component-specific styles in that component's `index.module.css`. Shared non-component helpers may live next to the component directories, such as `types.ts`, but avoid putting multiple React components in the same file.

For issue list UI components, follow this pattern under `src/app/issues/_components/issues-view/`:

```text
issues-view/
  index.tsx
  index.module.css
  issue-board/
    index.tsx
    index.module.css
  issue-card/
    index.tsx
    index.module.css
  panel-message/
    index.tsx
    index.module.css
```

The top-level `issues-view/index.tsx` should compose child components and hold only the `IssuesView` component.

Issue UI components shared outside the issue list view live under `src/components/issue/`:

```text
issue/
  card/
    index.tsx
    index.module.css
  pane/
    index.tsx
    index.module.css
```
