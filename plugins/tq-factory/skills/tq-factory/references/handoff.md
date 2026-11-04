# Handoff, verification, and QA

Use a handoff comment whenever execution changes owner, pauses, or reaches
review. The comment preserves the information a fresh session cannot infer from
the issue status alone.

```sh
tq comment add 42 --author codex --type handoff --body $'## Handoff\n\n- Outcome: API now returns an import failure reason.\n- Workspace: .worktrees/42\n- Evidence: commit abc1234; `go test ./internal/import/...` passed.\n- Review request: check retry behavior and API compatibility.\n- Remaining risk: integration test requires local object storage.\n- Next owner: independent reviewer.'
```

The reviewer reads the issue, dependencies, and all comments, then verifies the
acceptance criteria against the actual outcome. Record review evidence in a
comment. If changes are needed, assign work back deliberately; do not erase the
earlier handoff. One refinement pass may improve clarity or maintainability
without changing the accepted behavior. Any refinement requires re-verification.

QA checks that the issue's acceptance criteria, state, comments, and linked
outcomes agree. Only then may an authorized owner run `tq issue close <id>`.
For code changes, merge and publication remain governed by the target
repository's rules; a Tasq issue state never grants those permissions.
