# Web Frontend UI Design System

This document summarizes the design tokens, layout primitives, UI primitives,
and feature-level patterns that the React/Vite frontend under
`cmd/web/frontend` uses today. It complements
[docs/design.md](design.md), which defines routing and component placement
rules.

The goal is to give a single reference so that new screens look and behave
consistently with the existing app shell, issue board, issue detail page, table
view, workflow settings, dashboard, and dialogs.

## Source of Truth

- Tokens: `src/app/globals.css` defines CSS custom properties and global
  element defaults.
- Visual reference: `ui-design-system.html` provides a browser-friendly token
  and component pattern overview for quick inspection.
- UI primitives: `src/components/ui/<name>/index.module.css` owns the styles for
  each shared component.
- Layout: `src/components/layout/` owns the app shell, header, sidebar,
  breadcrumb, and shell-level dialogs.
- Feature patterns: `src/features/<feature>/components/<name>/index.module.css`
  owns domain-aware sections such as the issue board, issue detail page, table
  view, dashboard panels, and workflow settings panels.

CSS Modules are the only styling mechanism. There is no Tailwind, no
styled-components, no theme provider, and no runtime CSS-in-JS. Component
styles must stay co-located with the component in `index.module.css`.

## Foundations

Tasq is a work-focused issue and agent orchestration UI. It should feel quiet,
dense, legible, and operational. The interface should optimize scanning,
comparison, repeated action, and handoff clarity over decoration or
marketing-style presentation.

These foundations are the design decision layer above tokens and components.
When a local UI decision is not explicitly covered by a component rule, choose
the option that best preserves these principles.

### Product Personality

- **Quiet and utilitarian**: avoid decorative gradients, oversized hero
  treatments, ornamental illustrations, and one-off visual effects in product
  screens.
- **Dense but organized**: expose enough information for repeated operational
  work, but use predictable grouping, alignment, and table/card structure so the
  density remains scannable.
- **Systematic before expressive**: reuse existing tokens, surfaces, and
  primitives before introducing a new visual treatment.
- **State-forward**: status, priority, project, run state, and workflow state
  must be visually easy to compare without relying on color alone.

### Information Hierarchy

- Page-level hierarchy comes from layout, spacing, and type weight before color.
- Tables are for comparison and scanning. Use restrained row hover, clear
  headers, stable columns, and compact badges.
- Cards are for grouped work items, modals, and genuinely framed tools. Do not
  wrap page sections in decorative card stacks when a full-width layout or table
  is more direct.
- Primary actions should be visually scarce. Most screens should have one
  dominant action cluster and use neutral or tertiary actions for the rest.

### Interaction Principles

- Controls should look like controls before hover. Filter chips, menus, buttons,
  and tab triggers need a visible boundary, affordance, or active rule.
- Repeated workflows should minimize layout shift. Fixed-format elements such as
  tables, badges, controls, counters, and toolbars need stable dimensions.
- Status and priority should use labels plus icons, dots, or shape where
  appropriate. Color should reinforce meaning, not carry it alone.
- Keyboard and screen-reader behavior is part of the component contract, not a
  later enhancement.

### Layout And Density

- Start from the app shell rhythm: `--space-6` page padding, compact controls,
  and `--radius-sm` / `--radius-md` surfaces.
- Use compact type inside panels, tables, sidebars, and tools. Reserve large
  type for actual page titles.
- Prefer existing breakpoints (`1060px`, `900px`, `860px`, `720px`, `640px`)
  before adding a new one.
- Avoid nested cards. If a surface already frames content, inner groups should
  usually be borders, dividers, tables, or unframed layout.

### Accessibility Principles

- Every icon-only interactive element needs an accessible label on the wrapping
  control.
- Focus-visible styles must remain visible when customizing controls.
- State controls must expose their state (`aria-expanded`, `aria-controls`,
  checked state, selected tab state, menu roles) in the component that owns the
  interaction.
- Do not encode state only through color. Pair color with text, icon, dot,
  border, or position.

### Anti-Patterns

- Adding a new hex value for a role that already has a token.
- Creating feature-local buttons, badges, menus, or tables when a shared
  primitive can be composed.
- Introducing spacing values such as `13px` or `17px` to tune one screen
  instead of using the spacing scale.
- Using decorative cards, gradients, large hero treatments, or illustrative
  sections in operational product views.
- Moving feature semantics such as issue status labels into domain-independent
  UI primitives.

## Design Tokens

Design tokens are the implementation contract for the foundations above. All
global tokens are declared on `:root` in `src/app/globals.css`. Components must
reference tokens through CSS variables and must not hardcode values for
surfaces, text, borders, status tones, spacing, radius, shadows, z-index, or
font stacks when a matching token exists.

### Token Model

Tasq uses three practical token layers:

