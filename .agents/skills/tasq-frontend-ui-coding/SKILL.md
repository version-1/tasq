---
name: tasq-frontend-ui-coding
description: Code Tasq web frontend UI changes using the repository UI design system. Use this skill whenever editing React, CSS Modules, Storybook stories, layout, UI primitives, issue screens, dashboard, settings, or any visual/frontend behavior under cmd/web/frontend, even when the user only asks for a small style refinement.
---

# Tasq Frontend UI Coding

## Purpose

Implement Tasq web frontend UI changes in the same visual language, component
ownership model, and CSS token system as the existing app.

The source of truth is:

- `../../../cmd/web/frontend/docs/ui-design-system.md`
- `../../../cmd/web/frontend/docs/design.md`

Before editing frontend UI code, read the source-of-truth documents above. Do
not start implementation until you have checked the sections relevant to the
change. This skill is intentionally strict because local UI changes can easily
drift from the product foundations, token system, and component ownership rules.

Read `ui-design-system.md` first for visual direction, foundations, tokens,
components, feature patterns, accessibility, and Storybook expectations. Read
`design.md` whenever placement, routing, ownership, dependency direction, or
component file shape is involved.

## When To Use

Use this skill for work under `cmd/web/frontend` that changes any of:

- React components or view composition.
- CSS Modules, layout, spacing, color, typography, surfaces, or responsive rules.
- Shared UI primitives under `src/components/ui`.
- App shell, header, sidebar, dialogs, breadcrumbs, or issue layout under
  `src/components/layout`.
- Feature UI under `src/features`.
- Storybook stories for components or feature views.

Do not use it for backend-only changes, CLI-only changes, or documentation-only
work unless the documentation describes frontend UI behavior.

## Workflow

1. Read the required design context before editing.
   - Always read `ui-design-system.md` sections `Foundations`, `Design Tokens`,
     `Token Model`, and `Token Governance`.
   - Read `design.md` sections `Top-Level Directories`, `Architecture Layers`,
     and `Component Shape` when creating, moving, or reshaping components.
2. Locate the owner of the UI concern before editing.
   - Shared, domain-independent controls belong in `src/components/ui`.
   - Shell-level navigation and dialogs belong in `src/components/layout`.
   - Product-specific compositions belong in `src/features/<feature>/components`.
3. Read the task-specific sections in `ui-design-system.md`.
   - Layout or shell changes: read `Layout System`, `Sidebar`, `Header`,
     `Breadcrumb`, `Content padding`, and `Responsive Breakpoints`.
   - Shared primitive changes: read `UI Primitives`, the target primitive
     section, `Accessibility Patterns`, `Storybook`, and `Adding New UI`.
   - Issue table or filtering work: read `Table`, `Issue Table View`, `Filter
     Options`, `Pagination`, and `Status and Priority Semantics`.
   - Badge/status/priority work: read `Badge`, `Status and Priority Semantics`,
     and the relevant token rows under `Color`.
   - Form/dialog work: read `Modal`, `Forms`, `Accessibility Patterns`, and
     `Responsive Breakpoints`.
   - Feature screen work: read the matching subsection under `Feature Patterns`
     before editing feature-owned CSS or JSX.
4. Reuse existing primitives and tokens first.
   - Colors come from `src/app/globals.css` or scoped component variables.
   - Spacing should follow the documented scale and nearby sibling components.
   - Icons should go through `IconProxy` when a matching icon exists.
5. Keep component boundaries narrow.
   - State and data shaping stay in the owning view or hook.
   - Reusable visual atoms stay prop-driven and domain-independent.
   - Feature-specific labels, filters, and status semantics stay in the feature.
6. Add or update Storybook when adding or reshaping a component exported from
   `index.tsx`.
   - Use `UI/...` for `src/components/ui`.
   - Use `Layout/...` for `src/components/layout`.
   - Use `Features/<Feature>/...` for `src/features/<feature>`.
7. Verify with the closest available checks.
   - Prefer the repository's frontend typecheck, lint, tests, or Storybook build
     commands when available.
   - For visual-only CSS changes, inspect the diff for token use, responsive
     behavior, and accidental hardcoded colors or spacing.

## Implementation Rules

- Use CSS Modules only. Do not introduce Tailwind, styled-components, runtime
  CSS-in-JS, or global utility classes for component-specific styling.
- Co-locate `index.tsx`, `index.module.css`, and `index.stories.tsx` for
  components that own a reusable UI surface.
- Do not hardcode hex colors for standard surfaces, text, borders, or accents
  when an existing token fits.
- Do not invent a parallel primitive if a documented UI primitive can be
  composed.
- Keep cards at `8px` radius or less unless the documented primitive uses a
  different value.
- Keep focus-visible states when customising interactive elements.
- Preserve accessibility contracts such as `aria-label`, `aria-expanded`,
  `aria-controls`, `role`, and screen-reader-only labels.
- Update `ui-design-system.md` when adding a new primitive, token role, surface
  pattern, or recurring feature pattern.

## Review Checklist

Before finishing, confirm:

- The component lives in the documented ownership layer.
- Styling uses existing tokens or a documented new token.
- CSS values match the documented scale and nearby component patterns.
- Responsive behavior follows existing breakpoints instead of inventing a new
  one without reason.
- Storybook coverage exists for new shared or feature components.
- The change does not regress keyboard, focus, or screen-reader behavior.
