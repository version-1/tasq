---
id: testing
title: テスト
sidebar_position: 3
---

# テスト

まず最も狭く有用な verification を実行し、変更が shared behavior、contract、persistence、user-facing flow に影響する場合は check を広げます。

## Go test

```sh
go test ./...
```

iteration 中は targeted package を使い、shared behavior を変更した場合は handoff 前に full suite を実行します。

## Web UI check

```sh
cd cmd/web/frontend
npm run typecheck
npm run build
```

OpenAPI contract を変更した場合は `npm run generate:api` で API client を再生成します。

## Docs site check

```sh
cd docs/site
npm run build
```

repository-level workflow では、`make dev-docs-build` が docs-site build を wrap します。

## Manual verification

1. `make dev-up` で dev environment を起動するか、`tq service start` で host service を起動します。
2. `tq` または Web UI で issue を作成・更新します。
3. issue summary に status change が反映されることを確認します。
4. orchestrator が有効な場合、orchestrator runtime inspection に到達できることを確認します。
