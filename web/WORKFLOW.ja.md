# Web UI Workflow

Web UI は static export 可能な Next.js App Router client です。issue-tracker API だけを呼び、orchestrator を直接呼びません。

`web` 配下を変更するときは、この workflow を使います。

## Scope

- Runtime behavior は Client Components に置きます。
- API calls は issue-tracker API client 経由にします。
- `web/lib/generated` 配下の generated API files は再生成可能に保ちます。
- User-facing strings は `web/lib/i18n.ts` に置きます。
- Feature-specific styles は owning component または route の近くに置きます。

Web UI structure と styling conventions は [docs/design.md](docs/design.md) を参照してください。

## Local Run

issue-tracker と一緒に Web UI を確認するときは、repository-level の Compose flow を優先します。

```sh
make run-web
make dev-ports
```

Host-only development の場合:

```sh
cd web
npm install
NEXT_PUBLIC_ISSUE_TRACKER_URL=http://localhost:8080 npm run dev
```

## API Generation

issue-tracker API client は `../docs/openapi/issue-tracker.yml` から生成します。

OpenAPI document を変更した場合は、`web` から次を実行します。

```sh
npm run generate:api
```

`web/lib/generated` は手動編集しません。

## Component Flow

1. Route-owned UI は route の `_components` directory に置きます。
2. 本当に複数 route で共有する component だけを `web/components` に移動します。
3. Component directory shape は `web/docs/design.md` の定義に従います。
4. Translated display strings は `web/lib/i18n.ts` に置きます。
5. API-facing types は、generated operation を直接必要とする場合を除き、`web/lib/types.ts` 経由で import します。

## Verification

Web UI 変更を渡す前に type checking を実行します。

```sh
cd web
npm run typecheck
```

Routing、static export behavior、generated API usage、global styling に影響する変更では production build を実行します。

```sh
cd web
npm run build
```

Go と Web UI の変更を Compose でまとめて確認するときは `make dev-build` を使います。

## Static Export Notes

- Web UI runtime needs のために server actions、route handlers、server redirects、cookies、headers、Metadata API behavior を追加しません。
- Static UI を別 origin から配信する場合は、`NEXT_PUBLIC_ISSUE_TRACKER_URL` で API origin を設定します。
- Runtime で Next.js rewrites、redirects、headers に依存しません。
