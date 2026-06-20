# `tq service`, `tq logs`, `tq web`, `tq version`

Operational commands for the local stack: the issue-tracker API, the orchestrator, and the web UI.

## `tq service`

| Action | Usage | Notes |
| --- | --- | --- |
| `start` | `tq service start` | Starts issue-tracker, orchestrator, and web in that order. Errors if any of the three is already running. Runs a migration pre-flight first and aborts with `migration pre-flight check failed: pending migrations: …; run `tq migrate` before starting services` if there is anything pending. |
| `stop` | `tq service stop` | Stops web, orchestrator, then issue-tracker. |
| `status` | `tq service status` | Text output prints `pid`, `port`, `uptime`, and `state` (`running` / `stopped`). JSON output adds `addr` and `started_at`. |

None of the actions accept positional arguments. All honor `--output text|json`.

## `tq logs`

```
tq logs <service> [-n LINES] [-f]
```

| Flag | Default | Notes |
| --- | --- | --- |
| `<service>` (positional) | — | One of `issue-tracker` (alias `tracker`), `orchestrator`, `web`. |
| `-n` | `1000` | Tail size. Must be non-negative. |
| `-f` | off | Follow new log output. |

`tq logs` does not honor `--output json` — it always streams text. Run `-f` under the Monitor tool, not a blocking Bash call.

## `tq web`

```
tq web
```

Opens the running web UI in the default browser. Errors with `web UI is not running; run `tq service start` first` if the web service has not been started.

## `tq version`

```
tq version
```

Prints the embedded version string. Use this after an update to confirm the binary matches the running services.

## See also

- [usecases/operate-services.md](../usecases/operate-services.md) — boot / inspect / stop the stack.
- [usecases/run-migrations.md](../usecases/run-migrations.md) — recommended ordering when migrations and services overlap.
