# Web フロントエンド設計

この文書は `cmd/web/frontend` 配下の React/Vite application における
frontend の責務境界と配置ルールを定義します。

目的は、routing、feature UI、app shell UI、汎用 UI primitive を分離することです。新しい code は、まず所有責務で置き場所を決め、その後に再利用範囲で粒度を決めます。

## Routing

Route は `src/App.tsx` で `react-router-dom` を使って手動定義します。

`src/app/` tree は route entry file のためだけに使います。route file は route parameter を読み、layout hook を使い、route-level data loading を行い、feature component を render して構いません。`src/app/` 配下に `_components` directory は作りません。

scoped resource と detail page には path parameter を使います。

- `/issues` は全 project の issue を表示します。
- `/projects/:projectKey/issues` は単一 project の issue を表示します。
- `/projects/:projectKey` は `/projects/:projectKey/issues` に redirect します。
- `/projects/:projectKey/settings` は project workflow settings を表示します。
- `/issues/:id` は issue detail page を表示します。

## Top-Level Directories

frontend code を追加または移動するときは、次の責務境界を使います。

```text
src/
  app/          route entry file のみ
  components/   app-shell component と domain-independent UI primitive
  features/     page-level component と domain-aware feature component
  lib/          API client、generated type、i18n、store、shared runtime helper
  mocks/        local/mock 実行用の MSW handler と fixture data
  stories/      Storybook 専用 fixture builder と helper
```

dependency direction は単純に保ちます。

- `src/app` は `components`、`features`、`lib` から import できます。
- `src/features` は `components` と `lib` から import できます。
- `src/components/ui` は domain-independent に保ちます。
- `src/components/layout` は app shell、navigation、modal slot、layout-level context を扱えます。

## Component Shape

React component は 1 component につき 1 file にします。style を所有する component は次の形の directory に置きます。

```text
component-name/
  index.tsx
  index.module.css
  index.stories.tsx
```

component-specific style は `index.module.css` に置きます。component-specific test、story、helper type、小さな helper function は、その component の近くに置きます。

`types.ts`、`format.ts`、`rows.ts` のような shared non-component helper は、その feature または component group が所有する場合に component directory の近くへ置いて構いません。

## Route Entries

Route entry は `src/app/**/page.tsx` に置きます。

Route entry は薄く保ちます。Route entry が担当してよいことは次のとおりです。

- route parameter を読む。
- layout-level data hook を使う。
- route-specific data loading を行う。
- render する feature component を選ぶ。
- `Suspense` や route-level fallback behavior を提供する。

Route entry に reusable UI section、card、table、panel、domain presentation component を置きません。それらは対応する feature directory に置きます。

## Feature Components

Feature component は `src/features/<feature>/components/` に置きます。

Page-level view、domain-aware UI、props や behavior が product domain に結び付く component は feature directory に置きます。

現在の feature component group は次のとおりです。

```text
features/
  dashboard/components/dashboard-view/
  issues/components/
    board/
    card/
    conversation-page/
    issue-detail-page/
    issues-view/
    pane/
  projects/components/workflow-settings-view/
  settings/components/settings-view/
```

Feature page component はより小さな feature component を compose して構いません。view-model helper、test、story は同じ feature area に置きます。

### Feature Placement Policy

feature directory の判断基準は `code-react` の方針を基準にします。Atomic Design は補助観点であり、directory taxonomy にはしません。`atoms` と `molecules` は、domain language を持たず、visual style、variant、layout、interaction state、accessibility だけを表現できる場合に限り design-system layer に置きます。

次のいずれかに当てはまる component は `src/features/<feature>/components/` に置きます。

- issue、project、workflow、dashboard、settings の domain data を受け取る。
- props に status、priority、assignee、workflow、run、frontmatter などの domain term が入る。
- domain-specific display rule や許可される user action を表現する。
- 1 つの feature に属する page-level view または section である。
- generic UI primitive を組み合わせて product-specific experience を作る。

feature の state と rendering responsibility は分けます。

- route-level URL と loading concern は `src/app/**/page.tsx` に置く。
- feature view state と derived view model は feature component の近くに置く。
- domain vocabulary を持たない reusable display-only UI は `src/components/ui` に置く。
- API client、generated type、i18n、store、runtime helper は `src/lib` に置く。

同じ component が 2 回使われたという理由だけで `src/components/ui` に昇格しません。feature 外で具体的な再利用があり、product-domain props や behavior なしに表現できる場合だけ昇格します。props が広すぎる、boolean が増える、呼び出し側にとって分かりにくい場合は feature-owned のままにし、domain-free な primitive だけを切り出します。

## Shared UI Components

Domain-independent な UI primitive は `src/components/ui/` に置きます。

issue、project、dashboard の domain concept を知らない再利用 UI はこの directory に置きます。例は次のとおりです。

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

Shared Markdown rendering は複数 domain で使うため `components/ui/markdown/` が所有します。Shared modal rendering は generic portal slot を提供するため `components/ui/modal/` が所有します。

Domain-specific label、issue status transition、project workflow table、page-specific layout は `components/ui` に置きません。

## Layout Components

Application shell と global navigation component は `src/components/layout/` に置きます。

この directory は次を所有します。

- sidebar と header
- layout shell composition
- layout-level context と hook
- shell-level dialog の modal entry point
- default、issue detail、project layout などの route layout wrapper

Feature view は layout shell の中に render できますが、layout component が feature-specific presentation を所有しないようにします。

## Project Workflow Settings

Project workflow settings は `src/features/projects/components/workflow-settings-view/` 配下の feature component で表示します。

settings route は read-only で、`GET /api/v1/projects/{id}/workflow` から取得した同期済み `ProjectWorkflow` を表示します。frontend code で local `WORKFLOW.md` file を読んではいけません。

Workflow frontmatter は tree-friendly な table として表示します。Workflow body content は `src/components/ui/markdown` の shared Markdown renderer を使います。

## Storybook

`src/components/**/index.tsx` と `src/features/**/index.tsx` 配下のすべての React component は、対応する `index.stories.tsx` を持つ必要があります。

Storybook title は所有責務に合わせます。

- `src/components/ui` は `UI/...`
- `src/components/layout` は `Layout/...`
- `src/features/<feature>` は `Features/<Feature>/...`

Storybook 専用 fixture と builder は `src/stories/` に置きます。Production code から Storybook helper を import してはいけません。

Storybook static build output は `storybook-static/` に生成されますが、commit してはいけません。