| Layer | Purpose | Examples |
| ----- | ------- | -------- |
| Base tokens | Shared primitive values that define the system scale. | `--space-4`, `--radius-sm`, `--font-mono` |
| Semantic tokens | Product roles that can be used across components. | `--surface`, `--text`, `--danger`, `--surface-hover` |
| Component / feature tokens | Local aliases that adapt global roles to a component or feature. | `--badge-bg`, `--status-color`, `--ledger-rule` |

Component and feature tokens are allowed when they reduce duplication or keep a
feature palette scoped. They should still point back to global tokens whenever
the value is part of the shared system.

### Token Governance

- Add a token only when a value represents a repeated role or an intentional
  system scale step.
- Name tokens by role, not by appearance. Prefer `--surface-hover` over
  `--gray-100` when the value is used as an interaction surface.
- Keep feature-specific palettes local until at least two independent areas need
  the same role.
- Do not tokenise every number. Fixed component dimensions, typography sizes,
  and breakpoints can remain literal until they become repeated contracts.
- When introducing a new token, update this document, the Japanese counterpart,
  and any visual reference that displays the affected token family.
- When replacing a token, keep the old token only if existing call sites need a
  migration path. Otherwise update call sites in the same change.

### Color

| Token                | Value     | Role |
| -------------------- | --------- | --- |
| `--bg`               | `#FFFFFF` | Default page background |
| `--surface`          | `#FFFFFF` | Card, panel, modal, dropdown background |
| `--surface-strong`   | `#0A0A0A` | Reserved strong surface (rare use) |
| `--border`           | `#E5E5E5` | Default border for panels, fields, tables |
| `--text`             | `#0A0A0A` | Default foreground text |
| `--muted`            | `#666666` | Secondary text, metadata, helper copy |
| `--accent`           | `#0A0A0A` | Primary accent (same as `--primary-black`) |
| `--accent-color`     | `#7C3AED` | Native `accent-color` for form controls |
| `--glow-color`       | `#A78BFA` | Reserved purple glow accent |
| `--accent-strong`    | `#1A1A1A` | Hover/pressed variant of accent |
| `--danger`           | `#B42318` | Errors, destructive feedback |
| `--warning`          | `#A15C07` | Caution, approval prompts |
| `--info-accent`      | `#2563EB` | Informational emphasis and request-approval card accent |
| `--info-bg`          | `#EFF6FF` | Informational callout background |
| `--info-border`      | `#BFDBFE` | Informational callout border |
| `--info-text`        | `#1E3A8A` | Informational callout text |
| `--ok`               | `#087443` | Success, ready state |
| `--primary-black`    | `#0A0A0A` | Primary button background, headings |
| `--dark-gray`        | `#1A1A1A` | Strong icon color, split-button divider |
| `--medium-gray`      | `#666666` | Inactive tab, label color |
| `--light-gray`       | `#E5E5E5` | Border, switch track, ledger rule |
| `--extra-light-gray` | `#F5F5F5` | Hover surface, code-block background, chip background |
| `--white`            | `#FFFFFF` | Reserved pure white |
| `--surface-wash`     | `#FBFBFB` | App shell and low-contrast page wash |
| `--surface-hover`    | `#F0F0F0` | Sidebar active and hover row surface |
| `--surface-row-hover`| `#FAFAFA` | Table row hover surface |
| `--control-strong`   | `#111111` | Strong action hover / filter apply |
| `--control-strong-hover` | `#252525` | Strong action hover variant |
| `--control-divider`  | `#303030` | Split-button internal divider |
| `--markdown-link`    | `#0075FF` | Markdown links and task-list checkbox accent |
| `--markdown-link-hover` | `#005FCC` | Markdown link hover and focus tone |
| `--markdown-link-visited` | `#0075FF` | Visited Markdown links aligned with the link accent |
| `--markdown-checkbox` | `#0075FF` | Task-list checkbox accent aligned with Markdown links |
| `--markdown-inline-code-bg` | `#EAF4FF` | Inline code background in Markdown text |
| `--markdown-inline-code-text` | `#005FCC` | Inline code text in Markdown text |
| `--markdown-quote-bg` | `#F0F7FF` | Blockquote background in Markdown text |
| `--markdown-quote-border` | `#0075FF` | Blockquote left accent in Markdown text |
| `--markdown-quote-text` | `#174A7C` | Blockquote text in Markdown text |

The following groups record the dark-mode values for the semantic tokens above.

