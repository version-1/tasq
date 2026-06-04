# Issue Tracker Workflow

issue-tracker はユーザー向け API であり、issue、project、workspace、issue summary の source of truth です。

`cmd/issue-tracker`、`internal/issue`、`db/schema/issue_tracker.sql`、issue-tracker の OpenAPI contract を変更するときは、この workflow を使います。

## Scope

- issue state は issue-tracker が所有します。
- orchestrator run state は issue-tracker の persistence model に含めません。
- UI と TUI client は issue-tracker API だけを呼びます。
- contract 変更は `docs/openapi/issue-tracker.yml` に反映します。

Architecture boundary は [../../docs/design.md](../../docs/design.md) を参照してください。

## Local Run

Service interaction を確認するときは、repository-level の dev container flow を優先します。

```sh
make run-issue-tracker
make dev-ports
```

Host-only development の場合は、repository-local の `TQ_HOME` を使います。

```sh
TQ_HOME=./.tasq go run ./cmd/issue-tracker -addr :37651
```

## Change Flow

1. Public API の behavior を変える場合は、domain、store、API handler、OpenAPI contract をまとめて更新します。
2. SQLite schema 変更は `db/schema/issue_tracker.sql` に置きます。
3. `docs/openapi/issue-tracker.yml` を変更した場合は、Web UI API client を再生成します。

   ```sh
   cd cmd/web/frontend
   npm run generate:api
   ```

4. `cmd/web/frontend/src/lib/generated` 配下の generated files は手動編集しません。

## Verification

開発中は focused Go tests を実行します。

```sh
go test ./internal/issue/...
```

Contract や persistence の変更を渡す前に、repository 全体の checks を実行します。

```sh
go test ./...
cd web
npm run typecheck
```

Compose toolchain で確認するときは `make dev-test` を使います。

## Operational Notes

- issue status change は issue API 経由で行います。
- Summary response は orchestrator run snapshot ではなく issue data だけを反映します。
- work item claim や orchestrator event receiver など削除済み endpoint family は、明示的な contract 変更なしに再導入しません。
