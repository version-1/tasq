---
id: testing
title: Testing
sidebar_position: 3
---

# Testing

まず narrowest useful verification を実行し、shared behavior、contracts、persistence、user-facing flows に影響する変更では checks を広げます。

## Go Tests

```sh
go test ./...
```

iteration 中は targeted packages を使い、shared behavior を変更した場合は handoff 前に full suite を実行します。

## Web UI Checks

```sh
cd cmd/web/frontend
npm run typecheck
npm run build
```

OpenAPI contracts を変更した場合は `npm run generate:api` で API clients を regenerate します。

## Docs Site Checks

```sh
cd docs/site
npm run build
```

repository-level workflow では、`make dev-docs-build` が docs-site build を wrap します。

## Manual Verification

1. `make dev-up` で dev environment を起動するか、`tq service start` で host services を起動します。
2. `tq` または Web UI で issues を作成・更新します。
3. issue summaries が status changes を反映することを確認します。
4. orchestrator が enabled の場合は、orchestrator runtime inspection に到達できることを確認します。
