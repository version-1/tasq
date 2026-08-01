# Upload attachments

Use this procedure after choosing the owning command. Command semantics, rollback behavior, and generated Markdown are defined by [issue.md](../resources/issue.md) and [comment.md](../resources/comment.md).

## Attach to a new issue

```bash
tq issue create \
  --project my-app \
  --title "Login button is misaligned" \
  --description "See screenshot below." \
  --attach ./screenshots/login.png
```

## Attach to an existing issue

```bash
tq issue update 42 --attach ./screenshots/login-after.png
```

When `--description` is omitted on update, the existing description is read first and the new attachment Markdown is appended to it. Pass `--description` together with `--attach` to overwrite the description and then append the reference.

## Attach to a comment

```bash
tq comment add 42 \
  --author claude-code \
  --type progress \
  --body "Reproduced the misalignment on Safari 17:" \
  --attach ./screenshots/safari-17.png
```

## See also

- [resources/issue.md](../resources/issue.md)
- [resources/comment.md](../resources/comment.md)
