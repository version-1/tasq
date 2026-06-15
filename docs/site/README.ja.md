# Tasq Docs Site

このディレクトリには、Tasq の Docusaurus ドキュメントサイトがあります。サイトは、リポジトリ内の `docs/` にある開発者向け・利用者向けドキュメントを公開用に整理します。

英語ドキュメントを一次情報として扱います。トピックを追加・変更する場合は、日本語ドキュメントも同期してください。

## コマンド

リポジトリルートから実行するコマンド:

```sh
make dev-docs
make dev-docs-build
make dev-docs-open
```

`docs/site/` の中で Docusaurus プロジェクトを直接扱う場合のコマンド:

```sh
npm install
npm start
npm run build
npm run serve
```

開発サーバーは Docusaurus のデフォルトポートと、設定済みの base URL を使います。

```text
http://localhost:3000/tasq/
```

## ディレクトリ構成

```text
docs/site/
  docs/                    Docusaurus の docs コンテンツ。
    design/                リポジトリ docs から取り込む architecture / design ページ。
    cli-reference.md       CLI reference ページ。
    getting-started.md     Getting started ページ。
    setup-guide.md         ローカル Codex / agent setup guide。
  i18n/ja/                 Docusaurus の日本語 localized content。
  src/
    css/custom.css         サイト全体のテーマ上書き。
  docusaurus.config.ts     サイト設定、base URL、locale、navbar、footer。
  sidebars.ts              docs sidebar の構成。
  package.json             npm scripts と Docusaurus 依存。
  package-lock.json        npm 依存グラフの lockfile。
  tsconfig.json            サイト用 TypeScript 設定。
```

次のディレクトリは生成物で、ソースファイルではありません。

- `node_modules/` は `npm install` で作成されます。
- `.docusaurus/` は Docusaurus の開発・build 時に作成されます。
- `build/` は `npm run build` で作成されます。

## コンテンツ構成

公開される docs のソースは `docs/site/docs/` に置きます。現在のサイトには次のページがあります。

- `getting-started.md`
- `setup-guide.md`
- `cli-reference.md`
- `design/architecture.md`
- `design/api.md`
- `design/operations.md`
- `design/schema.md`

sidebar は明示的に管理しています。`docs/site/docs/` に新しいページを追加し、そのページを navigation に表示したい場合は `sidebars.ts` も更新してください。

## ローカライズ

Docusaurus は英語を default locale、日本語を追加 locale として設定しています。

このサイト外のリポジトリドキュメントでは、通常 `docs/development.md` と `docs/development.ja.md` のように英語版と日本語版を同期したペアとして管理します。docs-site の README や元になるドキュメントで新しいトピックを追加する場合も、同じルールを適用してください。

Docusaurus page では、英語 source を `docs/site/docs/` に置き、日本語 localized page を `docs/site/i18n/ja/docusaurus-plugin-content-docs/current/` に置きます。

## 編集ガイドライン

- サイト設定は `docusaurus.config.ts` に置きます。
- navigation 構成は `sidebars.ts` に置きます。
- 見た目の調整は `src/css/custom.css` または page-local な CSS module に置きます。
- `.docusaurus/`、`build/`、`node_modules/` 配下の生成物は編集しません。
- サイトの content、configuration、styling に影響する変更は、公開前に `make dev-docs-build` を実行してください。
