# Database Management

Tasq manages SQLite schema changes with versioned migration files.

Service startup does not create or modify schema objects automatically. The issue-tracker and
orchestrator stores only check whether migrations are pending. If pending migrations exist, service
startup fails with guidance to run `tq migrate`.

## Layout

- `db/migrations/issue-tracker/`: ordered migrations for the issue-tracker database.
- `db/migrations/orchestrator/`: ordered migrations for the orchestrator database.
- `db/schema/issue_tracker.sql`: reference snapshot of the issue-tracker schema after all migrations.
- `db/schema/orchestrator.sql`: reference snapshot of the orchestrator schema after all migrations.

`db/schema/*.sql` files are documentation and review aids. Runtime code must not embed or execute
them for startup schema changes.

## Migration Files

Each migration is a pair of SQL files:

```text
<version>_<name>.up.sql
<version>_<name>.down.sql
```

Versions are timestamp-like strings and are applied in lexical order. Every `.up.sql` file must have
a matching `.down.sql` file.

## Commands

Apply all pending migrations for local databases:

```sh
tq migrate
```

Roll back one migration per local database:

```sh
tq migrate down
```

Inspect applied and pending migrations:

```sh
tq migrate status
```

In the dev container workflow, use:

```sh
make run-migrate
```

`make dev-up` runs migrations through `run-all` before starting services.

## Changing Schema

When changing a database schema:

1. Add a new migration pair under the target database directory.
2. Update the matching `db/schema/*.sql` reference snapshot.
3. Add or update tests for migration behavior and affected store behavior.
4. Run `go test ./...`.

Do not add inline schema migration code to store startup. Schema changes must be represented by
migration files so applied state is visible in `schema_migrations`.

