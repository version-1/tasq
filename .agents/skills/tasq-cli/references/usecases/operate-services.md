# Operate local services

Boot, inspect, and stop the issue-tracker / orchestrator / web stack, plus tail logs and open the UI.

## Boot the stack

```bash
tq service start
```

Starts issue-tracker, orchestrator, and web in that order. Fails if:

- Any of the three is already running — run `tq service status` first if unsure.
- Any local DB has pending migrations — `migration pre-flight check failed: pending migrations: …; run `tq migrate` before starting services`. The safe boot order is therefore `tq migrate status` → `tq migrate` (if pending) → `tq service start`. See [run-migrations.md](run-migrations.md).

## Inspect

```bash
tq service status                  # text
tq service status --output json    # machine-readable
```

Text output prints `pid`, `port`, `uptime`, and `state` per service. JSON output adds `addr` and `started_at`.

## Stop

```bash
tq service stop
```

Stops web, orchestrator, then issue-tracker.

## Open the web UI

```bash
tq web
```

Errors with `web UI is not running; run `tq service start` first` if the web service is not running.

## Tail logs

```bash
tq logs issue-tracker              # last 1000 lines, then exit
tq logs tracker -n 200             # `tracker` is an alias for issue-tracker
tq logs orchestrator -n 200        # last 200 lines
tq logs web -f                     # follow new output (run under Monitor)
```

`<service>` is one of `issue-tracker` (alias `tracker`), `orchestrator`, `web`. `-n` defaults to `1000`. `tq logs` does not honor `--output json` — it always streams text. Run `-f` under the Monitor tool, not a blocking Bash call.

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

Use `tq update` when the installed binary should be replaced from GitHub Release artifacts. It stops services, installs the release artifacts, verifies the newly installed `tq version`, runs migrations, and starts services. `--tag` accepts both formal release tags and prerelease tags. `-y` skips the confirmation prompt, so use it only when service interruption is acceptable.

## See also

- [resources/service-and-logs.md](../resources/service-and-logs.md)
- [run-migrations.md](run-migrations.md) — recommended ordering when migrations and services overlap
