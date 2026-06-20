# Upload attachments

`--attach PATH` is accepted by `tq issue create`, `tq issue update`, and `tq comment add`. The file is uploaded, then a Markdown image reference of the form `![alt](attachment://<id>)` is appended to the issue description (or comment body). If the issue / comment write fails afterwards, the uploaded attachment is deleted so no orphan file remains.

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

## Path resolution

`--attach` is resolved from the CLI's current working directory. Either pass an absolute path or `cd` into the directory that contains the file before running `tq`.

## Markdown produced

The appended snippet is:

```
![<sanitized filename>](attachment://<attachment-id>)
```

Square brackets and newlines in the filename are stripped from the alt text.

## See also

- [resources/issue.md](../resources/issue.md)
- [resources/comment.md](../resources/comment.md)
