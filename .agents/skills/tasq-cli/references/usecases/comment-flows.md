# Comment flows: progress, blocker, handoff

Comment `--type` exists so the orchestrator UX (and human readers) can distinguish routine notes from blockers and handoffs. Always pass `--author` from an agent so the source is attributable.

## Progress update

```bash
tq comment add 42 \
  --author claude-code \
  --type progress \
  --body "Migrated the login endpoint. Logout still uses the old cookie."
```

## Recording a blocker

A blocker comment usually pairs with `tq issue update --status blocked`:

```bash
tq comment add 42 \
  --author claude-code \
  --type blocker \
  --body "Cannot proceed: the OIDC client secret is not in the local env. Need ENV.OIDC_CLIENT_SECRET."
tq issue update 42 --status blocked
```

## Handoff to the next agent / reviewer

```bash
tq comment add 42 \
  --author claude-code \
  --type handoff \
  --body $'Working branch: feat/oidc-migration\nPR: https://github.com/org/repo/pull/123\nRemaining: write integration test for logout.'
```

Bash's `$'…'` form lets you embed newlines without `printf`.

## Listing comments

```bash
tq comment list 42                  # text
tq comment list 42 --output json    # structured
```

## Author resolution

Without `--author`, `tq` picks the first non-empty value from:

1. `TQ_AUTHOR`
2. config `author` field
3. `$USER`

An agent should not rely on `$USER` — pass `--author` explicitly.

## See also

- [resources/comment.md](../resources/comment.md)
- [resources/enums.md](../resources/enums.md)
- [attachments.md](attachments.md) — attach images to comments
