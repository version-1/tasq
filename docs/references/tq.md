# `tq` in the Compose Development Environment

The [CLI Reference](../site/docs/reference/cli-reference.md) is the canonical
user-facing specification for `tq` commands, flags, output, and exit statuses.
Use it when running an installed `tq` binary.

## Invocation

For local Compose development, run the installed CLI inside the running `dev`
container with the Makefile wrapper. Pass only the arguments after `tq`:

```sh
make run-tq ARGS="issue list"
```

Start the environment first with `make dev-up`. The wrapper resolves the
issue-tracker API from the development service state. See the
[Makefile reference](makefile.md#main-development-targets) for the wrapper's
runtime prerequisites and related development targets.

## Related Documentation

- [CLI Reference](../site/docs/reference/cli-reference.md) — canonical command
  reference for users.
- [Tasq Operations](../design/operations.md) — local service behavior and
  verification.
- [Makefile Reference](makefile.md) — development commands and variables.
