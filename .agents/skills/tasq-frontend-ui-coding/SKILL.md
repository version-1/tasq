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

Read the UI design system document before changing UI code. Read the frontend
design document when placement, routing, component ownership, or Storybook
location is involved.

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

1. Locate the owner of the UI concern before editing.
   - Shared, domain-independent controls belong in `src/components/ui`.
   - Shell-level navigation and dialogs belong in `src/components/layout`.
   - Product-specific compositions belong in `src/features/<feature>/components`.
2. Read the relevant section in `ui-design-system.md`.
   - For table work, read `Table`, `Issue Table View`, and `Filter Options`.
   - For badges, read `Badge` and `Status and Priority Semantics`.
   - For app shell changes, read `Layout System`, `Sidebar`, `Header`, and
     `Content padding`.
   - For new components, read `Storybook` and `Adding New UI`.
3. Reuse existing primitives and tokens first.
   - Colors come from `src/app/globals.css` or scoped component variables.
   - Spacing should follow the documented scale and nearby sibling components.
   - Icons should go through `IconProxy` when a matching icon exists.
4. Keep component boundaries narrow.
   - State and data shaping stay in the owning view or hook.
   - Reusable visual atoms stay prop-driven and domain-independent.
   - Feature-specific labels, filters, and status semantics stay in the feature.
5. Add or update Storybook when adding or reshaping a component exported from
   `index.tsx`.
   - Use `UI/...` for `src/components/ui`.
   - Use `Layout/...` for `src/components/layout`.
   - Use `Features/<Feature>/...` for `src/features/<feature>`.
6. Verify with the closest available checks.
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