| Tokens | Dark values |
| --- | --- |
| `--bg` / `--white` / `--surface-wash` | `#111315` |
| `--surface` / `--surface-strong` | `#181B1F` / `#F0F3F6` |
| `--border` / `--light-gray` | `#30363D` |
| `--text` / `--accent` / `--primary-black` | `#F0F3F6` |
| `--muted` / `--medium-gray` / `--control-divider` | `#9DA7B3` |
| `--dark-gray` / `--extra-light-gray` | `#D0D7DE` / `#21262D` |
| `--surface-hover` / `--surface-row-hover` | `#262C36` / `#1C2128` |
| `--accent-color` / `--glow-color` | `#A78BFA` / `#C4B5FD` |
| `--accent-strong` / `--control-strong` / `--control-strong-hover` | `#FFFFFF` / `#F0F3F6` / `#FFFFFF` |
| `--danger` / `--warning` / `--ok` | `#FF7B72` / `#F2CC60` / `#3FB950` |
| `--info-accent` / `--info-bg` / `--info-border` / `--info-text` | `#79C0FF` / `#12253D` / `#28547A` / `#B6D7FF` |
| `--markdown-link` / `--markdown-link-hover` / `--markdown-link-visited` / `--markdown-checkbox` | `#79C0FF` / `#A5D6FF` / `#C4B5FD` / `#79C0FF` |
| `--markdown-inline-code-bg` / `--markdown-inline-code-text` | `#1D2D3D` / `#A5D6FF` |
| `--markdown-quote-bg` / `--markdown-quote-border` / `--markdown-quote-text` | `#172536` / `#58A6FF` / `#B6D7FF` |

Status, priority, project, approval, toast, and filter-chip tones are global
tokens. Components may assign them to local variables such as `--badge-bg`, but
the source value must come from `:root`.

Each three-value state row is `accent / background / text` and documents both
theme values. This table, together with the semantic-token table above, is the
complete color-token reference; `src/app/globals.css` remains the executable
source of truth.

| Role and tokens | Light values | Dark values |
| --- | --- | --- |
| Priority high (`--priority-high-accent`, `--priority-high-bg`, `--priority-high-text`) | `#D97706` / `#FEF3C7` / `#92400E` | `#F2CC60` / `#3D310F` / `#F8E3A1` |
| Priority normal (`--priority-normal-accent`, `--priority-normal-bg`, `--priority-normal-text`) | `#0284C7` / `#E0F2FE` / `#075985` | `#58A6FF` / `#112D4A` / `#A5D6FF` |
| Priority low (`--priority-low-accent`, `--priority-low-bg`, `--priority-low-text`) | `#9CA3AF` / `#F3F4F6` / `#4B5563` | `#8B949E` / `#24292F` / `#C9D1D9` |
| Status backlog (`--status-backlog-accent`, `--status-backlog-bg`, `--status-backlog-text`) | `#8A958D` / `#F3F4F6` / `#4B5563` | `#8B949E` / `#24292F` / `#C9D1D9` |
| Status ready (`--status-ready-accent`, `--status-ready-bg`, `--status-ready-text`) | `#0E8F70` / `#DFF8EF` / `#047857` | `#56D364` / `#12372A` / `#7EE787` |
| Status in progress (`--status-in-progress-accent`, `--status-in-progress-bg`, `--status-in-progress-text`) | `#2F6FB3` / `#E0F2FE` / `#075985` | `#58A6FF` / `#112D4A` / `#A5D6FF` |
| Status review (`--status-review-accent`, `--status-review-bg`, `--status-review-text`) | `#B7791F` / `#FEF3C7` / `#92400E` | `#E3B341` / `#3D2E0C` / `#F2CC60` |
| Status done (`--status-done-accent`, `--status-done-bg`, `--status-done-text`) | `#2F7D4F` / `#DCFCE7` / `#166534` | `#3FB950` / `#12372A` / `#7EE787` |
| Status blocked (`--status-blocked-accent`, `--status-blocked-bg`, `--status-blocked-text`) | `#9A4F14` / `#FFEDD5` / `#9A3412` | `#F0883E` / `#3D260F` / `#F5B77A` |
| Status failed (`--status-failed-accent`, `--status-failed-bg`, `--status-failed-text`) | `#B42318` / `#FEE2E2` / `#991B1B` | `#FF7B72` / `#3D1F22` / `#FFA198` |
| Status muted (`--status-muted-accent`, `--status-muted-bg`, `--status-muted-text`) | `#667085` / `#F3F4F6` / `#4B5563` | `#8B949E` / `#24292F` / `#C9D1D9` |
| Project (`--project-bg`, `--project-text`) | `#EEF6F8` / `#14636F` | `#12343B` / `#7DD3DC` |
| Approval (`--approval-bg`) | `#FFF2CC` | `#3D310F` |
| Toast error (`--toast-error-accent`, `--toast-error-bg`, `--toast-error-border`, `--toast-error-text`) | `#DC2626` / `#FEE2E2` / `#FECACA` / `#991B1B` | `#FF7B72` / `#3D1F22` / `#6E2B32` / `#FFA198` |
| Toast success (`--toast-success-accent`, `--toast-success-bg`, `--toast-success-icon`) | `#16A34A` / `#DCFCE7` / `#15803D` | `#56D364` / `#12372A` / `#7EE787` |
| Filter chip (`--filter-chip-bg`, `--filter-chip-border`) | `#F3F3F3` / `#EEEEEE` | `#24292F` / `#30363D` |
| Backdrop (`--backdrop`) | `rgb(20 20 20 / 36%)` | `rgb(0 0 0 / 60%)` |

