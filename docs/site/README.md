# Tasq Docs Site

This directory contains the Docusaurus documentation site for Tasq. The site publishes the developer and user-facing documentation that is curated from the repository documentation under `docs/`.

The English documentation is the primary source of truth. Keep Japanese documentation synchronized when adding or changing a topic.

## Commands

Run these commands from the repository root:

```sh
make dev-docs
make dev-docs-ja
make dev-docs-build
make dev-docs-open
```

Run these commands from `docs/site/` when working directly inside the Docusaurus project:

```sh
npm install
npm start
npm run start:ja
npm run build
npm run serve
```

The English development server uses the Docusaurus default port and the configured base URL:

```text
http://localhost:3000/tasq/
```

Start the Japanese development server separately when previewing localized content:

```sh
npm run start:ja
```

```text
http://localhost:3000/tasq/ja/
```

Docusaurus `start` serves one locale at a time. Use `npm run build` followed by `npm run serve` when you need to preview the production build with all locales.

## Directory Layout

```text
docs/site/
  docs/                    Docusaurus docs content.
    getting-started/       Overview, installation, tutorial, setup, and concepts.
    guides/                User-facing task and operations guides.
    reference/             CLI, API, configuration, and schema reference.
  static/img/              Static diagrams and other docs images.
  i18n/ja/                 Japanese localized Docusaurus content.
  src/
    css/custom.css         Site-wide theme overrides.
  docusaurus.config.ts     Site configuration, base URL, locales, navbar, footer.
  sidebars.ts              Docs sidebar structure.
  package.json             npm scripts and Docusaurus dependencies.
  package-lock.json        Locked npm dependency graph.
  tsconfig.json            TypeScript configuration for the site.
```

Generated directories are not source files:

- `node_modules/` is created by `npm install`.
- `.docusaurus/` is created by Docusaurus during development and build steps.
- `build/` is created by `npm run build`.

## Content Model

Source content for the published docs lives under `docs/site/docs/`. The docs-site is the user-facing documentation source of truth. Internal design documents under the repository `docs/design/` directory remain developer-facing design records and may contain lower-level implementation detail.

See `sidebars.ts` for the current page list and section structure; the sidebar is explicit. When adding a new page under `docs/site/docs/`, update `sidebars.ts` if the page should appear in the navigation.

Use Mermaid for normal diagrams. Keep hand-authored SVG diagrams under `static/img/` when a page needs a richer static overview.

## Localization

Docusaurus is configured with English as the default locale and Japanese as an additional locale.

Repository documentation outside this site usually keeps English and Japanese files as synchronized pairs, such as `docs/development.md` and `docs/development.ja.md`. Apply the same rule when the docs-site README or source documentation introduces a topic that needs both languages.

For Docusaurus pages, keep the English source under `docs/site/docs/` and the Japanese localized page under `docs/site/i18n/ja/docusaurus-plugin-content-docs/current/`.

## Editing Guidelines

- Keep site configuration in `docusaurus.config.ts`.
- Keep navigation structure in `sidebars.ts`.
- Keep visual styling in `src/css/custom.css` or page-local CSS modules.
- Do not edit generated files under `.docusaurus/`, `build/`, or `node_modules/`.
- Run `make dev-docs-build` before publishing changes that affect site content, configuration, or styling.
