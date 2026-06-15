# Database Management

Tasq は SQLite schema change を versioned migration files で管理します。

Service startup は schema object を自動作成・自動変更しません。Issue-tracker と orchestrator の store は pending migration の有無だけを確認します。Pending migration がある場合、service startup は `tq migrate` の実行を促して失敗します。

## Layout

- `db/migrations/issue-tracker/`: issue-tracker database 用の ordered migrations。
- `db/migrations/orchestrator/`: orchestrator database 用の ordered migrations。
- `db/schema/issue_tracker.sql`: 全 migration 適用後の issue-tracker schema reference snapshot。
- `db/schema/orchestrator.sql`: 全 migration 適用後の orchestrator schema reference snapshot。

`db/schema/*.sql` files は documentation と review aid です。Runtime code は startup schema change のためにこれらを embed または実行してはいけません。

## Migration Files

各 migration は SQL file の pair です。

```text
<version>_<name>.up.sql
<version>_<name>.down.sql
```

Version は timestamp-like string で、lexical order で適用されます。すべての `.up.sql` file には対応する `.down.sql` file が必要です。

## Commands

Local database の pending migration をすべて適用します。

```sh
tq migrate
```

Local database ごとに migration を 1 つ rollback します。

```sh
tq migrate down
```

Applied / pending migration を確認します。

```sh
tq migrate status
```

Dev container workflow では次を使います。

```sh
make run-migrate
```

`make dev-up` は service startup 前に `run-all` 経由で migration を実行します。

## Changing Schema

Database schema を変更するとき:

1. 対象 database directory に新しい migration pair を追加する。
2. 対応する `db/schema/*.sql` reference snapshot を更新する。
3. Migration behavior と影響を受ける store behavior の test を追加または更新する。
4. `go test ./...` を実行する。

Store startup に inline schema migration code を追加してはいけません。Schema changes は migration files で表現し、適用状態を `schema_migrations` で確認できるようにします。

