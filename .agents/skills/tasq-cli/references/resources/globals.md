# Globals: flags, env, output, exit codes

## Global flags

| Flag | Default | Notes |
| --- | --- | --- |
| `--api-url URL` | resolved (see below) | Issue-tracker API endpoint. Both `--api-url X` and `--api-url=X` are accepted. The Go-stdlib `flag`-style single dash form (`-api-url`) also works. |
| `--output text\|json` | `text` | Output format for most commands. `tq logs` and `tq issue watch` ignore this flag. |

Global flags must appear *before* the resource.

## Environment variables

| Variable | Effect |
| --- | --- |
| `TQ_API_URL` | Fallback for `--api-url` when the flag is omitted. |
| `TQ_AUTHOR` | Default `--author` value for `tq comment add`. |
| `USER` | Final fallback for the comment author when `TQ_AUTHOR` and the config `author` are empty. |

## API URL resolution order

```
--api-url flag  >  TQ_API_URL  >  $TQ_HOME/system/state.json (issue_tracker.addr)  >  http://localhost:37651
```

## Output formats

- `--output text` (default) — human-readable tables / messages.
- `--output json` — structured JSON for the same response. Use this when a script or agent needs to parse the result. See [usecases/script-with-json.md](../usecases/script-with-json.md).
- `tq logs` always streams text. `tq workflow show` has its own `--json` flag separate from the global `--output`.
- `tq issue watch` always emits NDJSON envelopes (one JSON object per line) regardless of `--output`.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success |
| `1` | Runtime error (API failure, validation rejected by server, etc.) |
| `2` | Usage error (bad flag, missing required argument) |

Text-mode errors are written to stderr as colored `Error: <message>`. With `--output json`, errors remain `{"error":"<message>"}` so scripts can parse them without ANSI escape sequences.
