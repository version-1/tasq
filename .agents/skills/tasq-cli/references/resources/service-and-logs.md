# Local operations

Commands for the issue-tracker, orchestrator, and Web UI. Global output behavior is in [globals.md](globals.md).

## Services

| Command | Behavior |
| --- | --- |
| `tq service start [-y]` | Checks migrations, then starts issue-tracker, orchestrator, and Web UI. It fails if a service is already running. If default ports are occupied, it proposes alternate loopback ports and asks for confirmation; `-y` accepts them. |
| `tq service stop` | Stops Web UI, orchestrator, then issue-tracker. |
| `tq service status` | Shows each service's state, PID, port, and uptime; JSON also includes address and start time. |

`start` refuses pending migrations and directs the caller to `tq migrate`. All service commands accept no positional arguments and honor `--output text|json`.

## Logs and Web UI

```text
tq logs <issue-tracker|tracker|orchestrator|web> [-n LINES] [-f]
tq web
```

`-n` defaults to `1000` and must be non-negative. Logs always stream text, ignoring `--output`; run `-f` under Monitor. `tq web` opens the running UI and fails if it is not running.

## Version, configuration, and update

| Command | Behavior |
| --- | --- |
| `tq version` | Prints the embedded version. |
| `tq config` | Shows build profile, home resolution, and resolved configuration; use global `--output json` for structured output. |
| `tq update [-y] [--tag TAG]` | Installs release artifacts, migrates local databases, and restarts services. It confirms the service interruption unless `-y` is supplied. |

`update` targets the latest formal release by default; `--tag` also accepts prereleases. After confirmation it stops services, installs the artifacts, verifies the new `tq version`, migrates, and starts services. A failure stops the remaining steps; after an install failure, fix the problem and run `tq service start`.

## See also

- [operate-services.md](../usecases/operate-services.md) for an operational sequence.
- [run-migrations.md](../usecases/run-migrations.md) for safe migration ordering.
