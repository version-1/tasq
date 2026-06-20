# `tq migrate`

Apply, roll back, and inspect local database migrations. Operates on the local SQLite databases the services use.

## Actions

| Action | Usage |
| --- | --- |
| (default) | `tq migrate` — apply every pending migration across local databases. |
| `down` | `tq migrate down` — roll back exactly one migration per local database. |
| `status` | `tq migrate status` — list applied and pending migrations. |

## Notes

- `tq migrate` and `tq migrate down` change state on disk; stop the affected services first (`tq service stop`) when possible to avoid migrations running against an open writer.
- `tq migrate status` is read-only and safe to run any time.
- All three actions honor the global `--output text|json`.

## See also

- [usecases/run-migrations.md](../usecases/run-migrations.md) — safe ordering for stop → migrate → start.
- [service-and-logs.md](service-and-logs.md) — service lifecycle commands.
