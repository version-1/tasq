# Parse `tq` output from a script

Use `--output json` whenever an agent or a shell script needs to consume `tq` output. Pair with `jq` for extraction.

## Read a single field

```bash
# Get an issue's status.
tq issue get 42 --output json | jq -r '.status'

# Get the project key for an issue.
tq issue get 42 --output json | jq -r '.project.key'
```

## Iterate a list

```bash
# Print "id<TAB>status<TAB>title" for every issue in a project.
tq issue list --project my-app --output json \
  | jq -r '.[] | [.id, .status, .title] | @tsv'

# Just the ids of every ready issue.
tq issue list --project my-app --output json \
  | jq -r '.[] | select(.status=="ready") | .id'
```

## Capture into shell variables

```bash
ISSUE_JSON=$(tq issue get 42 --output json)
TITLE=$(jq -r '.title' <<<"$ISSUE_JSON")
STATUS=$(jq -r '.status' <<<"$ISSUE_JSON")
```

## Errors

With `--output json`, `tq` writes failures to **stderr** as `{"error":"<message>"}` and exits non-zero (`1` for runtime errors, `2` for usage errors). The success channel on stdout stays clean JSON, so a pipeline does not need to filter the error envelope out. Text output instead renders a colored `Error: <message>` for interactive use:

```bash
if ! ISSUE=$(tq issue get 42 --output json 2>/tmp/tq.err); then
  cat /tmp/tq.err >&2
  exit 1
fi
```

## Commands that ignore `--output`

- `tq logs` — always streams text.
- `tq issue watch` — always emits its own NDJSON envelopes regardless of `--output`. See [watch-and-dispatch.md](watch-and-dispatch.md).
- `tq workflow show` — has its own `--json` convenience flag. The global `--output json` also produces JSON; either works.

## See also

- [resources/globals.md](../resources/globals.md) — output formats and exit codes
