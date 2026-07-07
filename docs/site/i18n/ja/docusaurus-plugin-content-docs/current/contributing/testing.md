---
id: testing
title: Testing
sidebar_position: 3
---

# Testing

まず narrowest useful verification を実行し、shared behavior、contracts、persistence、user-facing flows に影響する変更では checks を広げます。

## Recommended Matrix

| Change area | Start with | Broaden to |
| --- | --- | --- |
| Go package or service logic | `go test ./internal/<package>` | `go test ./...` |
| Issue-tracker or orchestrator API contract | targeted Go tests | API generation、Web typecheck、`make dev-test` |
| Web UI component or route | `cmd/web/frontend` で `npm run typecheck` | `npm run build`、`make dev-build` |
| Docs site | `docs/site` で `npm run build` | English と Japanese pages を browser で確認 |
| End-to-end service flow | local smoke flow | `make dev-build` |

## Go Tests

```sh
go test ./...
```

iteration 中は targeted packages を使い、shared behavior を変更した場合は handoff 前に full suite を実行します。

Compose では次のように実行します。

```sh
make dc-exec CMD="go test ./internal/config"
make dc-exec CMD="go test ./..."
```

## Web UI Checks

```sh
cd cmd/web/frontend
npm run typecheck
npm run build
```

OpenAPI contracts を変更した場合は `npm run generate:api` で API clients を regenerate
します。OpenAPI change と一緒に `cmd/web/frontend/src/lib/generated` 配下の generated
files を commit し、standalone mock development で変更した endpoint を使う場合は MSW
handlers や fixtures も更新します。

in-memory mock data で frontend behavior だけを確認する場合は `npm run dev:mock` を
使います。real issue-tracker や orchestrator behavior が必要な場合は Compose を
使ってください。

## Docs Site Checks

```sh
cd docs/site
npm run build
```

repository-level workflow では、`make dev-docs-build` が docs-site build を wrap します。

docs changes では、両方が存在する場合に English と Japanese pages を一緒に更新します。
new pages を追加する場合は links、sidebar placement、code block languages も確認します。

## Compose Verification

複数の runtime areas にまたがる変更では、handoff 前に repository-level targets を使います。

```sh
make dev-test
make dev-build
```

`make dev-test` は dev container 内で Go tests と Web UI typecheck を実行します。
`make dev-build` は Go tests と Web production build を実行します。

## Manual Verification

1. `make dev-up` で dev environment を起動するか、`tq service start` で host services を起動します。
2. `tq` または Web UI で issues を作成・更新します。
3. issue summaries が status changes を反映することを確認します。
4. orchestrator が enabled の場合は、orchestrator runtime inspection に到達できることを確認します。

verification step が applicable でない、または local で実行できない場合は、pull request
に skipped check と理由を記録してください。
