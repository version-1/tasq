# Run migrations safely

For forward migrations, use **stop → migrate → verify → start**. A rollback also requires switching to application binaries compatible with the rolled-back schema before restarting. Command semantics are in [migrate.md](../resources/migrate.md).

## Check first (read-only)

```bash
tq migrate status                  # text
tq --output json migrate status    # script-friendly
```

## Apply pending migrations

```bash
tq service stop          # release writers
tq migrate               # apply every pending migration
tq migrate status        # verify no migration remains pending
tq service start         # bring the stack back up
```

## Roll back one migration

```bash
tq service stop
tq migrate down          # rolls back exactly one migration per local database
tq migrate status        # verify the intended migration is now pending
```

Do not immediately restart the same application version: its startup pre-flight sees the rolled-back migration as pending and refuses to start. Switch to or install the application version that expects the rolled-back schema, then run `tq service start` with that compatible version.

## When the stack is already down

If `tq service status` shows nothing running, skip `tq service stop`. After a forward migration, start the services normally. After a rollback, first switch to a compatible application version as described above.

## See also

- [resources/migrate.md](../resources/migrate.md)
- [operate-services.md](operate-services.md)
