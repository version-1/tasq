# Operate local services

Boot, inspect, and stop the local stack. Command-specific behavior is in [service-and-logs.md](../resources/service-and-logs.md).

## Boot the stack

```bash
tq service start
```

If startup reports pending migrations, use [run-migrations.md](run-migrations.md), then retry. Check status first if the current state is unknown.

## Inspect

```bash
tq service status                  # text
tq --output json service status    # machine-readable
```

## Stop

```bash
tq service stop
```

Stops web, orchestrator, then issue-tracker.

## Open the web UI

```bash
tq web
```

## Tail logs

```bash
tq logs issue-tracker              # last 1000 lines, then exit
tq logs tracker -n 200             # `tracker` is an alias for issue-tracker
tq logs orchestrator -n 200        # last 200 lines
tq logs web -f                     # follow new output (run under Monitor)
```

## Verify the binary version

After updating, confirm the CLI and services agree:

```bash
tq version
```

## Update the installed stack

```bash
tq update
tq update -y
tq update --tag v0.2.0-rc.1
```

## See also

- [resources/service-and-logs.md](../resources/service-and-logs.md)
- [run-migrations.md](run-migrations.md) — recommended ordering when migrations and services overlap
