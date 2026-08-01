# Comment flows: progress, blocker, handoff

Use explicit author and type values so orchestration and human readers can classify the note. See [comment.md](../resources/comment.md) for defaults and attachment behavior.

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
tq --output json comment list 42    # structured
```

## See also

- [resources/comment.md](../resources/comment.md)
- [resources/enums.md](../resources/enums.md)
- [attachments.md](attachments.md) — attach images to comments
