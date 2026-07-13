#!/usr/bin/env bash

set -euo pipefail

rules_file=""

usage() {
  echo "Usage: $0 --rules <rules-file>" >&2
}

while (($# > 0)); do
  case "$1" in
    --rules)
      rules_file="${2:-}"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

if [[ -z "$rules_file" || ! -f "$rules_file" ]]; then
  echo "Rules file was not found: ${rules_file:-<missing>}" >&2
  exit 2
fi

for dependency in codex jq; do
  if ! command -v "$dependency" >/dev/null 2>&1; then
    echo "Missing dependency: $dependency" >&2
    exit 2
  fi
done

failures=0

check() {
  local step="$1"
  shift

  local result decision command
  command="$(printf '%q ' "$@")"
  result="$(codex execpolicy check --rules "$rules_file" -- "$@")"
  decision="$(jq -r '.decision // "unknown"' <<<"$result")"

  if [[ "$decision" == "allow" ]]; then
    printf 'PASS\t%s\t%s\t%s\n' "$step" "$decision" "$command"
    return
  fi

  printf 'GAP\t%s\t%s\t%s\n' "$step" "$decision" "$command"
  failures=1
}

check "inspect-tracker" tq project list
check "inspect-issue" tq issue get 1
check "start-work" tq issue update 1 --status in_progress
check "record-handoff" tq comment add 1 --author codex --body started
check "finish-work" tq issue update 1 --status review
check "inspect-repository" git status --short
check "update-base" git fetch origin
check "create-worktree" git worktree add .worktrees/tasq/1-example -b agent/1-example
check "stage-change" git add README.md
check "commit-change" git commit -m "Implement issue 1"
check "publish-branch" git push origin HEAD
check "create-pr" gh pr create --title "Implement issue 1" --body ""
check "inspect-pr" gh pr checks 1
check "close-worktree" git worktree remove .worktrees/tasq/1-example
check "remove-merged-branch" git branch -d agent/1-example
check "update-existing-pr" gh pr edit 1 --title "Implement issue 1"

exit "$failures"