Status tones used by table rows (`statusToneClassName`) map to a single
`--status-color`. They are declared in
`src/features/issues/components/status-badge/index.module.css` and are reused by
`src/components/ui/table/index.module.css` through the same status-aware row
class.

The table view declares a scoped "ledger" palette so that the issue table reads
like an accounting ledger:

| Token             | Source token |
| ----------------- | --------- |
| `--ledger-ink`    | `--text` |
| `--ledger-muted`  | `--muted` |
| `--ledger-rule`   | `--border` |
| `--ledger-surface`| `--surface` |
| `--ledger-wash`   | `--surface-wash` |

These are local aliases in `table-view` and `filter-options`, and resolve to
global semantic tokens. Generic table primitives read them with
`var(--ledger-rule, var(--border))` so they degrade to the global tokens
elsewhere.

### Dark Theme

Light mode tokens are declared on `:root`. Setting `data-theme="dark"` on an
ancestor (normally `html`) overrides every color and shadow token while
preserving component token references. The semantic-token table and state table
in this document include the corresponding dark values.

`--bg` and `--surface-wash` resolve to `#111315` in dark mode; `--surface`
resolves to `#181B1F`. All shadow tokens are also overridden in the theme
selector to preserve hierarchy on dark surfaces. Do not add component-level
shadow color literals.

`index.html` sets `html[data-theme]` before React mounts. A valid
`localStorage` value at `tasq.theme` (`light` or `dark`) takes precedence; when
there is no valid saved value, the initial theme follows `prefers-color-scheme`.
The layout theme switch persists an explicit value with the same key. Only when
no value is stored does the runtime hook follow subsequent operating system
preference changes.

Theme changes must be implemented by changing `data-theme` and semantic token
values. Components must not add theme-specific selector branches or hardcoded
color and shadow values.

### Typography

- Base font family: `--font-sans` (set on `body`).
- Monospace font stack for IDs and code-like values: `--font-mono`.
- Heading sizes seen in production:
  - Page title (`h1`, header): `30px / 800`.
  - Issue detail title (`h2`): `28px / default weight`, drops to `22px` under
    `640px`.
  - Dashboard / page section (`h2`): `22px`.
  - Section heading (`h3`, panel): `16px`.
  - Sidebar brand: `34px / 800`.
- Body text: `14px` for table cells, helper copy, labels.
- Metadata / caption: `12–13px` with `--muted`.
- Form field label: `13px / 700`.
- Form field input: inherited font, `400` weight.

Buttons and form controls inherit the page font via the global
`button, input, select { font: inherit }` reset.

### Spacing

Spacing tokens are declared in `:root` and use a 4px-based scale with half
steps for values that already recur in the UI:

| Token | Value |
| ----- | ----- |
| `--space-0` | `0` |
| `--space-1` | `4px` |
| `--space-1-5` | `6px` |
| `--space-2` | `8px` |
| `--space-2-5` | `10px` |
| `--space-3` | `12px` |
| `--space-3-5` | `14px` |
| `--space-4` | `16px` |
| `--space-4-5` | `18px` |
| `--space-5` | `20px` |
| `--space-5-5` | `22px` |
| `--space-6` | `24px` |
| `--space-7` | `28px` |
| `--space-8` | `32px` |
| `--space-10` | `40px` |
| `--space-20` | `80px` |

Use these tokens for margin, padding, gap, and positioning offsets when the
value represents layout spacing. Keep sizing, radius, typography, and breakpoint
values literal unless they become documented tokens.

### Radius

| Token | Value | Role |
| ----- | ----- | ---- |
| `--radius-xs` | `5px` | Chip outline inside filter popovers |
| `--radius-sm` | `6px` | Button, form field, small surface |
| `--radius-md` | `8px` | Panel, card, modal dialog, table wrapper |
| `--radius-lg` | `10px` | Toast container |
| `--radius-pill` | `999px` | Pill badge, switch track and thumb, timestamp chip |

### Shadow

| Token | Role |
| ----- | ---- |
| `--shadow-panel` | Standard panel and section surface |
| `--shadow-card` | Issue card surface |
| `--shadow-dropdown` | Dropdown and context menu |
| `--shadow-dialog` | Modal dialog |
| `--shadow-toast` | Toast container |
| `--shadow-switch-thumb` | Switch thumb |
| `--shadow-header-action` | Header action cluster |
| `--shadow-filter-trigger` | Filter trigger resting state |
| `--shadow-filter-active` | Filter trigger expanded/focused state |
| `--shadow-filter-popover` | Filter popover |

### Z-Index

| Token | Value | Role |
| ----- | ----- | ---- |
| `--z-menu` | `10` | Context menu |
| `--z-sticky` | `10` | Sticky header |
| `--z-popover` | `20` | Filter popover |
| `--z-dialog` | `30` | Dialog backdrop |
| `--z-toast` | `1000` | Toast stack |

Use the smallest layer that satisfies the role. Do not introduce ad-hoc
z-index values outside these layers without a written reason.

## Layout System

