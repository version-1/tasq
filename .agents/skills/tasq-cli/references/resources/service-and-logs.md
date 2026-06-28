# `tq service`, `tq logs`, `tq web`, `tq version`, `tq update`

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

## `tq update`

```
tq update [-y] [--tag TAG]
```

Installs `tq`, `issue-tracker`, `orchestrator`, and `web` from GitHub Release artifacts, applies migrations, and restarts local services. By default it targets the latest formal release. `--tag` installs a specific release or prerelease tag.

The command prints the current version and target version before stopping services. It asks for confirmation because services will stop and restart; pass `-y` only when that disruption is expected.

Flow:

1. Resolve target release.
2. Print current and target versions.
3. Confirm unless `-y` is set.
4. Stop services.
5. Install release artifacts into the fixed user install location.
6. Run the newly installed `tq version`.
7. Apply migrations.
8. Start services.

If any step fails, later steps are not run. If install fails after services stop, inspect the error, then run `tq service start` after fixing the install problem.

## See also

- [usecases/operate-services.md](../usecases/operate-services.md) — boot / inspect / stop the stack.
- [usecases/run-migrations.md](../usecases/run-migrations.md) — recommended ordering when migrations and services overlap.
