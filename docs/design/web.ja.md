# Web UI 設計

Web UI は、`cmd/web` から配信される Vite + React + TypeScript の single-page app です。Go サーバーは本番ビルドを埋め込み、SPA fallback route を配信し、API 呼び出しをプロキシします。

## Routes

ユーザー向けビューは React Router のルートで分割します。

- `/dashboard`
- `/dashboard/table`
- `/dashboard/stats`
- `/projects/:projectKey`
- `/projects/:projectKey/issues`
- `/projects/:projectKey/table`
- `/projects/:projectKey/settings`
- `/issues/:id`
- `/settings`

root `/` route は `/dashboard` にリダイレクトします。
project detail root `/projects/:projectKey` は `/projects/:projectKey/issues` にリダイレクトします。
dashboard page は固定で 3 つのタブを持ちます。

- `board`: 全 project の issue board を表示します。
- `table`: 全 project の issue をページング付き table で表示します。
- `stats`: dashboard metrics と実行中 agent summary を表示します。

project detail page は固定で 3 つのタブを持ちます。

- `issues`: 選択したプロジェクトに絞り込んだ既存の課題ボードを表示します。
- `table`: 選択したプロジェクトに絞り込んだページング付き issue table を表示します。
- `settings`: 選択したプロジェクトの同期済み `WORKFLOW.md` を読み取り専用データとして表示します。

project settings tab は、既存の issue-tracker エンドポイント `GET /api/v1/projects/{id}/workflow` から読み込みます。`ProjectWorkflow.body` はサニタイズ済み Markdown として表示し、`ProjectWorkflow.frontmatter` は再帰的な key-value tree として表示し、`ProjectWorkflow.updatedAt` は同期時刻として表示します。フロントエンドコードからローカルの `WORKFLOW.md` ファイルを直接読んではいけません。編集用コントロールも表示しません。

## Rendering Model

Web UI は完全にブラウザ上で動作します。ブラウザ状態、API 呼び出し、ルーティング、翻訳に関する責務は `cmd/web/frontend/src` 配下の React コードに置きます。

Go サーバーの責務は次のとおりです。

- `go:embed` で `cmd/web/frontend/dist` を配信する。
- non-API SPA route に `index.html` を返す。
- `/tracker/*` を issue-tracker にプロキシする。
- `/orchestrator/*` を orchestrator にプロキシする。

`cmd/web` に server-rendered UI behavior を追加しないでください。実行時設定が必要な場合は、ビルド時のブラウザ環境変数よりも、明示的な Go flag やプロキシパスを優先します。

## API Client

issue-tracker API client は `docs/openapi/issue-tracker.yml` から Orval で生成します。
orchestrator API client は `docs/openapi/orchestrator.yml` から Orval で生成します。

どちらかの OpenAPI definition を変更した場合は、`cmd/web/frontend` から generator を実行します。

```sh
npm run generate:api
```

生成ファイルは `cmd/web/frontend/src/lib/generated` 配下に置き、手動編集してはいけません。ルート向けコードは、生成された operation を直接必要とする場合を除き、API 型を `cmd/web/frontend/src/lib/types.ts` 経由で import します。

## Component Structure

ナビゲーション、サマリー読み込み、更新処理、ページ選択など、共有シェルに関する責務はルート固有のコンポーネントディレクトリの外に置きます。

複数のルートで意図的に共有する common component は `cmd/web/frontend/src/components` 配下に置きます。

ページ固有のコンポーネントは、所有元ルートの `_components` directory に置きます。

- Issue page component は `cmd/web/frontend/src/app/issues/_components` に置きます。
- Project detail tab component は `cmd/web/frontend/src/app/projects/[projectKey]/**/_components` に置きます。
- Workspace page component は `cmd/web/frontend/src/app/workspace/_components` に置きます。
- Settings page component は `cmd/web/frontend/src/app/settings/_components` に置きます。

コンポーネントは、shared か route-specific かにかかわらず、コンポーネント名の directory を使います。実装、CSS Module、コンポーネントテストは同じ directory に置きます。

```text
<component-name>/
├── index.tsx
├── index.module.css
└── index.test.tsx
```

## Internationalization

Web UI は表示文字列に `react-i18next` を使います。

対応言語:

- Japanese (`ja`)
- English (`en`)

ユーザーに表示する UI text は `cmd/web/frontend/src/lib/i18n.ts` に置きます。コンポーネントは hard-coded display string ではなく、`useTranslation()` で翻訳済みテキストを表示します。ユーザーが入力した課題本文、API identifier、route path segment は翻訳しないままで構いません。

## Styling

フロントエンドはコンポーネントとページのスタイルに CSS Modules を使います。

`cmd/web/frontend/src/app/globals.css` はグローバルトークンと基本要素のリセットに限定します。機能固有のスタイルは、その markup を所有するコンポーネントの隣に置きます。

グローバルカラートークンは [Web UI Color Palette](web-color-pallete.md) を参照してください。

Examples:

- `cmd/web/frontend/src/components/layout/index.module.css`
- `cmd/web/frontend/src/app/issues/_components/issues-view/index.module.css`

機能固有の class selector を `cmd/web/frontend/src/app/globals.css` に追加しないでください。
