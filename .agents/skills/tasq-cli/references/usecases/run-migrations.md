# Run migrations safely

Use the safe operational order **stop → migrate → verify → start**. Command semantics are in [migrate.md](../resources/migrate.md).

## Check first (read-only)

```bash
tq migrate status                  # text
tq --output json migrate status    # script-friendly
```

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

## When the stack is already down

If `tq service status` shows nothing running, skip the stop / start and just call `tq migrate` or `tq migrate down` directly.

## See also

- [resources/migrate.md](../resources/migrate.md)
- [operate-services.md](operate-services.md)
