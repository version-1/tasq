# Web UI 設計

Web UI は `cmd/web` から配信される Vite + React + TypeScript の single-page app です。Go server は production build を embed し、SPA fallback route を配信し、API call を proxy します。

## Routes

User-facing view は React Router route で分割します。

- `/issues`
- `/projects/:projectKey`
- `/projects/:projectKey/issues`
- `/projects/:projectKey/settings`
- `/issues/:id`
- `/dashboard`
- `/settings`

root `/` route は `/issues` に redirect します。
project detail root `/projects/:projectKey` は `/projects/:projectKey/issues` に redirect します。
project detail page は固定で 2 つの tab を持ちます。

- `issues`: 選択 project に scope した既存 issue board を表示します。
- `settings`: 選択 project の同期済み `WORKFLOW.md` を read-only data として表示します。

project settings tab は既存の issue-tracker endpoint `GET /api/v1/projects/{id}/workflow` から読み込みます。`ProjectWorkflow.body` は sanitize 済み Markdown として表示し、`ProjectWorkflow.frontmatter` は recursive key-value tree として表示し、`ProjectWorkflow.updatedAt` を同期時刻として表示します。frontend code から local `WORKFLOW.md` file を直接読んではいけません。編集 control も表示しません。

## Rendering Model

Web UI は完全に browser 上で動きます。browser state、API call、routing、translation concern は `cmd/web/frontend/src` 配下の React code に置きます。

Go server の責務は次のとおりです。

- `go:embed` で `cmd/web/frontend/dist` を配信する。
- non-API SPA route に `index.html` を返す。
- `/tracker/*` を issue-tracker に proxy する。
- `/orchestrator/*` を orchestrator に proxy する。

`cmd/web` に server-rendered UI behavior を追加しないでください。runtime configuration が必要な場合は、build-time browser environment variable よりも明示的な Go flag や proxy path を優先します。

## API Client

issue-tracker API client は `docs/openapi/issue-tracker.yml` から Orval で生成します。
orchestrator API client は `docs/openapi/orchestrator.yml` から Orval で生成します。

OpenAPI definition を変更した場合は `cmd/web/frontend` から generator を実行します。

```sh
npm run generate:api
```

generated file は `cmd/web/frontend/src/lib/generated` 配下に置き、手動編集してはいけません。route-facing code は、generated operation を直接必要とする場合を除き、API type を `cmd/web/frontend/src/lib/types.ts` 経由で import します。

## Component Structure

navigation、summary loading、refresh handling、page selection などの shared shell concern は route-specific component directory の外に置きます。

複数 route で意図的に共有する common component は `cmd/web/frontend/src/components` 配下に置きます。

Page-specific component は owner route の `_components` directory に置きます。

- Issue page component は `cmd/web/frontend/src/app/issues/_components` に置きます。
- Project detail tab component は `cmd/web/frontend/src/app/projects/[projectKey]/**/_components` に置きます。
- Workspace page component は `cmd/web/frontend/src/app/workspace/_components` に置きます。
- Settings page component は `cmd/web/frontend/src/app/settings/_components` に置きます。

component は shared か route-specific かにかかわらず、component 名の directory を使います。implementation、CSS Module、component test は同じ directory に置きます。

```text
<component-name>/
├── index.tsx
├── index.module.css
└── index.test.tsx
```

## Internationalization

Web UI は display string に `react-i18next` を使います。

supported language:

- Japanese (`ja`)
- English (`en`)

user-facing UI text は `cmd/web/frontend/src/lib/i18n.ts` に置きます。component は hard-coded display string ではなく `useTranslation()` で translated text を表示します。user-provided issue content、API identifier、route path segment は untranslated のままで構いません。

## Styling

frontend は component と page styling に CSS Modules を使います。

`cmd/web/frontend/src/app/globals.css` は global token と base element reset に限定します。feature-specific style は markup を所有する component の隣に置きます。

global color token は [Web UI Color Palette](web-color-pallete.md) を参照してください。

Examples:

- `cmd/web/frontend/src/components/layout/index.module.css`
- `cmd/web/frontend/src/app/issues/_components/issues-view/index.module.css`

feature-specific class selector を `cmd/web/frontend/src/app/globals.css` に追加しないでください。
