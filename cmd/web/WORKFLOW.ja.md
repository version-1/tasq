# Web UI Workflow

Web UI は Vite + React + TypeScript の single-page app を embed する Go HTTP server です。Server は frontend assets の配信、SPA fallback、local backend services への API reverse proxy を担当します。

`cmd/web` 配下を変更するときはこの workflow を使います。

## Scope

- Go server は static asset serving、SPA fallback、API reverse proxying を担当します。
- Browser UI code は `cmd/web/frontend/src` 配下に置きます。
- Generated issue-tracker API files は `cmd/web/frontend/src/lib/generated` 配下で再生成可能に保ちます。
- User-facing strings は `cmd/web/frontend/src/lib/i18n.ts` に置きます。
- Feature-specific styles は owning component または route の近くに置きます。

Web UI structure と styling conventions は [docs/design.md](docs/design.md) を参照してください。

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

## API Generation

Issue-tracker API client は `docs/openapi/issue-tracker.yml` から生成します。

OpenAPI document を変更したら、`cmd/web/frontend` から次を実行します。

```sh
npm run generate:api
```

`cmd/web/frontend/src/lib/generated` は手動編集しません。

## Component Flow

1. Route-owned UI は route の `_components` directory に置きます。
2. 複数 route で本当に共有する component だけ `cmd/web/frontend/src/components` へ移します。
3. Component directory shape は `cmd/web/docs/design.md` に従います。
4. Translated display strings は `cmd/web/frontend/src/lib/i18n.ts` に置きます。
5. Generated operation を直接使う必要がない限り、API-facing types は `cmd/web/frontend/src/lib/types.ts` 経由で import します。

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

Go と Web UI を Compose 経由で一緒に検証するときは `make dev-build` を使います。
