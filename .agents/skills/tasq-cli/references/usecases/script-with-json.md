# Parse `tq` output from a script

Use global `--output json` before the resource whenever a script consumes supported `tq` output. Pair it with `jq` for extraction; exceptions are in [globals.md](../resources/globals.md).

## Read a single field

```bash
# Get an issue's status.
tq --output json issue get 42 | jq -r '.status'

# Get the project key for an issue.
tq --output json issue get 42 | jq -r '.project.key'
```

## Iterate a list

```bash
# Print "id<TAB>status<TAB>title" for every issue in a project.
tq --output json issue list --project my-app \
  | jq -r '.[] | [.id, .status, .title] | @tsv'

# Just the ids of every ready issue.
tq --output json issue list --project my-app \
  | jq -r '.[] | select(.status=="ready") | .id'
```

## Capture into shell variables

```bash
ISSUE_JSON=$(tq --output json issue get 42)
TITLE=$(jq -r '.title' <<<"$ISSUE_JSON")
STATUS=$(jq -r '.status' <<<"$ISSUE_JSON")
```

## Errors

With `--output json`, `tq` writes failures to **stderr** as `{"error":"<message>"}` and exits non-zero (`1` for runtime errors, `2` for usage errors). The success channel on stdout stays clean JSON, so a pipeline does not need to filter the error envelope out. Text output instead renders a colored `Error: <message>` for interactive use:

```bash
if ! ISSUE=$(tq --output json issue get 42 2>/tmp/tq.err); then
  cat /tmp/tq.err >&2
  exit 1
fi
```

## See also

- [resources/globals.md](../resources/globals.md) — output formats and exit codes
