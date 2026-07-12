# Web Frontend Design

This document defines the frontend ownership and placement rules for the
React/Vite application under `cmd/web/frontend`.

The goal is to keep routing, feature UI, app shell UI, and generic UI primitives
separate. New code should be placed by ownership first, then by reuse scope.

For visual tokens and component styling conventions, see
[ui-design-system.md](ui-design-system.md). For a browser-friendly visual
reference, open [ui-design-system.html](ui-design-system.html).

## Routing

Routes are declared manually in `src/App.tsx` with `react-router-dom`.

The `src/app/` tree is for route entry files. A route file may read route
parameters, use layout hooks, perform route-level data loading, and render a
feature component. Do not create `_components` directories under `src/app/`.

Use path parameters for scoped resources and detail pages:

- `/dashboard` lists issues across all projects.
- `/dashboard/table` renders a paginated issue table across all projects.
- `/dashboard/stats` renders dashboard metrics and running agent summaries.
- `/projects/:projectKey/issues` lists issues for one project.
- `/projects/:projectKey/table` renders a paginated issue table for one project.
- `/projects/:projectKey` redirects to `/projects/:projectKey/issues`.
- `/projects/:projectKey/settings` renders project workflow settings.
- `/issues/:id` renders an issue detail page.

## Top-Level Directories

Use these ownership boundaries when adding or moving frontend code:

```text
src/
  app/          route entry files only
  components/   app-shell components and domain-independent UI primitives
  features/     page-level and domain-aware feature components
  lib/          API clients, generated types, i18n, stores, and shared runtime helpers
  mocks/        MSW handlers and fixture data for local/mock execution
  stories/      Storybook-only fixture builders and helpers
```

The dependency direction should stay simple:

- `src/app` may import from `components`, `features`, and `lib`.
- `src/features` may import from `components` and `lib`.
- `src/components/ui` must stay domain-independent.
- `src/components/layout` may coordinate the app shell, navigation, modal slots,
  and layout-level context.

## Architecture Layers

Use these layers when deciding where frontend code belongs:

```text
src/app
src/components
src/features/<name>/components
  Presentation layer

src/features/<name>/hooks
  Application-oriented layer
  Owns UI state, use-case orchestration, and view-model creation.

src/features/<name>/context
  Application-oriented layer
  Owns feature-scoped state and providers.

src/features/<name>/api
  Repository / adapter layer
  Provides the feature-facing data access interface.
  Wraps src/lib/api.

src/lib/api
  Infrastructure layer
  Centralizes Orval-generated clients, HTTP, error handling, and the transport boundary.

src/lib/generated
  Generated infrastructure detail
  Feature code must not import this directly.
```

React hooks are React-dependent, so they are not a pure domain layer by
themselves. If pure domain rules grow inside a hook, extract those rules into
framework-independent functions or a feature-owned domain/model module before
they become hard to test.

## Component Shape

Use one React component per file. Components that own styles must live in a
directory with this shape:

```text
component-name/
  index.tsx
  index.module.css
  index.stories.tsx
```

Keep component-specific styles in `index.module.css`. Keep component-specific
tests, stories, helper types, and small helper functions next to the component
that owns them.

Shared non-component helpers may live next to the component directories, such
as `types.ts`, `format.ts`, or `rows.ts`, when they are owned by that feature or
component group.

## Route Entries

Route entries live under `src/app/**/page.tsx`.

They should stay thin. A route entry can:

- read route parameters;
- use layout-level data hooks;
- perform route-specific data loading;
- select the feature component to render;
- provide `Suspense` or route-level fallback behavior.

Route entries should not contain reusable UI sections, cards, tables, panels,
or domain presentation components. Put those in the matching feature directory.

## Feature Components

Feature components live under `src/features/<feature>/components/`.

Use feature directories for page-level views, domain-aware UI, and components
whose props or behavior are coupled to a product domain.

Current feature component groups include:

```text
features/
  dashboard/
    api/
    hooks/
    context/
    components/
      dashboard-view/
  issues/
    api/
    hooks/
    context/
    components/
      board/
      card/
      conversation-page/
      issue-detail-page/
      issues-view/
      pane/
  projects/
    api/
    hooks/
    context/
    components/
      workflow-settings-view/
  settings/
    api/
    hooks/
    context/
    components/
      settings-view/
```

Each feature domain owns these subdirectories when the responsibility exists:

- `components/` for page-level and domain-aware React components;
- `hooks/` for feature-specific state, effects, and derived view-model hooks;
- `context/` for feature-scoped providers and context values;
- `api/` for feature-facing request wrappers around `src/lib/api`.

