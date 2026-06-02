# Orchestrator Workflow

orchestrator は run state と optional runtime inspection を所有します。自身の SQLite database に run を記録し、issue ごとの workspace を準備し、HTTP serving が有効な場合は loopback API を公開します。

`cmd/orchestrator`、`internal/orchestrator`、`db/schema/orchestrator.sql` を変更するときは、この workflow を使います。

## Scope

- Run record、run attempt、runner event、workspace metadata は orchestrator が所有します。
- orchestrator が issue state を直接変更しないようにします。
- issue-facing state の read / reconciliation には issue-tracker issue API だけを使います。
- runtime inspection は orchestrator HTTP API の local な責務として扱います。

Service boundary は [../../docs/design.md](../../docs/design.md) を参照してください。

## Local Run

Service boundary を確認するときは、issue-tracker と orchestrator を dev container 内で一緒に起動します。

```sh
make orchestrator-up
make dev-ports
```

Host-only development の場合は、先に issue-tracker を起動してから実行します。

```sh
TQ_HOME=./.tasq \
go run ./cmd/orchestrator \
  -issue-tracker http://localhost:8080 \
  -port 8081
```

## Change Flow

1. runstore、workflow、runner、workspace、HTTP server の変更は、実装上の責務を分けて扱います。
2. SQLite schema 変更は `db/schema/orchestrator.sql` に置きます。
3. issue-tracker contract の変更は意図的に行い、対応する OpenAPI document に反映します。
4. runner behavior を追加するときは、run progress を orchestrator runstore に記録します。

## Verification

開発中は focused orchestrator tests を実行します。

```sh
go test ./internal/orchestrator/...
```

Run、workspace、runtime inspection を変える behavior を渡す前に、broader checks を実行します。

```sh
go test ./...
```

Compose toolchain で確認するときは `make dev-test` を使います。

## Operational Notes

- orchestrator は自身の run record と runner event の source of truth です。
- issue status change は issue-tracker issue API 経由で行います。
- refresher が設定されていない場合、`/api/v1/refresh` は `503` を返します。
- workspace setup metadata は debugging と recovery に十分な情報を残します。
