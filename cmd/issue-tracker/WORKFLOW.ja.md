# Issue Tracker Workflow

issue-tracker はユーザー向け API であり、issue、work item、受信済み orchestrator event id、最新 run snapshot の source of truth です。

`cmd/issue-tracker`、`internal/issue`、`db/schema/issue_tracker.sql`、issue-tracker の OpenAPI contract を変更するときは、この workflow を使います。

## Scope

- issue state と work item claim state は issue-tracker が所有します。
- orchestrator の run state は受信した fact として扱い、issue-tracker が直接管理する状態にしません。
- UI と TUI client は issue-tracker API だけを呼びます。
- contract 変更は `docs/openapi/issue-tracker.yml` に反映します。

Architecture boundary は [../../docs/design.md](../../docs/design.md) を参照してください。

## Local Run

Service interaction を確認するときは、repository-level の Compose flow を優先します。

```sh
make issue-tracker-up
make dev-ports
```

Host-only development の場合:

```sh
go run ./cmd/issue-tracker -addr :8080 -db tasq-issues.sqlite
```

## Change Flow

1. Public API の behavior を変える場合は、domain、store、API handler、OpenAPI contract をまとめて更新します。
2. SQLite schema 変更は `db/schema/issue_tracker.sql` に置きます。
3. `docs/openapi/issue-tracker.yml` を変更した場合は、Web UI API client を再生成します。

   ```sh
   cd web
   npm run generate:api
   ```

4. `web/lib/generated` 配下の generated files は手動編集しません。

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

- Claim token は work item claim の generation marker です。
- Duplicate orchestrator event id は idempotent に受け付ける必要があります。
- Expired claim からの late event は記録してもよいですが、current issue state を更新してはいけません。
