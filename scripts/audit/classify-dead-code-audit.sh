#!/bin/sh
set -eu

# Emit TSV records: path, audit owner, runtime or exclusion reason. Each input
# has exactly one first-match classification; unmapped files fail the command.
# Pass a revision such as 81ad6df to reproduce a historical inventory.
revision=${1:-HEAD}
git ls-tree -r --name-only "$revision" | sort | awk '
function emit(owner, kind) { print $0 "\t" owner "\t" kind }
function excluded(reason) { print $0 "\tEXCLUDED\t" reason }
/^internal\/issue\/domain\// || /^internal\/issue\/api\// || /^cmd\/issue-tracker\// { emit("#45", "runtime"); next }
/^internal\/issue\/store\// || /^db\/schema\/issue_tracker\.sql$/ || /^db\/migrations\/issue-tracker\// { emit("#46", "runtime"); next }
/^internal\/orchestrator\/coordinator\// || /^cmd\/orchestrator\/main\.go$/ { emit("#47", "runtime"); next }
/^internal\/orchestrator\/runner\// { emit("#48", "runtime"); next }
/^internal\/orchestrator\/(workflow|workspace|tracker|workflowsync)\// || /^internal\/orchestrator\/README/ { emit("#49", "runtime"); next }
/^internal\/orchestrator\/(run|runstore|httpserver)\// { emit("#50", "runtime"); next }
/^db\/migrations\/orchestrator\// || /^db\/schema\/orchestrator\.sql$/ { emit("#50", "runtime"); next }
/^internal\/cli\/tq\/(service|update|logs|migrate|config|version|tui|tui_client|story|project_files)(_test)?\.go$/ || /^internal\/(config|buildinfo|migration)\// || /^cmd\/(tasq-tui|tqstory)\// || /^db\/migrations\/migrations\.go$/ { emit("#52", "runtime"); next }
/^internal\/cli\/tq\// || /^cmd\/tq\// { emit("#51", "runtime"); next }
/^cmd\/web\/frontend\/src\/features\/issues\// || /^cmd\/web\/frontend\/src\/app\/issues\// || /^cmd\/web\/frontend\/src\/components\/dialog\/add-issue\// { emit("#53", "runtime"); next }
/^cmd\/web\// || /^docs\/openapi\// { emit("#54", "runtime"); next }
/^(Makefile|compose\.yml|\.gitignore|\.goreleaser\.yml|\.air\..*\.toml|go\.mod|go\.sum)$/ || /^scripts\// || /^\.github\// || /^\.devcontainer\// || /^\.claude\// || /^codex\// || /^\.agents\// { emit("#55", "runtime"); next }
/^docs\/site\// || /^docs\/html\// { emit("#56", "runtime"); next }
/^LICENSE$/ || /^(README|ROADMAP|AGENTS|CLAUDE|WORKFLOW)(\.ja)?\.md$/ || /^cmd\/(issue-tracker|orchestrator|web)\/(AGENTS|CLAUDE|WORKFLOW)(\.ja)?\.md$/ || /^docs\/(adr\/|design(\/|(\.ja)?\.md$)|development(\.ja)?\.md$|references\/|symphony\/)/ || /^db\/README/ { excluded("documentation"); next }
/^plugins\// { excluded("plugin-metadata"); next }
{ print "unmapped audit file: " $0 > "/dev/stderr"; exit 1 }
' | awk -F '\t' '
NF != 3 { print "invalid classification row: " $0 > "/dev/stderr"; exit 1 }
{ counts[$2]++; print }
END { for (owner in counts) print "count\t" owner "\t" counts[owner] > "/dev/stderr" }
'
