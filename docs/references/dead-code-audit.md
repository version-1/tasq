# Dead-Code Audit, 2026-11-15

## Scope and method

This report consolidates Issues #45 through #56 and the cross-cutting review in
#57. The review started from the 628 files returned by `git ls-files` at
`origin/main` commit `81ad6df`. It classified every tracked file by its owning
audit area or by a documented non-executable category, then traced public
entrypoints, generated outputs, embeds, configuration, and dynamic references.

### Reproducible classification

The inventory command was `git ls-files | sort`. Path prefixes assigned runtime
files to #45–#56 as follows: `internal/issue/domain`, `internal/issue/api`, and
`cmd/issue-tracker` (#45); `internal/issue/store`, issue-tracker schema and
migrations (#46); orchestrator coordinator/startup and runstore/HTTP paths
(#47/#50); runner (#48); workflow/workspace/tracker (#49); `internal/cli/tq`
and `cmd/tq` (#51/#52); issue frontend features (#53); remaining `cmd/web` and
OpenAPI (#54); development/CI/release roots (#55); and docs site/HTML (#56).
`internal/config`, `internal/buildinfo`, `internal/migration`, `cmd/tasq-tui`,
`cmd/tqstory`, and the migrations embed belong to #52.

The 628-file inventory produced 533 executable/configuration/assets files in
these path groups and 95 non-executable materials. The latter are 30 repository
or component instructions/readmes, 45 ADR/design/reference/Symphony documents,
14 plugin and skill definitions, five static documentation media files, and the
license. This category is an exclusion from code deletion, not an exclusion
from the inventory: every item was checked for its consumer (reader, tool
discovery, link, or static publication). No file was left unassigned.

| Area | Issue | Scope outcome |
| --- | --- | --- |
| Issue Tracker domain and HTTP API | #45 | Removed unused `IsActiveDependencyStatus`. |
| Issue Tracker store and schema | #46 | Removed two unreachable nil-slice fallbacks; kept migrations and schema history. |
| Orchestrator coordinator/startup | #47 | Reviewed in the area audit; no remaining confirmed removal in the integrated base. |
| Runner and transports | #48 | Retained the documented experimental WebSocket transport. |
| Workflow, workspace, tracker integration | #49 | Removed redundant workspace wrapper code. |
| Run model, persistence, HTTP API | #50 | Reviewed in the area audit; no remaining confirmed removal in the integrated base. |
| CLI resource operations, API, output | #51 | No confirmed removal; raw API and CLI contracts retained. |
| CLI operations, TUI, configuration, common code | #52 | No confirmed removal; process, OS, and update paths retained. |
| Web issue feature | #53 | Removed retired issue UI code. |
| Web common code, other screens, generated boundary, Go server | #54 | Removed unused logo variants and a generated-client alias. |
| Development, CI, release, install | #55 | No final code removal; a proposed stop-script removal was restored after review. |
| Docs site and executable HTML | #56 | Removed unused site dependency and static HTML CSS. |

The cross-cutting pass also classified the remaining tracked files: repository
and component instructions, ADRs/design/reference prose, plugin manifests and
skills, static documentation media, and license files are non-executable
materials. They were retained because they are user-facing, tool-discovered,
or linked documentation. SQL migrations are embedded historical compatibility
inputs, not removable merely because a current code path does not name an old
column. Generated API clients and public OpenAPI documents remain governed by
their sources and external consumers.

## Retained boundary cases

- Public HTTP and CLI endpoints, JSON/YAML fields, migration history, embeds,
  and platform-specific release/install paths remain compatibility contracts.
- The runner WebSocket package is intentionally transport-neutral support;
  runtime selection and a real server integration are deferred in
  `docs/symphony/DEVIATIONS.md`.
- Dynamic frontend references include Vite environment-controlled MSW loading,
  Docusaurus discovery, MDX, Storybook, static URLs, and generated clients.
- Failure, cancellation, cleanup, timeout, and process-identity branches were
  retained unless reachability was proven impossible.

## Integrated verification

- Host Go `go1.27.3 darwin/arm64`: `go test ./...`, `go vet ./...`, and
  `go build ./cmd/...` passed. The first test run correctly failed only because
  `cmd/web` embeds `frontend/dist` before the frontend build; all passed after
  that build produced `dist`.
- Dev-container Node `22.23.4`: `npm test` passed (29 files, 164 tests), then
  `npm run typecheck` and `npm run build` passed in `cmd/web/frontend`.
- Dev-container Node `22.23.4`: `npm run build` built all docs locales, and
  `npm run build -- --locale ja` separately passed in `docs/site`.
- `git diff --check` passed. No service stop, release, installer, or live
  update action was used as a verification step.

The API generator was also run. With no OpenAPI input change, the installed
generator produced whitespace-only changes in `orchestrator.ts`; those changes
were not committed, so the generated output remains at the tracked canonical
form. This is recorded as a generator formatting non-idempotence rather than a
contract change.

## Limit

Repository-wide text references are candidate evidence, not proof of runtime
reachability. Deletions were limited to code or assets with a verified unused
or unreachable condition. #54 removed logo variants with no repository use;
the retained `agent-task-queue.svg` in #56 is a public static URL whose outside
use cannot be disproved. Future documented protocol extensions remain
intentionally retained.