The application shell is a two-column grid declared in
`src/components/layout/index.module.css`:

```text
.appFrame {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  min-height: 100vh;
}
```

- `260px` sidebar on the left, fluid content on the right.
- The content column has `min-width: 0` so wide tables can scroll without
  forcing the layout.
- On `≤ 900px` the grid collapses to a single column and the sidebar becomes a
  top strip with a bottom border.
- The shell background is `var(--surface-wash)`. Cards, panels, dialogs, and
  dropdowns sit on `var(--surface)`.

### Sidebar

- Vertical stack with `gap: 22px` and `padding: 28px 18px 18px`.
- Brand link: `34px / 800` in `var(--primary-black)`.
- Primary navigation: bottom border separator, `4px` row gap.
- Project list: own `8px` row gap, plus a section header in
  `12px / 700 / uppercase / --medium-gray`.
- Nav rows: `44px` min height, `12px` icon gap, `--radius-sm`, and
  `--surface-hover` hover.
- The active route uses `--surface-hover` background (no border accent).
- Bottom of sidebar: settings link separated by `border-top`, a theme switch
  row with the same separator, and the current release version in small muted
  text. The commit is available only as hover text when known.

### Header

- Sticky to `top: 0` with `z-index: 10` and a `1px` bottom border.
- Three rows:
  - Utility row: notification button, search input (`360px` min width with
    keyboard hint), more button.
  - Title row: page title (`30px / 800`) and a split create button cluster.
  - View row: tabs.
- Tabs: `36px` gap; active tab uses `box-shadow: inset 0 -2px 0 var(--primary-black)`
  and inherits foreground from `--primary-black`.
- Split create button: `44px` tall, `--primary-black` background, `#111111`
  hover, internal `#303030` separator.

### Breadcrumb

`src/components/layout/header/breadcrumb` renders a horizontal list:

- `14px` text in `--medium-gray`.
- `8px` gap between items.
- Current segment uses `--primary-black` with weight `600`.

### Content padding

`.content` in the shell adds `24px` of breathing room around the routed
feature. Feature views that need to bleed into the gutters (such as the issue
table view) cancel this with `margin: -24px` and reapply their own padding.

## UI Primitives

UI primitives live under `src/components/ui/`. They are domain-independent and
must remain reusable.

### Button (`ui/button`)

- `Button` exposes `primary`, `positive`, `secondary`, and `tertiary` variants.
  Primary uses a black fill, positive uses the ready-state green fill,
  secondary uses a neutral outline, and tertiary uses muted text without a
  visible border.
- The default size is `40px` high with `0 16px` horizontal padding. The
  `compact` size matches Issue Card quick actions at `32px` high with `0 12px`
  horizontal padding.
- Storybook exposes each variant independently and together for visual comparison.
- The global `button` reset gives every other native button a neutral
  surface with `1px` border, `6px` radius, `8px 10px` padding, and a disabled
  state in `--muted`.

Custom action variants (filter `Apply`, dialog submit, header `Create`) are
hand-styled in their owning module but always reuse: `--primary-black`,
`6px` radius, `--white` text, `--accent-strong` (`#1A1A1A`) on hover.

### Badge (`ui/badge`)

`Badge` renders a labelled chip plus an optional icon. The component picks a
visual variant via the `variant` prop:

- `project` — soft teal chip used in the issue table and card.
- `priority-high` / `priority-normal` / `priority-low` — colored pills with a
  matching accent dot.
- `status-backlog` / `status-ready` / `status-in-progress` / `status-review` /
  `status-done` / `status-blocked` / `status-failed` / `status-muted` — same
  pill geometry with status-specific accent.

All variants share `12px / 600` text inside `4px 10px` padding with a fully
rounded `999px` radius. Icons inherit `--badge-accent`; dot accents inherit
the same custom property.

### Switch (`ui/switch`)

- `42px × 24px` track, `20px` thumb, `999px` radius.
- Inactive track: `--light-gray`. Active track: `--primary-black`.
- Thumb shifts `18px` horizontally with a `160ms` ease transition.
- Focus ring: `2px solid var(--primary-black)` with `2px` offset.

### Pagination (`ui/pagination`)

- Inline cluster aligned to the end with `12px` gap.
- Summary text uses `14px` in `--ledger-muted` (falls back to
  `--muted-foreground`).
- Stacks to `justify-content: space-between` under `860px`.

### Table (`ui/table`)

- Wrapper: `1px` border in `--ledger-rule` with `6px` radius and `920px` min
  height so the empty state still feels intentional.
- Headers: `12px / 600`, ledger-muted color, sticky-feeling whitespace.
- Cells: `14px`, `14px 12px` padding, `1px` bottom border using
  `--ledger-rule`.
- Hover row: `#FAFAFA` background.
- ID cell uses the monospace stack.
- Sort buttons: chevron icon with `0.48` opacity unless active.

### Context Menu (`ui/context-menu`)

