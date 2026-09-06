# Tasq Docs Site

このディレクトリには、Tasq の Docusaurus ドキュメントサイトがあります。このサイトは、リポジトリ内の `docs/` にある開発者向け・利用者向けドキュメントを公開用に整理して提供します。

英語ドキュメントを一次情報として扱います。トピックを追加または変更する場合は、日本語ドキュメントも同期してください。

## コマンド

リポジトリルートから実行するコマンド:

```sh
make dev-docs
make dev-docs-ja
make dev-docs-build
make dev-docs-open
```

`docs/site/` の中で Docusaurus プロジェクトを直接扱う場合のコマンド:

```sh
npm install
npm start
npm run start:ja
npm run build
npm run serve
```

英語の開発サーバーは Docusaurus の既定ポートと、設定済みのベース URL を使います。

```text
http://localhost:3000/tasq/
```

ローカライズされたコンテンツをプレビューする場合は、日本語の開発サーバーを別に起動します。

```sh
npm run start:ja
```

```text
http://localhost:3000/tasq/ja/
```

Docusaurus の `start` は一度に 1 つの locale だけを配信します。すべての locale を含む本番ビルドをプレビューしたい場合は、`npm run build` のあとに `npm run serve` を使ってください。

## ディレクトリ構成

```text
docs/site/
  docs/                    Docusaurus の docs コンテンツ。
    getting-started/       概要、インストール、チュートリアル、セットアップ、コンセプト。
    guides/                利用者向けのタスクと運用ガイド。
    reference/             CLI、API、設定、スキーマのリファレンス。
  static/img/              静的な図表やその他のドキュメント画像。
  i18n/ja/                 Docusaurus の日本語ローカライズコンテンツ。
  src/
    css/custom.css         サイト全体のテーマ上書き。
  docusaurus.config.ts     サイト設定、ベース URL、locale、navbar、footer。
  sidebars.ts              docs sidebar の構成。
  package.json             npm scripts と Docusaurus 依存。
  package-lock.json        npm 依存グラフの lockfile。
  tsconfig.json            サイト用 TypeScript 設定。
```

次のディレクトリは生成物で、ソースファイルではありません。

- `node_modules/` は `npm install` で作成されます。
- `.docusaurus/` は Docusaurus の開発時と build 時に作成されます。
- `build/` は `npm run build` で作成されます。

## コンテンツ構成

公開される docs のソースは `docs/site/docs/` に置きます。docs-site は利用者向けドキュメントの正とします。リポジトリの `docs/design/` 配下にある内部設計ドキュメントは、開発者向けの設計記録として別管理し、より低レベルの実装詳細を含む場合があります。

現在のページ一覧とセクション構成は `sidebars.ts` を参照してください。sidebar は明示的に管理しています。`docs/site/docs/` に新しいページを追加し、そのページを navigation に表示したい場合は `sidebars.ts` も更新してください。

通常の図表には Mermaid を使います。より情報量の多い静的 overview が必要なページでは、手書き SVG を `static/img/` 配下に置きます。

## ローカライズ

Docusaurus は英語を default locale、日本語を追加 locale として設定しています。

このサイト外のリポジトリドキュメントでは、通常 `docs/development.md` と `docs/development.ja.md` のように英語版と日本語版を同期したペアとして管理します。docs-site の README や元になるドキュメントで新しいトピックを追加する場合も、同じルールを適用してください。

Docusaurus page では、英語 source を `docs/site/docs/` に置き、日本語 localized page を `docs/site/i18n/ja/docusaurus-plugin-content-docs/current/` に置きます。

## 編集ガイドライン

- サイト設定は `docusaurus.config.ts` に置きます。
- navigation 構成は `sidebars.ts` に置きます。
- 見た目の調整は `src/css/custom.css` または page-local な CSS module に置きます。
- `.docusaurus/`、`build/`、`node_modules/` 配下の生成物は編集しません。
- サイトのコンテンツ、設定、スタイルに影響する変更は、公開前に `make dev-docs-build` を実行してください。
