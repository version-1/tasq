# Run migrations safely

`tq migrate` operates on the local SQLite databases the services use. Running it while a writer is open can race; the safe order is **stop → migrate → start**.

`tq service start` runs a migration pre-flight and refuses to boot if any pending migration exists (`migration pre-flight check failed: pending migrations: …; run `tq migrate` before starting services`). After applying or rolling back migrations, run `tq migrate status` to confirm before starting the stack again.

## Check first (read-only)

```bash
tq migrate status                  # text
tq migrate status --output json    # script-friendly
```

Lists applied and pending migrations per local database. Always safe to run.

## Apply pending migrations

```bash
tq service stop          # release writers
tq migrate               # apply every pending migration
tq service start         # bring the stack back up
```

## Roll back one migration

```bash
tq service stop
tq migrate down          # rolls back exactly one migration per local database
tq service start
```

`down` rolls back one step per database — run it again for each step you want to undo. Verify with `tq migrate status` between runs.

## When the stack is already down

If `tq service status` shows nothing running, skip the stop / start and just call `tq migrate` or `tq migrate down` directly.

## See also

- [resources/migrate.md](../resources/migrate.md)
- [operate-services.md](operate-services.md)