- Positioned absolutely against its trigger, opens `40px` below. Default
  placement aligns the menu end edge to the trigger; `bottom-start` opens from
  the trigger toward the right.
- `220px` minimum width with `8px` padding, `4px` row gap, `8px` radius, and a
  card shadow. Item labels stay on one line and the menu grows with its content.
- Items: `8px 8px` padding, `6px` radius, `--extra-light-gray` hover.
- Destructive items use `variant="danger"` and color text with `--danger`.
- Group label: `12px` muted, used to title menu groups.
- Help footer: small muted paragraph for keyboard hints.

### Modal (`ui/modal`)

`ModalOutlet` exposes a portal slot. Dialog presentation is owned by the
calling layout component. Recurring rules (see `components/dialog/add-issue`
and `components/dialog/add-project`):

- Backdrop: `rgb(20 20 20 / 36%)`, full viewport, `padding: 80px 24px 24px`,
  `z-index: 30`.
- Dialog surface: white card, `8px` radius, `0 20px 50px rgb(0 0 0 / 18%)`
  shadow, `min(100%, 560px)` (or `520px` for project dialog) width.
- Form layout: grid with `16px` gap; field grid drops to `1fr` on `≤ 900px`.
- Submit button reuses the primary-black look. Cancel/close buttons use the
  default neutral button.

### Markdown (`ui/markdown`)

- Reusable Markdown renderer powered by `react-markdown`, `remark-gfm`,
  `rehype-sanitize`, and `shiki`.
- Headings render with compact type, `40px` top / `24px` bottom spacing, and a
  `1px` underline using `--border` so long issue descriptions are easier to
  scan.
- Links use the blue `--markdown-link` / `--markdown-link-hover` tokens, with
  underline and focus outline preserved.
- GFM task list checkboxes use the same blue accent as Markdown links through
  the `--markdown-checkbox` token.
- Paragraphs, lists, blockquotes, inline code, and fenced code blocks get
  component-owned spacing and token-based colors. Blockquotes use the
  `--markdown-quote-*` tokens, and inline code uses the
  `--markdown-inline-code-*` tokens so both remain visible inside prose.
- Fenced code blocks use Shiki token highlighting with the `github-light`
  theme. Unsupported language names fall back to plain text.
- Tables get a `1px` border on every cell, `--extra-light-gray` headers,
  `8px 10px` padding, and horizontal scrolling when content is wider than the
  container.
- Body wraps with `overflow-wrap: anywhere` for long URLs and IDs.
- `attachment://<id>` references are rewritten to
  `/tracker/api/v1/attachments/<id>/content`.

### Markdown Editor (`ui/markdown-editor`)

- Wraps the shared Markdown renderer with read/edit modes for issue
  descriptions and Markdown form fields.
- Read mode preserves the normal Markdown presentation and can show a compact
  pencil icon action in the header.
- Edit mode uses underlined `Raw` / `Preview` tabs. `Raw` is a textarea with
  configurable rows; `Preview` renders the current draft immediately through
  `ui/markdown`.
- Dialogs can opt into a stable tab panel height so switching between `Raw` and
  `Preview` does not resize the surrounding form.
- Save/cancel controls are optional so dialogs can own their own submit/cancel
  actions while reusing the same editing surface.
- Failed saves keep the draft in edit mode and display the error message below
  the editor.

### Panel Message (`ui/pannel-message`)

- Wide panel surface (`max-width: 960px`) used for terminal errors and empty
  states. Heading is `--primary-black`, body is `--danger`.

### Toast (`ui/toast`)

- Fixed stack at top-right (`top: 24px`, `right: 24px`).
- Toast container: `10px` radius, `14px 16px` padding, `6px` left accent bar in
  the toast color.
- `error` accent: `#DC2626`. `success` accent: `#16A34A`.
- Icon bubble: `40px` circle with tinted background and accent foreground.
- Stack collapses to full width under `720px`.

### Icon Proxy (`ui/icon-proxy`)

- Centralised wrapper around a curated subset of `lucide-react` icons.
- Default `size={16}` and `strokeWidth={2}`; callers pass `name`, optional
  `className`, `size`, and `strokeWidth`.
- Always renders with `aria-hidden="true"` and `focusable="false"`.

When introducing a new icon, register it in the `icons` map so the prop type
stays a literal union.

## Feature Patterns

Feature components compose the primitives above into product-specific
sections. The shapes below are the patterns that already exist; new feature
work should reuse them rather than invent parallel ones.

### Card Surface

Used by panels, sections, dialogs, and the dashboard:

```css
background: var(--surface);
border: 1px solid var(--border);
border-radius: 8px;
box-shadow: 0 1px 2px rgb(0 0 0 / 8%);
padding: 16px; /* or 18px when the section owns a header */
display: grid;
gap: 14px;
```

This is the recurring shape for `runs-section`, `comment-list`,
`issue-description`, `status-actions`, `basic-info-panel`, `workflow-meta-panel`,
`workflow-body-section`, and `frontmatter-section`.

### Issue Card

