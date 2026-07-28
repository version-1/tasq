# Release Binary Startup Notes

Japanese counterpart: [release-binary-startup.ja.md](release-binary-startup.ja.md).

This note records the binary-only startup requirements for the README Getting Started flow.

The full release-binary path is reproducible without Docker when the release archive contains `tq`, `issue-tracker`, `orchestrator`, and `web`, the installer has placed the service binaries under `TQ_HOME`, and the local databases have had migrations applied.

## Release Artifact Shape

GoReleaser builds these binaries into each release archive:

- `tq`
- `issue-tracker`
- `orchestrator`
- `web`

The release workflow must build `cmd/web/frontend/dist` before GoReleaser runs. The `web` binary embeds that directory with Go `embed`, so a downloaded release archive does not need Node.js or frontend files at runtime.

## Runtime State

Tasq stores machine-local runtime data under `TQ_HOME`. If `TQ_HOME` is unset, it defaults to `~/.tasq`.

The relevant files are:

- `$TQ_HOME/WORKFLOW.md`: created with the default workflow template on first use.
- `$TQ_HOME/system/state.json`: service discovery state written by running services.
- `$TQ_HOME/system/data/issues.sqlite`: issue-tracker database.
- `$TQ_HOME/system/data/orchestrator.sqlite`: orchestrator database.
- `$TQ_HOME/system/log/*.log`: logs written by `tq service start`.
- `$TQ_HOME/system/bin/{issue-tracker,orchestrator,web}`: private binaries launched by `tq service start`.

Fresh databases must be migrated before services can start:

```sh
tq migrate
```

Without this step, `issue-tracker` exits with a pending migration error and tells the user to run `tq migrate`.

## Managed Startup

The shortest binary-only full-experience path is:

```sh
export TQ_HOME="${HOME}/.tasq"
tq migrate
tq service start
tq project add --key my-project /path/to/project
tq issue create --project my-project --title "Try Tasq from binaries"
```

Then open the Web UI at `http://127.0.0.1:<web-port>` using the Web port reported by:

```sh
tq service status
```

or use:

```sh
tq web
```

`tq service start` starts all three managed services with fixed loopback ports:

| Service | Port | Data |
| --- | ---: | --- |
| `issue-tracker` | `37651` | `$TQ_HOME/system/data/issues.sqlite` |
| `orchestrator` | `37652` | `$TQ_HOME/system/data/orchestrator.sqlite` |
| `web` | `37653` | static assets embedded in the `web` binary |

`tq service start` only starts `issue-tracker`, `orchestrator`, and `web` from `$TQ_HOME/system/bin`. It does not search next to the running `tq`, in `PATH`, or in the source tree. If any managed binary is missing or not executable, it reports every invalid path before starting a service and directs the user to reinstall with the same `TQ_HOME`.

`tq service start` does not expose flags for custom ports. If one of the default ports is already in use, the binary-only README should direct users to stop the conflicting process or use manual startup.

## Manual Startup

Manual startup is useful for custom ports or for documenting each binary's flags. It is a developer workflow; direct service binary execution is not a supported distribution interface.

```sh
export TQ_HOME="$(pwd)/.tasq"
tq migrate

issue-tracker -addr 127.0.0.1:37651
orchestrator -issue-tracker http://127.0.0.1:37651 -port 37652
web -addr 127.0.0.1:37653 \
  -tracker-url http://127.0.0.1:37651 \
  -orchestrator-url http://127.0.0.1:37652
```

The flags are:

| Binary | Flags | Defaults and behavior |
| --- | --- | --- |
| `issue-tracker` | `-addr`, `-db` | `-addr` defaults to `:37651`; `-db` defaults to `$TQ_HOME/system/data/issues.sqlite`; writes `issue_tracker` to `state.json`. |
| `orchestrator` | `-db`, `-issue-tracker`, `-port` | `-db` defaults to `$TQ_HOME/system/data/orchestrator.sqlite`; `-issue-tracker` can be omitted when `state.json` has an issue-tracker address; `-port -1` disables HTTP unless workflow config enables it; writes `orchestrator` to `state.json`. |
| `web` | `-addr`, `-tracker-url`, `-orchestrator-url` | `-addr` defaults to `:37653`; backend URLs default to `http://127.0.0.1:37651` and `http://127.0.0.1:37652`; writes `web` to `state.json`. |
| `tq` | `--api-url`, `--output` | API URL resolution is `--api-url` or `-api-url`, then `TQ_API_URL`, then `$TQ_HOME/system/state.json`, then `http://localhost:37651`. |

When using non-default ports, pass `-tracker-url` and `-orchestrator-url` to `web`. Unlike `tq` and `orchestrator`, `web` does not discover backend URLs from `state.json`.

## Verification Performed

The binary-only flow was verified in `.tmp/issue-161` with an isolated `TQ_HOME` and random loopback ports:

1. Built the frontend with `npm run build`, then built `tq`, `issue-tracker`, `orchestrator`, and `web` as local binaries.
2. Confirmed that starting `issue-tracker` before migrations fails with pending migration names and a `tq migrate` instruction.
3. Ran `tq migrate`, which created both SQLite databases under `$TQ_HOME/system/data`.
4. Started `issue-tracker -addr 127.0.0.1:0`; it wrote `issue_tracker.addr` and the database path to `state.json`.
5. Ran `tq project add --key issue-161-check .tmp/issue-161/project` and `tq issue create --project issue-161-check --title "Binary startup check"` without `--api-url`; `tq` resolved the API URL from `state.json`.
6. Started `orchestrator -port 0` without `-issue-tracker`; it resolved the issue-tracker URL from `state.json` and wrote its own state.
7. Started `web -addr 127.0.0.1:0` with explicit backend URLs.
8. Confirmed the embedded SPA responded at `/` and the Web proxy returned the created issue from `/tracker/api/v1/issues`.

## README Decision Points

The README can present the full binary experience as supported, with these constraints:

- Include `tq migrate` before `tq service start`.
- Say that the installer keeps `tq` in the public install directory and installs the three managed services under `$TQ_HOME/system/bin`.
- State the default ports `37651`, `37652`, and `37653`.
- Explain that custom ports require manual service startup.
- Keep Node.js out of the runtime requirements for downloaded releases because `web` embeds the built frontend.