Do not create empty subdirectories only to match the template. Add `api`,
`hooks`, or `context` when the feature has code with that responsibility.
Feature page components may compose smaller feature components. Keep their
view-model helpers, tests, and stories in the same feature area.

API clients must come from `src/lib/api`, which owns the Orval-generated clients
under `src/lib/generated`. Do not hand-write endpoint clients inside a feature,
and do not import generated clients directly from feature code. Feature `api/`
modules may wrap `src/lib/api` functions to provide feature-specific names,
request composition, error mapping, or view-model conversion, but the HTTP
contract itself must stay generated and centralized in `src/lib/api`.

### Feature Placement Policy

Use the `code-react` guidance as the baseline for feature directories. Atomic
Design is a supporting lens, not a directory taxonomy. `atoms` and `molecules`
belong to the design-system layer only when they express visual style, variants,
layout, interaction state, and accessibility without domain language.

Place a component under `src/features/<feature>/components/` when any of these
are true:

- it receives issue, project, workflow, dashboard, or settings domain data;
- its props use domain terms such as status, priority, assignee, workflow, run,
  or frontmatter;
- it encodes domain-specific display rules or allowed user actions;
- it is a page-level view or section that belongs to one feature;
- it composes generic UI primitives into a product-specific experience.

Keep feature state and rendering responsibilities separate:

- route-level URL and loading concerns stay in `src/app/**/page.tsx`;
- feature view state and derived view models stay in the feature's `hooks/` or
  next to the owning component when they are component-local;
- feature-scoped providers stay in the feature's `context/`;
- feature-facing API adapters stay in the feature's `api/` and delegate to
  `src/lib/api`;
- reusable display-only UI with no domain vocabulary stays in
  `src/components/ui`;
- API clients, generated types, i18n, stores, and runtime helpers stay in
  `src/lib`.

Do not promote a feature component into `src/components/ui` just because it is
used twice. Promote it only when concrete reuse exists outside the feature and
the component can be expressed without product-domain props or behavior. If
props become broad, boolean-heavy, or difficult for the caller to understand,
keep the component feature-owned and extract only the domain-free primitive.

## Shared UI Components

Domain-independent UI primitives live under `src/components/ui/`.

Use this directory for reusable UI that does not know about issue, project, or
dashboard domain concepts. Examples include:

```text
components/ui/
  badge/
  button/
  context-menu/
  icon-proxy/
  markdown/
  modal/
  pannel-message/
  switch/
  toast/
```

Shared Markdown rendering belongs in `components/ui/markdown/` because it is
used by multiple domains. Shared modal rendering belongs in
`components/ui/modal/` because it provides a generic portal slot.

Do not place domain-specific labels, issue status transitions, project workflow
tables, or page-specific layout in `components/ui`.

## Layout Components

Application shell and global navigation components live under
`src/components/layout/`.

This directory owns:

- the sidebar and header;
- layout shell composition;
- layout-level context and hooks;
- modal entry points for shell-level dialogs;
- route layout wrappers such as default, issue detail, and project layouts.

Feature views may be rendered inside the layout shell, but layout components
should not own feature-specific presentation.

### Theme ownership

`src/components/layout/use-theme.ts` owns the runtime theme state for the
layout sidebar switch. It writes the selected `light` or `dark` value to
`localStorage` under `tasq.theme` and reflects the resolved value through
`html[data-theme]`.

`index.html` applies the same resolution before React mounts: a valid
`tasq.theme` value wins, otherwise `prefers-color-scheme` supplies the initial
theme. The hook listens for operating system preference changes only while no
explicit value is stored. Keep this resolution logic aligned between the
pre-mount script and the hook.

## Project Workflow Settings

Project workflow settings are rendered by feature components under
`src/features/projects/components/workflow-settings-view/`.

The settings route is read-only and renders the synchronized `ProjectWorkflow`
from `GET /api/v1/projects/{id}/workflow`. Frontend code must not read local
`WORKFLOW.md` files.

Workflow frontmatter should be shown as a tree-friendly table. Workflow body
content should use the shared Markdown renderer from `src/components/ui/markdown`.

## Storybook

Every React component under `src/components/**/index.tsx` and
`src/features/**/index.tsx` must have a matching `index.stories.tsx`.

Use Storybook titles that match ownership:

- `UI/...` for `src/components/ui`.
- `Layout/...` for `src/components/layout`.
- `Features/<Feature>/...` for `src/features/<feature>`.

Use `src/stories/` for Storybook-only fixtures and builders. Do not import
Storybook helpers from production code.

Storybook static build output is generated under `storybook-static/` and must
not be committed.