`features/issues/components/card`:

- `8px` radius card with a heavier `0 12px 32px rgb(0 0 0 / 8%)` shadow.
- Title row: `16px / 500` link plus a `32px` menu trigger.
- Metric row: icon + value pairs separated by `1px` left border per metric.
- Quick actions use the shared compact `Button`: Ready uses `positive`, Done
  uses `primary`, and Resolve uses `secondary`.
- A blocked card replaces the direct `Ready` transition with `Resolve`. The
  text-only neutral outlined button distinguishes this dialog-opening action
  from the direct green `Ready` transition. The action loads the latest blocker
  comment across all comment pages and opens a continue-with-comment dialog.
  The blocker is read-only context; free text or the built-in `Ok` / `Retry`
  shortcuts create the change request before the issue moves to `ready`.

### Issue Board

`features/issues/components/board`:

- 4-column grid with `260px` minimum column width and an x-axis scroll for
  narrow viewports.
- Columns separated by `1px` left rule; first column hides the rule.
- Column header: title + count chip + action buttons. Count chip is a
  `999px` pill on `--extra-light-gray`.
- Task list uses `10px` gap.

### Issue Table View

`features/issues/components/table-view` extends `ui/table` with the ledger
palette:

- Surrounds the table with the wash background and ledger rule borders.
- Toolbar is a `flex` row that becomes a stacked column under `860px`.
- Filter cluster lives inside a `<fieldset>` whose `legend` is visually hidden
  for screen readers.
- Reset button reads as a tertiary action (transparent background, ledger
  muted text, becomes ink on hover).

### Filter Options

`features/issues/components/filter-options`:

- Trigger summary chip: `40px` min height, `6px` radius, ledger surface,
  rotating chevron via `aria-expanded="true"`.
- Selected values are shown as `999px → 5px` chips inside the trigger.
- Popover: `8px` radius card with a `0 22px 46px` shadow, `420px` min width,
  fixed header (`58px`) with clear/cancel buttons, scrolling option list
  (`max-block-size: 420px`), action row with primary apply button (`#111111`).
- On `≤ 860px`, the trigger and popover stretch to full width.

### Issue Detail Page

`features/issues/components/issue-detail-page`:

- Page grid with `16px` gap; each section is the standard card surface.
- Details tab keeps the main content width constrained while `basic-info-panel`
  sits in the right-side column on wide screens and stacks above the main
  content under `900px`.
- `basic-info-panel` uses a one-column meta grid for issue ID, project,
  priority, status, assignee, created date, and updated date.
- `artifacts-section` presents a pull request as a compact text link with a type
  icon and an inline label, repository, and PR number reference.
- `meta-item` renders each entry with a top rule, `12px` muted `dt`, normal
  weight `dd`, and `overflow-wrap: anywhere` for IDs.
- `runs-section` wraps run rows in a single `8px` bordered group.
- `run-row` is a 2-column grid (`1fr` link area + `40px` copy button). The
  link is a 3-column inner grid with monospace IDs and ledger-muted timestamps.
- `comment-list` reuses the card shape and prefixes each comment with a
  `border-top` separator.
- Change-request actions expose a labelled menu group. `Write a comment…`
  opens the editor, while configured shortcuts submit their instruction body
  immediately. Shortcut labels and bodies are separate values so future
  configuration can keep compact labels for longer agent instructions.
- Reject actions use a neutral secondary split button so they remain weaker
  than the primary Done action. The main segment opens the editor and the
  chevron segment opens the shortcut menu. Both segments share disabled and
  submitting states.
- The issue-detail sidebar labels its status shortcuts as `Quick Action`:
  backlog shows positive `Ready`, ready shows secondary `Draft`, and review
  shows the secondary Reject split button before primary `Done`. Blocked issues
  show secondary `Resolve`.
- `status-actions` is a wrap-flex action row inside the same card.
- `issue-description` uses the shared markdown renderer with `1.65` line
  height, anchor color `--primary-black`, and bordered inline images.

### Conversation Page

`features/issues/components/conversation-page`:

- Page grid with `16px` gap.
- Timeline list of cards. Request-approval items use the `--info-*` tone,
  including an `--info-accent` left rule and `--info-bg` top callout.
- Commands and code blocks use a `6px` radius, `--extra-light-gray`
  background, and the monospace font stack.
- Exit code callout uses a `rgb(220 38 38 / 10%)` wash with a danger border
  and `--danger` text.

### Dashboard

`features/dashboard/components/dashboard-view`:

- Vertical grid with `18px` gap.
- Metric grid: 4 cards on the default breakpoint, 2 cards under `1060px`, 1
  card under `720px`.
- Detail grid: 2 cards that collapse to 1 under `720px`.
- Compact metric cards use `--extra-light-gray` with no shadow and a smaller
  `24px` value.
- Distribution bars: `10px` tall `999px` track on `--extra-light-gray` with a
  `--primary-black` fill.
- Run table reuses the card shape plus the standard table conventions.

