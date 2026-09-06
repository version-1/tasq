# Web UI Workflow

Web UI は Vite + React + TypeScript の single-page app を embed する Go HTTP server です。Server は frontend assets の配信、SPA fallback、local backend services への API reverse proxy を担当します。

`cmd/web` 配下を変更するときはこの workflow を使います。

## Scope

- Go server は static asset serving、SPA fallback、API reverse proxying を担当します。
- Browser UI code は `cmd/web/frontend/src` 配下に置きます。
- Generated issue-tracker と orchestrator API files は `cmd/web/frontend/src/lib/generated` 配下で再生成可能に保ちます。
- User-facing strings は `cmd/web/frontend/src/lib/i18n.ts` に置きます。
- Feature-specific styles は owning component または route の近くに置きます。

Web UI structure と styling conventions は [docs/design/web.md](../../docs/design/web.md) を参照してください。

## Local Run

Issue-tracker と合わせて Web UI を確認するときは repository-level Compose flow を優先します。

```sh
make run-web
make dev-ports
```

Host-only development:

```sh
cd cmd/web/frontend
npm install
npm run build
cd ../../..
go run ./cmd/web
```

Issue-tracker や orchestrator を起動しない frontend standalone development:

```sh
cd cmd/web/frontend
npm run dev:mock
```

`dev:mock` は `VITE_MSW=true` を設定し、
`cmd/web/frontend/public/mockServiceWorker.js` の browser MSW worker を起動します。
Mock data は in-memory で、page reload 時に reset されます。

## API Generation

OpenAPI 再生成コマンドと generated file の所有ルールは [../../docs/development.md](../../docs/development.md#api-generation) を参照してください。

変更した endpoint が standalone frontend development で使われる場合は、`cmd/web/frontend/src/mocks` 配下の MSW handlers と fixtures も更新します。

## Component Flow

Frontend の routing、feature/component 配置、directory 所有ルールは [frontend/docs/design.md](frontend/docs/design.md) を参照してください。

- Translated display strings は `cmd/web/frontend/src/lib/i18n.ts` に置きます。
- Generated operation を直接使う必要がない限り、API-facing types は `cmd/web/frontend/src/lib/types.ts` 経由で import します。

## Verification

Web UI changes の handoff 前に type checking を実行します。

```sh
cd cmd/web/frontend
npm run typecheck
```

Routing、generated API usage、global styling、embedded assets に影響する変更では production build を実行します。

```sh
cd cmd/web/frontend
npm run build
cd ../../..
go build -o .tmp/tasq-web ./cmd/web
```

Compose 経由の `make dev-build` check は [../../docs/development.md](../../docs/development.md#verification) を参照してください。
