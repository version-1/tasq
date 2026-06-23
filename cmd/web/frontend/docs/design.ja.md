# Web フロントエンド設計

この文書は `cmd/web/frontend` 配下の React/Vite frontend における local structure rule を定義します。

## Routing

Route は `src/App.tsx` で `react-router-dom` を使って手動定義します。

scoped resource page と detail page には path parameter を使います。issue list は全 project 用に `/issues`、単一 project 用に `/projects/:projectKey/issues` を使います。project detail page は `/projects/:projectKey` を使い、`/projects/:projectKey/issues` に redirect します。project detail の固定 tab は `/projects/:projectKey/issues` と `/projects/:projectKey/settings` です。issue detail は `/issues/:id` を使い、source file は `src/app/issues/[id]/` 配下に置きます。

## Component Placement

1 component につき 1 file を使います。style を所有する component は次の形の directory に置きます。

```text
component-name/
  index.tsx
  index.module.css
```

component-specific style はその component の `index.module.css` に置きます。`types.ts` のような shared non-component helper は component directory の近くに置いて構いませんが、複数の React component を同じ file に入れることは避けます。

route-local component は各 route の `_components` directory に置きます。page state、hook、feature component を compose する場所であり、再利用する domain UI の共有場所にはしません。

issue list route では、route composition だけを `src/app/issues/_components/issues-view/` 配下に置きます。

```text
issues-view/
  index.tsx
  index.module.css
```

top-level の `issues-view/index.tsx` は `IssuesView` component だけを持ち、再利用する issue-domain UI は `src/features/issues/` から import します。

route をまたいで共有する issue-domain UI component は `src/features/issues/components/` 配下に置きます。

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

`src/components/ui/` は shared modal component を含む domain-independent な design-system primitive 用に保ち、`src/components/layout/` は application shell と global layout component 用に保ちます。

project detail tab page は `src/app/projects/[projectKey]/` 配下に置きます。tab-specific UI は各 route の `_components` directory に置きます。settings tab は read-only で、`GET /api/v1/projects/{id}/workflow` から取得した同期済み `ProjectWorkflow` を表示します。frontend code で local `WORKFLOW.md` file を読んではいけません。