### Workflow Settings

`features/projects/components/workflow-settings-view`:

- Single column `panelGrid` capped at `960px`.
- `workflow-meta-panel`: flex chip row of label/value pairs with `12px 16px`
  padding.
- `frontmatter-section` and `workflow-body-section`: standard card with an
  `18px` heading.
- `frontmatter-table`: `1px` bordered table inside a `8px` radius wrapper,
  monospace keys with `padding-inline-start: calc(var(--frontmatter-depth, 0) * 20px)`
  so nested keys indent visually. Header cells use `--extra-light-gray`.

### Issue Layout

`src/components/layout/issue/index.module.css` owns the issue subpage shell:

- `max-width: 1040px`, `14px` row gap, back link in `--primary-black`.
- Tabs use the same active-rule pattern as the global header
  (`box-shadow: inset 0 -2px 0 var(--primary-black)`).

## Forms

Form field conventions are visible in dialogs and the settings form:

- Field group: grid with `6px` gap. Label is `13px / 700` in `--dark-gray`
  (settings uses `--muted`).
- Input/select/textarea: `1px` `--border`, `6px` radius, `--primary-black`
  text, `10px` padding, `400` weight, `font: inherit`.
- Textareas use `resize: vertical`.
- Validation message: `13px` in `--danger`, no margin.
- Two-column form layout (`formGrid` in add-issue dialog) collapses to one
  column under `900px`.
- Submit row: flex row aligned end with `8px` gap; primary submit is black,
  secondary buttons keep the default neutral look.

Disabled controls reduce opacity to `0.6` and use `cursor: not-allowed`.

## Status and Priority Semantics

- Priority "high" reads as warning amber.
- Priority "normal" reads as informational sky blue.
- Priority "low" reads as neutral gray.
- Status `backlog`, `cancelled`, `duplicate`, `muted` all use neutral gray.
- Status `ready` uses success green.
- Status `in-progress` uses informational sky blue.
- Status `review` uses warning amber.
- Status `done` uses confirmation green (slightly darker than ready).
- Status `blocked` uses alert orange.
- Status `failed` uses danger red.

Run rows in the issue table apply `statusToneClassName(issue.status)` so each
row carries its own `--status-color`; reuse this helper anywhere a row needs to
inherit the status tone instead of duplicating the palette.

## Responsive Breakpoints

The codebase uses ad-hoc `max-width` breakpoints rather than a unified scale.
When adding rules, pick the closest existing breakpoint instead of inventing
new ones:

| Breakpoint | Where it is used |
| ---------- | ---------------- |
| `1060px`   | Dashboard metric/detail grids collapse to 2 columns. |
| `900px`    | App shell collapses to a single column. Header rows stack. Issue card and detail meta grids collapse. Dialog padding shrinks. |
| `860px`    | Table view toolbar stacks. Filter trigger stretches. Pagination switches to `space-between`. |
| `720px`    | Toasts go full width. Conversation page event header stacks. Dashboard metric grids collapse to 1 column. Issue tabs allow horizontal scroll. Run row inner grid collapses. |
| `640px`    | Issue detail title shrinks to `22px`. Comment header collapses. |

## Accessibility Patterns

- `Switch`, sort buttons, copy buttons, context menu items, and toasts expose
  meaningful `aria-label`s, often through i18n keys.
- Context menus apply `aria-haspopup`, `aria-expanded`, `aria-controls`, and
  `role="menu"` / `role="menuitem"`.
- Toasts wrap each item in `role="status"` and the stack in `aria-live="polite"`.
- Icons rendered through `IconProxy` are `aria-hidden="true"` and
  `focusable="false"`, so labels must live on the wrapping interactive element.
- Focus styles use `outline` or `box-shadow` ring rather than removing the
  outline. Keep `focus-visible` rules whenever you customise `focus` styles.

## Storybook

Every component under `src/components/**/index.tsx` and
`src/features/**/index.tsx` must ship an `index.stories.tsx`.

Use titles that match ownership:

- `UI/...` for `src/components/ui`.
- `Layout/...` for `src/components/layout`.
- `Features/<Feature>/...` for `src/features/<feature>`.

Storybook-only fixtures live in `src/stories`. Production code must not import
Storybook helpers.

Use the **Theme** toolbar to switch any story between the default light palette
and the dark palette. The Storybook preview applies the selected value to the
document's `data-theme` attribute, so the same global CSS tokens used by the
application are reviewed without changing component code.

## Adding New UI

1. Identify whether the new piece is a UI primitive, a layout shell concern,
   or a feature-aware component. Follow the placement rules in
   [design.md](design.md).
2. Reuse the existing tokens, surfaces, and patterns above. Only add a new
   token or shadow when none of the documented values fit the role.
3. Co-locate `index.tsx`, `index.module.css`, and `index.stories.tsx`.
4. Cover the new component with a Storybook entry under the correct
   ownership title.
5. Update this document when you introduce a new primitive, a new surface
   pattern, or a new token role.
