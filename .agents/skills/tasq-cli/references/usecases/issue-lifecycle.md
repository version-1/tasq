# Issue lifecycle: backlog → done

Use this procedure for the common flow. Command flags and enum definitions remain in [issue.md](../resources/issue.md) and [enums.md](../resources/enums.md).

## Create

```bash
# Minimum: title + project key. Status defaults to backlog, priority to normal.
tq issue create --project my-app --title "Migrate auth to OIDC"

# Fully specified.
tq issue create \
  --project my-app \
  --title "Migrate auth to OIDC" \
  --description "Replace the legacy session cookie with OIDC." \
  --status ready \
  --priority high \
  --assignee jiro
```

For a multi-line description, use ANSI-C quoting on one physical command line:

```bash
tq issue create --project my-app --title "Migrate auth to OIDC" --description $'Context: replace the legacy session cookie.\nAcceptance: OIDC login works and the old cookie path is removed.'
```

## Inspect

```bash
tq issue get 42                        # single issue
tq issue list                          # everything
tq issue list --project my-app         # one project
tq --output json issue list | jq …     # script-friendly
```

## Move between states

```bash
# Shortcuts.
tq issue ready 42      # → ready
tq issue draft 42      # → backlog
tq issue close 42      # → done
tq issue cancel 42     # → cancelled

# Long form for every other status.
tq issue update 42 --status in_progress
tq issue update 42 --status review
tq issue update 42 --status blocked
tq issue update 42 --status failed
tq issue update 42 --status duplicate
```

A typical orchestrator-driven flow is `ready → in_progress → review → done`, with `blocked` as a side path. Record the reason for `blocked` as a `blocker` comment ([comment-flows.md](comment-flows.md)).

## Edit metadata

```bash
# Update one or many fields in a single call. At least one is required.
tq issue update 42 --priority urgent
tq issue update 42 --assignee jiro --priority high

# Replace the title / description outright.
tq issue rename 42 "Migrate auth to OIDC and remove legacy cookie"
tq issue edit   42 "Updated description body, replaces the old one."
```

## See also

- [resources/issue.md](../resources/issue.md)
- [resources/enums.md](../resources/enums.md)
- [comment-flows.md](comment-flows.md)
